package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// WriteJSON write response as json format.
func (r *Router) WriteJSON(w http.ResponseWriter, code int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return &httpError{
			errmsg:     err.Error(),
			statuscode: http.StatusInternalServerError,
		}
	}

	return nil
}

// WriteError write error msg to response.
func (r *Router) WriteError(code int, err error) error {
	//w.WriteHeader(code)
	//if _, err := w.Write([]byte(err.Error())); err != nil {
	//	return err
	//}

	//return nil
	return httpError{
		errmsg:     err.Error(),
		statuscode: code,
	}

}

// CheckForJSON makes sure that the request's Content-Type is application/json.
func (r *Router) CheckForJSON(req *http.Request) error {
	if req.Header.Get("Content-Type") != "application/json" {
		return fmt.Errorf("Content-Type must be 'application/json'")
	}

	return nil
}

func (r *Router) decode(b io.ReadCloser, v interface{}) error {
	dec := json.NewDecoder(b)
	return dec.Decode(&v)
}
