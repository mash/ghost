package ghost

import (
	"context"
	"fmt"
	"strconv"
)

type Store[R Resource, Q Query, P PKey] interface {
	Create(context.Context, *R) error
	Read(context.Context, P, *Q) (*R, error)
	Update(context.Context, P, *R) error
	Delete(context.Context, P) error
	List(context.Context, *Q) ([]R, error)
}

type mapIntStore[R Resource, Q Query, P PIntKey] struct {
	mapStore[R, Q, P]
	nextID P
}

func NewMapStore[R Resource, Q Query, P PIntKey](r R, q Q, p P) Store[R, Q, P] {
	return &mapIntStore[R, Q, P]{
		mapStore: mapStore[R, Q, P]{
			m: make(map[P]*R),
		},
		nextID: 1,
	}
}

func (s *mapIntStore[R, Q, P]) Create(ctx context.Context, r *R) error {
	s.m[s.nextID] = r
	s.nextID++
	return nil
}

type mapStrStore[R Resource, Q Query, P PStrKey] struct {
	mapStore[R, Q, P]
	nextID P
}

func NewMapStrStore[R Resource, Q Query, P PStrKey](r R, q Q, p P) Store[R, Q, P] {
	return &mapStrStore[R, Q, P]{
		mapStore: mapStore[R, Q, P]{
			m: make(map[P]*R),
		},
		nextID: "1",
	}
}

func (s *mapStrStore[R, Q, P]) Create(ctx context.Context, r *R) error {
	s.m[s.nextID] = r
	i, err := strconv.ParseInt(string(s.nextID), 10, 64)
	if err != nil {
		return err
	}
	s.nextID = P(fmt.Sprintf("%d", i+1))
	return nil
}

type mapStore[R Resource, Q Query, P PKey] struct {
	m map[P]*R
}

func (s *mapStore[R, Q, P]) Read(ctx context.Context, pkey P, q *Q) (*R, error) {
	r, ok := s.m[pkey]
	if !ok {
		return r, ErrNotFound
	}
	return r, nil
}

func (s *mapStore[R, Q, P]) Update(ctx context.Context, pkey P, r *R) error {
	_, ok := s.m[pkey]
	if !ok {
		return ErrNotFound
	}
	s.m[pkey] = r
	return nil
}

func (s *mapStore[R, Q, P]) Delete(ctx context.Context, pkey P) error {
	delete(s.m, pkey)
	return nil
}

func (s *mapStore[R, Q, P]) List(ctx context.Context, q *Q) ([]R, error) {
	var r []R
	for _, v := range s.m {
		r = append(r, *v)
	}
	return r, nil
}

type hookStore[R Resource, Q Query, P PKey] struct {
	store Store[R, Q, P]
}

func NewHookStore[R Resource, Q Query, P PKey](store Store[R, Q, P]) Store[R, Q, P] {
	return hookStore[R, Q, P]{
		store: store,
	}
}

type BeforeCreate interface {
	BeforeCreate(context.Context) error
}

type AfterCreate interface {
	AfterCreate(context.Context) error
}

func (s hookStore[R, Q, P]) Create(ctx context.Context, r *R) error {
	if h, ok := any(r).(BeforeCreate); ok {
		if err := h.BeforeCreate(ctx); err != nil {
			return err
		}
	}
	if err := s.store.Create(ctx, r); err != nil {
		return err
	}
	if h, ok := any(r).(AfterCreate); ok {
		if err := h.AfterCreate(ctx); err != nil {
			return err
		}
	}
	return nil
}

type BeforeRead[Q Query, P PKey] interface {
	BeforeRead(context.Context, P, *Q) error
}

type AfterRead[Q Query, P PKey] interface {
	AfterRead(context.Context, P, *Q) error
}

func (s hookStore[R, Q, P]) Read(ctx context.Context, pkey P, q *Q) (*R, error) {
	var r R
	if h, ok := any(&r).(BeforeRead[Q, P]); ok {
		if err := h.BeforeRead(ctx, pkey, q); err != nil {
			return &r, err
		}
	}
	rr, err := s.store.Read(ctx, pkey, q)
	if err != nil {
		return rr, err
	}
	if h, ok := any(rr).(AfterRead[Q, P]); ok {
		if err := h.AfterRead(ctx, pkey, q); err != nil {
			return rr, err
		}
	}
	return rr, nil
}

type BeforeUpdate[P PKey] interface {
	BeforeUpdate(context.Context, P) error
}

type AfterUpdate[P PKey] interface {
	AfterUpdate(context.Context, P) error
}

func (s hookStore[R, Q, P]) Update(ctx context.Context, pkey P, r *R) error {
	if h, ok := any(r).(BeforeUpdate[P]); ok {
		if err := h.BeforeUpdate(ctx, pkey); err != nil {
			return err
		}
	}
	if err := s.store.Update(ctx, pkey, r); err != nil {
		return err
	}
	if h, ok := any(r).(AfterUpdate[P]); ok {
		if err := h.AfterUpdate(ctx, pkey); err != nil {
			return err
		}
	}
	return nil
}

type BeforeDelete[P PKey] interface {
	BeforeDelete(context.Context, P) error
}

type AfterDelete[P PKey] interface {
	AfterDelete(context.Context, P) error
}

func (s hookStore[R, Q, P]) Delete(ctx context.Context, pkey P) error {
	var r R
	if h, ok := any(&r).(BeforeDelete[P]); ok {
		if err := h.BeforeDelete(ctx, pkey); err != nil {
			return err
		}
	}
	if err := s.store.Delete(ctx, pkey); err != nil {
		return err
	}
	if h, ok := any(&r).(AfterDelete[P]); ok {
		if err := h.AfterDelete(ctx, pkey); err != nil {
			return err
		}
	}
	return nil
}

type BeforeList[Q Query] interface {
	BeforeList(context.Context, *Q) error
}

type AfterList[R Resource, Q Query] interface {
	AfterList(context.Context, *Q, []R) error
}

func (s hookStore[R, Q, P]) List(ctx context.Context, q *Q) ([]R, error) {
	var r R
	if h, ok := any(&r).(BeforeList[Q]); ok {
		if err := h.BeforeList(ctx, q); err != nil {
			return nil, err
		}
	}
	l, err := s.store.List(ctx, q)
	if err != nil {
		return l, err
	}
	if h, ok := any(&r).(AfterList[R, Q]); ok {
		if err := h.AfterList(ctx, q, l); err != nil {
			return nil, err
		}
	}
	return l, nil
}
