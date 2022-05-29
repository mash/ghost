package ghost

import (
	"context"
)

type Store[R Resource, Q Query] interface {
	Create(context.Context, R) error
	Read(context.Context, []PKey, Q) (R, error)
	Update(context.Context, []PKey, R) error
	Delete(context.Context, []PKey) error
	List(context.Context, Q) ([]R, error)
}

type mapStore[R Resource, Q Query] struct {
	m      map[PKey]R
	nextID PKey
}

func NewMapStore[R Resource, Q Query](r R, q Q) Store[R, Q] {
	return &mapStore[R, Q]{
		m:      map[PKey]R{},
		nextID: 1,
	}
}

func (s *mapStore[R, Q]) Create(ctx context.Context, r R) error {
	s.m[s.nextID] = r
	s.nextID++
	return nil
}

func (s *mapStore[R, Q]) Read(ctx context.Context, pkeys []PKey, q Q) (R, error) {
	r, ok := s.m[pkeys[0]]
	if !ok {
		return r, ErrNotFound
	}
	return r, nil
}

func (s *mapStore[R, Q]) Update(ctx context.Context, pkeys []PKey, r R) error {
	_, ok := s.m[pkeys[0]]
	if !ok {
		return ErrNotFound
	}
	s.m[pkeys[0]] = r
	return nil
}

func (s *mapStore[R, Q]) Delete(ctx context.Context, pkeys []PKey) error {
	delete(s.m, pkeys[0])
	return nil
}

func (s *mapStore[R, Q]) List(ctx context.Context, q Q) ([]R, error) {
	var r []R
	for _, v := range s.m {
		r = append(r, v)
	}
	return r, nil
}

type hookStore[R Resource, Q Query] struct {
	store Store[R, Q]
}

func NewHookStore[R Resource, Q Query](store Store[R, Q]) Store[R, Q] {
	return hookStore[R, Q]{
		store: store,
	}
}

type BeforeCreate[R Resource] interface {
	BeforeCreate(context.Context) error
}

type AfterCreate[R Resource] interface {
	AfterCreate(context.Context) error
}

func (s hookStore[R, Q]) Create(ctx context.Context, r R) error {
	if h, ok := any(r).(BeforeCreate[R]); ok {
		if err := h.BeforeCreate(ctx); err != nil {
			return err
		}
	}
	if err := s.store.Create(ctx, r); err != nil {
		return err
	}
	if h, ok := any(r).(AfterCreate[R]); ok {
		if err := h.AfterCreate(ctx); err != nil {
			return err
		}
	}
	return nil
}

type BeforeRead[R Resource, Q Query] interface {
	BeforeRead(context.Context, []PKey, Q) error
}

type AfterRead[R Resource, Q Query] interface {
	AfterRead(context.Context, []PKey, Q) error
}

func (s hookStore[R, Q]) Read(ctx context.Context, pkeys []PKey, q Q) (R, error) {
	var r R
	if h, ok := any(r).(BeforeRead[R, Q]); ok {
		if err := h.BeforeRead(ctx, pkeys, q); err != nil {
			return r, err
		}
	}
	r, err := s.store.Read(ctx, pkeys, q)
	if err != nil {
		return r, err
	}
	if h, ok := any(r).(AfterRead[R, Q]); ok {
		if err := h.AfterRead(ctx, pkeys, q); err != nil {
			return r, err
		}
	}
	return r, nil
}

type BeforeUpdate[R Resource] interface {
	BeforeUpdate(context.Context, []PKey) error
}

type AfterUpdate[R Resource] interface {
	AfterUpdate(context.Context, []PKey) error
}

func (s hookStore[R, Q]) Update(ctx context.Context, pkeys []PKey, r R) error {
	if h, ok := any(r).(BeforeUpdate[R]); ok {
		if err := h.BeforeUpdate(ctx, pkeys); err != nil {
			return err
		}
	}
	if err := s.store.Update(ctx, pkeys, r); err != nil {
		return err
	}
	if h, ok := any(r).(AfterUpdate[R]); ok {
		if err := h.AfterUpdate(ctx, pkeys); err != nil {
			return err
		}
	}
	return nil
}

type BeforeDelete[R Resource] interface {
	BeforeDelete(context.Context, []PKey) error
}

type AfterDelete[R Resource] interface {
	AfterDelete(context.Context, []PKey) error
}

func (s hookStore[R, Q]) Delete(ctx context.Context, pkeys []PKey) error {
	var r R
	if h, ok := any(r).(BeforeDelete[R]); ok {
		if err := h.BeforeDelete(ctx, pkeys); err != nil {
			return err
		}
	}
	if err := s.store.Delete(ctx, pkeys); err != nil {
		return err
	}
	if h, ok := any(r).(AfterDelete[R]); ok {
		if err := h.AfterDelete(ctx, pkeys); err != nil {
			return err
		}
	}
	return nil
}

type BeforeList[R Resource, Q Query] interface {
	BeforeList(context.Context, Q) error
}

type AfterList[R Resource, Q Query] interface {
	AfterList(context.Context, Q, []R) error
}

func (s hookStore[R, Q]) List(ctx context.Context, q Q) ([]R, error) {
	var r R
	if h, ok := any(r).(BeforeList[R, Q]); ok {
		if err := h.BeforeList(ctx, q); err != nil {
			return nil, err
		}
	}
	l, err := s.store.List(ctx, q)
	if err != nil {
		return l, err
	}
	if h, ok := any(r).(AfterList[R, Q]); ok {
		if err := h.AfterList(ctx, q, l); err != nil {
			return nil, err
		}
	}
	return l, nil
}
