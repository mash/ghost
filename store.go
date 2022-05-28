package ghost

import "context"

type Store[R Resource, Q Query] interface {
	Create(context.Context, R) error
	Read(context.Context, []PKey) (R, error)
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

func (s *mapStore[R, Q]) Read(ctx context.Context, pkeys []PKey) (R, error) {
	r, ok := s.m[pkeys[0]]
	if !ok {
		return r, errNotFound
	}
	return r, nil
}

func (s *mapStore[R, Q]) Update(ctx context.Context, pkeys []PKey, r R) error {
	_, ok := s.m[pkeys[0]]
	if !ok {
		return errNotFound
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
	BeforeCreate(context.Context, R) error
}

type AfterCreate[R Resource] interface {
	AfterCreate(context.Context, R) error
}

func (s hookStore[R, Q]) Create(ctx context.Context, r R) error {
	if h, ok := any(r).(BeforeCreate[R]); ok {
		if err := h.BeforeCreate(ctx, r); err != nil {
			return err
		}
	}
	if err := s.store.Create(ctx, r); err != nil {
		return err
	}
	if h, ok := any(r).(AfterCreate[R]); ok {
		if err := h.AfterCreate(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

type BeforeRead[R Resource] interface {
	BeforeRead(context.Context, []PKey) error
}

type AfterRead[R Resource] interface {
	AfterRead(context.Context, []PKey, R) error
}

func (s hookStore[R, Q]) Read(ctx context.Context, pkeys []PKey) (R, error) {
	var r R
	if h, ok := any(r).(BeforeRead[R]); ok {
		if err := h.BeforeRead(ctx, pkeys); err != nil {
			return r, err
		}
	}
	r, err := s.store.Read(ctx, pkeys)
	if err != nil {
		return r, err
	}
	if h, ok := any(r).(AfterRead[R]); ok {
		if err := h.AfterRead(ctx, pkeys, r); err != nil {
			return r, err
		}
	}
	return r, nil
}

type BeforeUpdate[R Resource] interface {
	BeforeUpdate(context.Context, []PKey, R) error
}

type AfterUpdate[R Resource] interface {
	AfterUpdate(context.Context, []PKey, R) error
}

func (s hookStore[R, Q]) Update(ctx context.Context, pkeys []PKey, r R) error {
	if h, ok := any(r).(BeforeUpdate[R]); ok {
		if err := h.BeforeUpdate(ctx, pkeys, r); err != nil {
			return err
		}
	}
	if err := s.store.Update(ctx, pkeys, r); err != nil {
		return err
	}
	if h, ok := any(r).(AfterUpdate[R]); ok {
		if err := h.AfterUpdate(ctx, pkeys, r); err != nil {
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