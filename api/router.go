package api

import (
	. "github.com/Dataman-Cloud/swan/store"
)

type Router struct {
	routes []*Route
	driver Driver
	db     Store
}

func NewRouter(d Driver, s Store) *Router {
	r := &Router{
		driver: d,
		db:     s,
	}

	r.setupRoutes()

	return r
}

func (r *Router) Routes() []*Route {
	return r.routes
}

func (r *Router) setupRoutes() {
	r.routes = []*Route{
		NewRoute("GET", "/v1/apps", r.listApps),
		NewRoute("POST", "/v1/apps", r.createApp),
		NewRoute("GET", "/v1/apps/{app_id}", r.getApp),
		NewRoute("DELETE", "/v1/apps/{app_id}", r.deleteApp),
		NewRoute("PATCH", "/v1/apps/{app_id}/scale-up", r.scaleUp),
		NewRoute("PATCH", "/v1/apps/{app_id}/scale-down", r.scaleDown),
		NewRoute("PUT", "/v1/apps/{app_id}", r.updateApp),
		NewRoute("PATCH", "/v1/apps/{app_id}/proceed-update", r.proceedUpdate),
		NewRoute("PATCH", "/v1/apps/{app_id}/cancel-update", r.cancelUpdate),
		NewRoute("PATCH", "/v1/apps/{app_id}/weights", r.updateWeights),
		NewRoute("GET", "/v1/apps/{app_id}/tasks/{task_id}", r.getTask),
		NewRoute("PATCH", "/v1/apps/{app_id}/tasks/{task_id}/weight", r.updateWeight),
		NewRoute("GET", "/v1/{app_id}/versions", r.getVersions),
		NewRoute("GET", "/v1/{app_id}/versions/{version_id}", r.getVersion),
		//NewRoute("GET", "/v1/{app_id}/service-discoveries", r.getAppDiscoveries),
		//NewRoute("GET", "/v1/{app_id}/service-discoveries/md5", r.getAppDiscoveriesMD5),
		//NewRoute("GET", "/v1/service-discoveries", r.getDiscoveriesMD5),

		//NewRoute("POST", "/v1/compose", r.runInstance),
		//NewRoute("POST", "/v1/compose/parse", r.parseYAML),
		//NewRoute("GET", "/v1/compose", r.listInstances),
		//NewRoute("GET", "/v1/compose/{iid}", r.getInstance),
		//NewRoute("DELETE", "/v1/compose/{iid}", r.removeInstance),

		NewRoute("GET", "/v1/framework/info", r.getFrameworkInfo),

		NewRoute("GET", "/v1/ping", r.ping),

		//NewRoute("GET", "/v1/events", r.getEvents),

		NewRoute("GET", "/v1/stats", r.getStats),

		NewRoute("GET", "/v1/version", r.getVersion),
	}
}
