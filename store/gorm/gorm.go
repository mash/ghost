package gorm

import (
	"context"

	"github.com/mash/ghost"
	"gorm.io/gorm"
)

type gormStore[R ghost.Resource, Q ghost.Query] struct {
	db *gorm.DB
}

func NewStore[R ghost.Resource, Q ghost.Query](r R, q Q, db *gorm.DB) ghost.Store[R, Q] {
	return gormStore[R, Q]{
		db: db,
	}
}

type Create[R ghost.Resource] interface {
	Create(context.Context, *gorm.DB, *R) error
}

func (s gormStore[R, Q]) Create(ctx context.Context, r *R) error {
	if rr, ok := any(r).(Create[R]); ok {
		return rr.Create(ctx, s.db, r)
	}

	result := s.db.Create(&r)
	return result.Error
}

type Read[R ghost.Resource, Q ghost.Query] interface {
	Read(context.Context, *gorm.DB, []ghost.PKey, *Q) (*R, error)
}

func (s gormStore[R, Q]) Read(ctx context.Context, pkeys []ghost.PKey, q *Q) (*R, error) {
	var r R
	if rr, ok := any(&r).(Read[R, Q]); ok {
		return rr.Read(ctx, s.db, pkeys, q)
	}

	result := s.db.First(&r, pkeys[0])
	return &r, result.Error
}

type Update[R ghost.Resource] interface {
	Update(context.Context, *gorm.DB, []ghost.PKey, *R) error
}

func (s gormStore[R, Q]) Update(ctx context.Context, pkeys []ghost.PKey, r *R) error {
	if rr, ok := any(r).(Update[R]); ok {
		return rr.Update(ctx, s.db, pkeys, r)
	}

	var orig R
	result := s.db.Find(&orig, pkeys)
	if result.Error != nil {
		return result.Error
	}

	result = s.db.Model(&orig).Updates(&r)
	return result.Error
}

type Delete interface {
	Delete(context.Context, *gorm.DB, []ghost.PKey) error
}

func (s gormStore[R, Q]) Delete(ctx context.Context, pkeys []ghost.PKey) error {
	var r R
	if rr, ok := any(&r).(Delete); ok {
		return rr.Delete(ctx, s.db, pkeys)
	}

	result := s.db.Delete(&r, pkeys)
	return result.Error
}

type List[R ghost.Resource, Q ghost.Query] interface {
	List(context.Context, *gorm.DB, *Q) ([]R, error)
}

func (s gormStore[R, Q]) List(ctx context.Context, q *Q) ([]R, error) {
	var r R
	if rp, ok := any(&r).(List[R, Q]); ok {
		return rp.List(ctx, s.db, q)
	}

	var rr []R
	result := s.db.Order("id desc").Find(&rr)
	return rr, result.Error
}
