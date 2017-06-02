package api

import (
	"net/http"
)

func (r *Router) ping(w http.ResponseWriter, req *http.Request) error {
	return r.WriteJSON(w, http.StatusOK, "pong")
}
