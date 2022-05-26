package ghost

import "context"

type Store[R Resource, Q Query] interface {
	Create(context.Context, R) error
	Read(context.Context, []PKey) (R, error)
	Update(context.Context, R) error
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
	r.SetPKeys([]PKey{s.nextID})
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

func (s *mapStore[R, Q]) Update(ctx context.Context, r R) error {
	_, ok := s.m[r.PKeys()[0]]
	if !ok {
		return errNotFound
	}
	s.m[r.PKeys()[0]] = r
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
