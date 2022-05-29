package ghost

import (
	"net/http"
)

type Server[R Resource, Q Query] interface {
	Create(http.ResponseWriter, *http.Request) error
	Read(http.ResponseWriter, *http.Request) error
	Update(http.ResponseWriter, *http.Request) error
	Delete(http.ResponseWriter, *http.Request) error
	List(http.ResponseWriter, *http.Request) error
}

type server[R Resource, Q Query] struct {
	store      Store[R, Q]
	encoding   Encoding[R]
	identifier Identifier
	querier    Querier[Q]
}

func NewServer[R Resource, Q Query](store Store[R, Q], encoding Encoding[R], identifier Identifier, querier Querier[Q]) Server[R, Q] {
	return server[R, Q]{
		store:      store,
		encoding:   encoding,
		identifier: identifier,
		querier:    querier,
	}
}

func (g server[R, Q]) Create(w http.ResponseWriter, r *http.Request) error {
	res, err := g.encoding.Decode(r)
	if err != nil {
		return err
	}
	if err := g.store.Create(r.Context(), res); err != nil {
		return err
	}
	return g.encoding.Encode(w, res, http.StatusCreated)
}

func (g server[R, Q]) Read(w http.ResponseWriter, r *http.Request) error {
	pkeys, err := g.identifier.PKeys(r)
	if err != nil {
		return err
	}
	q, err := g.querier.Query(r)
	if err != nil {
		return err
	}
	res, err := g.store.Read(r.Context(), pkeys, q)
	if err != nil {
		return err
	}
	return g.encoding.Encode(w, res, http.StatusOK)
}

func (g server[R, Q]) Update(w http.ResponseWriter, r *http.Request) error {
	pkeys, err := g.identifier.PKeys(r)
	if err != nil {
		return err
	}
	res, err := g.encoding.Decode(r)
	if err != nil {
		return err
	}
	if err := g.store.Update(r.Context(), pkeys, res); err != nil {
		return err
	}
	return g.encoding.Encode(w, res, http.StatusOK)
}

func (g server[R, Q]) Delete(w http.ResponseWriter, r *http.Request) error {
	pkeys, err := g.identifier.PKeys(r)
	if err != nil {
		return err
	}
	if err := g.store.Delete(r.Context(), pkeys); err != nil {
		return err
	}
	return g.encoding.EncodeEmpty(w, http.StatusNoContent)
}

func (g server[R, Q]) List(w http.ResponseWriter, r *http.Request) error {
	q, err := g.querier.Query(r)
	if err != nil {
		return err
	}
	res, err := g.store.List(r.Context(), q)
	if err != nil {
		return err
	}
	return g.encoding.EncodeList(w, res, http.StatusOK)
}
