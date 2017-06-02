package api

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/version"
)

func (r *Router) getVersion(w http.ResponseWriter, req *http.Request) error {
	return r.WriteJSON(w, http.StatusOK, version.GetVersion())
}
