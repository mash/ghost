package validator

import (
	"context"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/mash/ghost"
)

type validatorStore[R ghost.Resource, Q ghost.Query, P ghost.PKey] struct {
	store    ghost.Store[R, Q, P]
	validate *validator.Validate
}

func NewStore[R ghost.Resource, Q ghost.Query, P ghost.PKey](store ghost.Store[R, Q, P], validator *validator.Validate) ghost.Store[R, Q, P] {
	return validatorStore[R, Q, P]{
		store:    store,
		validate: validator,
	}
}

func (s validatorStore[R, Q, P]) Create(ctx context.Context, r *R) error {
	if err := s.validate.StructCtx(ctx, r); err != nil {
		return validationError(err)
	}
	return s.store.Create(ctx, r)
}

func (s validatorStore[R, Q, P]) Read(ctx context.Context, pkey P, q *Q) (*R, error) {
	return s.store.Read(ctx, pkey, q)
}

func (s validatorStore[R, Q, P]) Update(ctx context.Context, pkey P, r *R) error {
	if err := s.validate.StructCtx(ctx, r); err != nil {
		return validationError(err)
	}
	return s.store.Update(ctx, pkey, r)
}

func (s validatorStore[R, Q, P]) Delete(ctx context.Context, pkey P) error {
	return s.store.Delete(ctx, pkey)
}

func (s validatorStore[R, Q, P]) List(ctx context.Context, q *Q) ([]R, error) {
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
