package ghost

import (
	"net/http"
	"path"
)

type Server[R Resource, Q Query] interface {
	Create(http.ResponseWriter, *http.Request) error
	Read(http.ResponseWriter, *http.Request) error
	Update(http.ResponseWriter, *http.Request) error
	Delete(http.ResponseWriter, *http.Request) error
	List(http.ResponseWriter, *http.Request) error
}

type Handler = func(http.ResponseWriter, *http.Request) error

type Ghost[R Resource, Q Query] struct {
	Store        Store[R, Q]
	Encoding     Encoding[R, Q]
	PKeyer       PKeyer
	Querier      Querier[Q]
	Mux          func(Server[R, Q]) Handler
	ErrorHandler func(error) http.Handler
}

func New[R Resource, Q Query](store Store[R, Q]) http.Handler {
	return Ghost[R, Q]{
		Store:        store,
		Encoding:     JSON,
		PKeyer:       PathKeyer,
		Querier:      NewQuerier,
		Mux:          DefaultMux[R, Q],
		ErrorHandler: DefaultErrorHandler,
	}
}

func (g Ghost[R, Q]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := g.Mux(g)(w, r); err != nil {
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

func DefaultErrorHandler(err error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	})
}

func (g Ghost[R, Q]) Create(w http.ResponseWriter, r *http.Request) error {
	res, err := g.Encoding.Decode(r)
	if err != nil {
		return err
	}
	return g.Store.Create(r.Context(), res)
}

func (g Ghost[R, Q]) Read(w http.ResponseWriter, r *http.Request) error {
	pkeys, err := g.PKeyer.PKeys(r)
	if err != nil {
		return err
	}
	res, err := g.Store.Read(r.Context(), pkeys)
	if err != nil {
		return err
	}
	return g.Encoding.Encode(w, res)
}

func (g Ghost[R, Q]) Update(w http.ResponseWriter, r *http.Request) error {
	res, err := g.Encoding.Decode(r)
	if err != nil {
		return err
	}
	return g.Store.Update(r.Context(), res)
}

func (g Ghost[R, Q]) Delete(w http.ResponseWriter, r *http.Request) error {
	pkeys, err := g.PKeyer.PKeys(r)
	if err != nil {
		return err
	}
	if err := g.Store.Delete(r.Context(), pkeys); err != nil {
		return err
	}
	return g.Encoding.EncodeEmpty(w)
}

func (g Ghost[R, Q]) List(w http.ResponseWriter, r *http.Request) error {
	q, err := g.Querier.Query(r)
	if err != nil {
		return err
	}
	res, err := g.Store.List(r.Context(), q)
	if err != nil {
		return err
	}
	return g.Encoding.EncodeList(w, res)
}
