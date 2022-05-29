package validator

import (
	"context"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/mash/ghost"
)

type validatorStore[R ghost.Resource, Q ghost.Query] struct {
	store    ghost.Store[R, Q]
	validate *validator.Validate
}

func NewStore[R ghost.Resource, Q ghost.Query](store ghost.Store[R, Q], validator *validator.Validate) ghost.Store[R, Q] {
	return validatorStore[R, Q]{
		store:    store,
		validate: validator,
	}
}

func (s validatorStore[R, Q]) Create(ctx context.Context, r R) error {
	if err := s.validate.StructCtx(ctx, r); err != nil {
		return validationError(err)
	}
	return s.store.Create(ctx, r)
}

func (s validatorStore[R, Q]) Read(ctx context.Context, pkeys []ghost.PKey, q Q) (R, error) {
	return s.store.Read(ctx, pkeys, q)
}

func (s validatorStore[R, Q]) Update(ctx context.Context, pkeys []ghost.PKey, r R) error {
	if err := s.validate.StructCtx(ctx, r); err != nil {
		return validationError(err)
	}
	return s.store.Update(ctx, pkeys, r)
}

func (s validatorStore[R, Q]) Delete(ctx context.Context, pkeys []ghost.PKey) error {
	return s.store.Delete(ctx, pkeys)
}

func (s validatorStore[R, Q]) List(ctx context.Context, q Q) ([]R, error) {
	if err := s.validate.StructCtx(ctx, q); err != nil {
		return nil, validationError(err)
	}
	return s.store.List(ctx, q)
}

func validationError(err error) ghost.Error {
	return ghost.Error{
		Code: http.StatusBadRequest,
		Err:  err,
	}
}
