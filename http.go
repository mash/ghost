package ghost

import (
	"net/http"
	"path"
)

type Handler = func(http.ResponseWriter, *http.Request) error

func DefaultMux[R Resource, Q Query](s Server) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		switch r.Method {
		case http.MethodPost:
			return s.Create(w, r)
		case http.MethodPut:
			return s.Update(w, r)
		case http.MethodDelete:
			return s.Delete(w, r)
		case http.MethodGet:
			_, f := path.Split(r.URL.Path)
			if f == "" {
				return s.List(w, r)
			}
			return s.Read(w, r)
		default:
			return ErrMethodNotAllowed
		}
	}
}
