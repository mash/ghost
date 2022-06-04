package ghost

import "net/http"

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

// NewS returns a http.Handler.
// NewS requires PKey to be a string.
func NewS[R Resource, Q Query, P PStrKey](store Store[R, Q, P]) http.Handler {
	store = NewHookStore(store)
	return Ghost[R, Q, P]{
		Server:       NewServer[R, Q, P](store, JSON[R]{}, PathIdentifier[P](StrPath[P]), NewQueryParser[Q]()),
		Mux:          DefaultMux[R, Q],
		ErrorHandler: DefaultErrorHandler(JSON[Error]{}),
	}
}

func (g Ghost[R, Q, P]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := g.Mux(g.Server)(w, r); err != nil {
		g.ErrorHandler(err).ServeHTTP(w, r)
	}
}
