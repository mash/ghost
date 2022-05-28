package ghost

import (
	"net/http"
	"path"
)

type Handler = func(http.ResponseWriter, *http.Request) error

type Ghost[R Resource, Q Query] struct {
	Server       Server[R, Q]
	Mux          func(Server[R, Q]) Handler
	ErrorHandler func(error) http.Handler
}

func New[R Resource, Q Query](store Store[R, Q]) http.Handler {
	store = NewHookStore(store)
	return Ghost[R, Q]{
		Server:       NewServer[R, Q](store, JSON[R]{}, PathIdentifier{}, NewQueryParser[Q]()),
		Mux:          DefaultMux[R, Q],
		ErrorHandler: DefaultErrorHandler,
	}
}

func (g Ghost[R, Q]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := g.Mux(g.Server)(w, r); err != nil {
		g.ErrorHandler(err).ServeHTTP(w, r)
	}
}

func DefaultMux[R Resource, Q Query](s Server[R, Q]) Handler {
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
			return http.ErrNotSupported
		}
	}
}
