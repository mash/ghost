package ghost

import (
	"net/http"
)

type Server interface {
	Create(http.ResponseWriter, *http.Request) error
	Read(http.ResponseWriter, *http.Request) error
	Update(http.ResponseWriter, *http.Request) error
	Delete(http.ResponseWriter, *http.Request) error
	List(http.ResponseWriter, *http.Request) error
}

type server[R Resource, Q Query, P PKey] struct {
	store      Store[R, Q, P]
	encoding   Encoding[R]
	identifier Identifier[P]
	querier    Querier[Q]
}

func NewServer[R Resource, Q Query, P PKey](store Store[R, Q, P], encoding Encoding[R], identifier Identifier[P], querier Querier[Q]) Server {
	return server[R, Q, P]{
		store:      store,
		encoding:   encoding,
		identifier: identifier,
		querier:    querier,
	}
}

func (g server[R, Q, P]) Create(w http.ResponseWriter, r *http.Request) error {
	res, err := g.encoding.Decode(r)
	if err != nil {
		return err
	}
	if err := g.store.Create(r.Context(), &res); err != nil {
		return err
	}
	return g.encoding.Encode(w, res, http.StatusCreated)
}

func (g server[R, Q, P]) Read(w http.ResponseWriter, r *http.Request) error {
	pkey, err := g.identifier.PKey(r)
	if err != nil {
		return err
	}
	q, err := g.querier.Query(r)
	if err != nil {
		return err
	}
	res, err := g.store.Read(r.Context(), pkey, &q)
	if err != nil {
		return err
	}
	return g.encoding.Encode(w, *res, http.StatusOK)
}

func (g server[R, Q, P]) Update(w http.ResponseWriter, r *http.Request) error {
	pkey, err := g.identifier.PKey(r)
	if err != nil {
		return err
	}
	res, err := g.encoding.Decode(r)
	if err != nil {
		return err
	}
	if err := g.store.Update(r.Context(), pkey, &res); err != nil {
		return err
	}
	return g.encoding.Encode(w, res, http.StatusOK)
}

func (g server[R, Q, P]) Delete(w http.ResponseWriter, r *http.Request) error {
	pkey, err := g.identifier.PKey(r)
	if err != nil {
		return err
	}
	if err := g.store.Delete(r.Context(), pkey); err != nil {
		return err
	}
	return g.encoding.EncodeEmpty(w, http.StatusNoContent)
}

func (g server[R, Q, P]) List(w http.ResponseWriter, r *http.Request) error {
	q, err := g.querier.Query(r)
	if err != nil {
		return err
	}
	res, err := g.store.List(r.Context(), &q)
	if err != nil {
		return err
	}
	return g.encoding.EncodeList(w, res, http.StatusOK)
}
