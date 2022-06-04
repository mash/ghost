package gorm

import (
	"context"

	"github.com/mash/ghost"
	"gorm.io/gorm"
)

type gormStore[R ghost.Resource, Q ghost.Query, P ghost.PKey] struct {
	db *gorm.DB
}

func NewStore[R ghost.Resource, Q ghost.Query, P ghost.PKey](r R, q Q, p P, db *gorm.DB) ghost.Store[R, Q, P] {
	return gormStore[R, Q, P]{
		db: db,
	}
}

type Create[R ghost.Resource] interface {
	Create(context.Context, *gorm.DB, *R) error
}

func (s gormStore[R, Q, P]) Create(ctx context.Context, r *R) error {
	if rr, ok := any(r).(Create[R]); ok {
		return rr.Create(ctx, s.db, r)
	}

	result := s.db.Create(&r)
	return result.Error
}

type Read[R ghost.Resource, Q ghost.Query, P ghost.PKey] interface {
	Read(context.Context, *gorm.DB, P, *Q) (*R, error)
}

func (s gormStore[R, Q, P]) Read(ctx context.Context, pkey P, q *Q) (*R, error) {
	var r R
	if rr, ok := any(&r).(Read[R, Q, P]); ok {
		return rr.Read(ctx, s.db, pkey, q)
	}

	result := s.db.First(&r, pkey)
	return &r, result.Error
}

type Update[R ghost.Resource, P ghost.PKey] interface {
	Update(context.Context, *gorm.DB, P, *R) error
}

func (s gormStore[R, Q, P]) Update(ctx context.Context, pkey P, r *R) error {
	if rr, ok := any(r).(Update[R, P]); ok {
		return rr.Update(ctx, s.db, pkey, r)
	}

	var orig R
	result := s.db.Find(&orig, pkey)
	if result.Error != nil {
		return result.Error
	}

	result = s.db.Model(&orig).Updates(&r)
	return result.Error
}

type Delete[P ghost.PKey] interface {
	Delete(context.Context, *gorm.DB, P) error
}

func (s gormStore[R, Q, P]) Delete(ctx context.Context, pkey P) error {
	var r R
	if rr, ok := any(&r).(Delete[P]); ok {
		return rr.Delete(ctx, s.db, pkey)
	}

	result := s.db.Delete(&r, pkey)
	return result.Error
}

type List[R ghost.Resource, Q ghost.Query] interface {
	List(context.Context, *gorm.DB, *Q) ([]R, error)
}

func (s gormStore[R, Q, P]) List(ctx context.Context, q *Q) ([]R, error) {
	var r R
	if rp, ok := any(&r).(List[R, Q]); ok {
		return rp.List(ctx, s.db, q)
	}

	var rr []R
	result := s.db.Order("id desc").Find(&rr)
	return rr, result.Error
}
