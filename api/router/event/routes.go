package event

import (
	"github.com/Dataman-Cloud/swan/api/router"
)

type Router struct {
	routes  []*router.Route
	backend Backend
}

// NewRouter initializes a new application router.
func NewRouter(b Backend) *Router {
	r := &Router{
		backend: b,
	}

	r.initRoutes()
	return r
}

func (r *Router) Routes() []*router.Route {
	return r.routes
}

func (r *Router) initRoutes() {
	r.routes = []*router.Route{
		router.NewRoute("GET", "/v1/event", r.EventStream),
	}
}
