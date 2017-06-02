package api

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/types"
)

func (r *Router) getFrameworkInfo(w http.ResponseWriter, req *http.Request) error {
	info := new(types.FrameworkInfo)

	return r.WriteJSON(w, http.StatusOK, info)
}
