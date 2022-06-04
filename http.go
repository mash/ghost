package ghost

import (
	"net/http"
	"path"
)

type Handler = func(http.ResponseWriter, *http.Request) error

type Ghost[R Resource, Q Query, P PKey] struct {
	Server       Server
	Mux          func(Server) Handler
	ErrorHandler func(error) http.Handler
}

// New returns a http.Handler.
// New requires PKey to be an integer.
func New[R Resource, Q Query, P PIntKey](store Store[R, Q, P]) http.Handler {
	store = NewHookStore(store)
	return Ghost[R, Q, P]{
		Server:       NewServer[R, Q, P](store, JSON[R]{}, PathIdentifier[P](IntPath[P]), NewQueryParser[Q]()),
		Mux:          DefaultMux[R, Q],
		ErrorHandler: DefaultErrorHandler(JSON[Error]{}),
	}
}

func (g Ghost[R, Q, P]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := g.Mux(g.Server)(w, r); err != nil {
		g.ErrorHandler(err).ServeHTTP(w, r)
	}
}

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
