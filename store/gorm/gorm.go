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

func (s gormStore[R, Q]) Create(ctx context.Context, r R) error {
	// TODO call method on r if r implements an interface

	result := s.db.Create(&r)
	return result.Error
}

func (s gormStore[R, Q]) Read(ctx context.Context, pkeys []ghost.PKey, q Q) (R, error) {
	var r R

	// TODO call method on r if r implements an interface

	result := s.db.First(&r, pkeys[0])
	return r, result.Error
}

func (s gormStore[R, Q]) Update(ctx context.Context, pkeys []ghost.PKey, r R) error {
	var orig R
	result := s.db.Find(&orig, pkeys)
	if result.Error != nil {
		return result.Error
	}

	// TODO call method on r if r implements an interface

	result = s.db.Model(&orig).Updates(&r)
	return result.Error
}

func (s gormStore[R, Q]) Delete(ctx context.Context, pkeys []ghost.PKey) error {
	var r R

	// TODO call method on r if r implements an interface

	result := s.db.Delete(&r, pkeys)
	return result.Error
}

func (s gormStore[R, Q]) List(ctx context.Context, q Q) ([]R, error) {
	// TODO
	return nil, nil
}
