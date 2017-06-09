package api

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	//. "github.com/Dataman-Cloud/swan/api"
	//"github.com/Dataman-Cloud/swan/api/middleware"
)

type Server struct {
	listen      string
	leader      string
	router      *Router
	server      *http.Server
	middlewares []Middleware
}

func NewServer(addr string) *Server {
	srv := &Server{
		listen: addr,
	}

	//srv.initMiddlewares()

	return srv
}

// createMux initializes the main router the server uses.
func (s *Server) createMux() *mux.Router {
	m := mux.NewRouter()

	log.Debug("Registering HTTP route")
	for _, r := range s.router.Routes() {
		f := s.makeHTTPHandler(r.Handler())

		log.Debugf("Registering %s, %s", r.Method(), r.Path())

		m.Path(r.Path()).Methods(r.Method()).Handler(f)
	}

	return m
}

func (s *Server) enableCORS(w http.ResponseWriter) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, X-Registry-Auth")
	w.Header().Add("Access-Control-Allow-Methods", "HEAD, GET, POST, DELETE, PUT, OPTIONS")
}

func (s *Server) makeHTTPHandler(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var herr httpError

		defer func() {
			log.Printf("%s - %s \"%s %s %s\" %d",
				strings.Split(r.RemoteAddr, ":")[0],
				time.Now().Format("02/Jan/2006:15:04:05 -0700"),
				r.Method,
				r.URL.RequestURI(),
				r.Proto,
				herr.StatusCode(),
			)

		}()

		s.enableCORS(w)

		if s.listen == s.leader {
			err := handler(w, r)

			if e, ok := err.(httpError); ok {
				herr = e
			}

			if err != nil {
				http.Error(w, herr.Error(), herr.StatusCode())
			}

			return
		} else {
			if r.Method == "GET" {
				err := handler(w, r)

				if e, ok := err.(httpError); ok {
					herr = e
				}

				if err != nil {
					http.Error(w, herr.Error(), herr.StatusCode())
				}

			}
			return
		}

		s.forwardRequest(w, r)
	}
}

func (s *Server) wrapHandlerWithMiddlewares(handler HandlerFunc) HandlerFunc {
	for _, m := range s.middlewares {
		handler = m.WrapHandler(handler)
	}

	return handler
}

//func (s *Server) initRouter(driver Driver, db Store) {
//	s.router = api.NewRouter(driver, db)
//}

func (s *Server) InstallRouter(r *Router) {
	s.router = r
}

//func (s *Server) initMiddlewares() {
//	s.middlewares = []Middleware{
//		middleware.NewCORSMiddleware(),
//		//middleware.NewNCSACommonLogMiddleware(),
//	}
//}
//
//func (s *Server) InstallMiddleware(mid Middleware) {
//	s.middlewares = append(s.middlewares, mid)
//}
//
//func (s *Server) UninstallMiddleware(mid Middleware) {
//	for idx, midle := range s.middlewares {
//		if midle.Name() == mid.Name() {
//			s.middlewares = append(s.middlewares[:idx], s.middlewares[idx+1:]...)
//		}
//	}
//}

func (s *Server) Run() error {
	srv := &http.Server{
		Addr:    s.listen,
		Handler: s.createMux(),
	}

	s.server = srv

	log.Infof("API Server listening on %s", s.listen)

	return srv.ListenAndServe()
}

// gracefully shutdown.
func (s *Server) Shutdown() error {
	// If s.server is nil, api server is not running.
	if s.server != nil {
		// NOTE(nmg): need golang 1.8+ to run this method.
		return s.server.Shutdown(nil)
	}

	return nil
}

func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}

	return nil
}

func (s *Server) Reload() error {
	log.Println("Reload api server for leader change")

	if err := s.Stop(); err != nil {
		return fmt.Errorf("Shutdown api server error: %v", err.Error())
	}
	// NOTE(nmg): Sometimes the api server can't be closed immediately.
	// In this situation the `bind: address already in use` error will be occured.
	// So we use a `for loop` to aviod this.
	// TODO(nmg): Fix this more elegant.
	for {
		err := s.Run()
		if strings.Contains(err.Error(), "bind: address already in use") {
			log.Errorf("Start apiserver error %s. Retry after 1 second.", err.Error())
			time.Sleep(1 * time.Second)
			continue
		}

		return fmt.Errorf("apiserver run error: %v", err)
	}

}

func (s *Server) Update(leader string) {
	s.leader = leader
}

func (s *Server) forwardRequest(w http.ResponseWriter, r *http.Request) {
	// NOTE(nmg): If you just use ip address here, the `url.Parse` with get error with
	// `first path segment in URL cannot contain colon`.
	// It's golang 1.8's bug. more details see https://github.com/golang/go/issues/18824.
	leaderUrl := s.leader
	if !strings.HasPrefix(leaderUrl, "http://") {
		leaderUrl = "http://" + s.leader
	}

	leaderURL, err := url.Parse(leaderUrl + r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rr, err := http.NewRequest(r.Method, leaderURL.String(), r.Body)
	rr.URL.RawQuery = r.URL.RawQuery
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	copyHeader(r.Header, &rr.Header)

	// Create a client and query the target
	client := &http.Client{}
	lresp, err := client.Do(rr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Request forwarding %s %s %s", rr.Method, rr.URL, lresp.Status)

	dH := w.Header()
	copyHeader(lresp.Header, &dH)
	dH.Add("Requested-Host", rr.Host)

	reader := bufio.NewReader(lresp.Body)
	for {
		line, err := reader.ReadBytes('\n')

		if err == io.EOF {
			if _, err := w.Write(line); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(line) == 0 {
			continue
		}

		if _, err := w.Write(line); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	return
}

func copyHeader(src http.Header, dest *http.Header) {
	for n, v := range src {
		for _, vv := range v {
			dest.Set(n, vv)
		}
	}
}
