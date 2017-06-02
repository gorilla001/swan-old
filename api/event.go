package api

import (
	"net/http"
)

func (r *Router) getEvents(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(nil)

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if catchUp := req.Form.Get("catchUp"); catchUp == "true" {
		return
	}

	if err := r.driver.SubscribeEvent(w, req.RemoteAddr); err != nil {
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
	}

	return
}
