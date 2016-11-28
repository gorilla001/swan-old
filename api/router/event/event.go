package event

import (
	"net/http"
)

func (r *Router) EventStream(w http.ResponseWriter, req *http.Request) error {
	if err := r.backend.EventStream(w); err != nil {
		return err
	}

	return nil
}
