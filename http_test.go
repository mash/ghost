package ghost_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/go-cmp/cmp"
	"github.com/mash/ghost"
	v "github.com/mash/ghost/store/validator"
)

type User struct {
	Name string
}

type SearchQuery struct {
	Name string
}

func TestHttp(t *testing.T) {
	t.Run("uint64", func(t *testing.T) {
		store := ghost.NewMapStore(User{}, SearchQuery{}, uint64(0))
		g := ghost.New(store)
		testHandler(t, g)
	})
	t.Run("string", func(t *testing.T) {
		store := ghost.NewMapStrStore(User{}, SearchQuery{}, string(""))
		g := ghost.NewS(store)
		testHandler(t, g)
	})
}

func testHandler(t *testing.T, h http.Handler) {
	tests := []struct {
		name, method, path, reqBody string
		expectedCode                int
		expectedResBody             string
	}{
		{
			name:            "POST /",
			method:          "POST",
			path:            "/",
			reqBody:         `{"Name":"John"}`,
			expectedCode:    201,
			expectedResBody: `{"Name":"John"}`,
		}, {
			name:            "Read /1",
			method:          "GET",
			path:            "/1",
			expectedCode:    200,
			expectedResBody: `{"Name":"John"}`,
		}, {
			name:            "PUT /1",
			method:          "PUT",
			path:            "/1",
			reqBody:         `{"Name":"Bob"}`,
			expectedCode:    200,
			expectedResBody: `{"Name":"Bob"}`,
		}, {
			name:            "GET /",
			method:          "GET",
			path:            "/",
			expectedCode:    200,
			expectedResBody: `[{"Name":"Bob"}]`,
		}, {
			name:            "DELETE /1",
			method:          "DELETE",
			path:            "/1",
			expectedCode:    204,
			expectedResBody: `{}`,
		}, {
			name:            "PATCH /1",
			method:          "PATCH",
			path:            "/1",
			expectedCode:    405,
			expectedResBody: `{"error":"Method Not Allowed"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			var body io.Reader
			if test.method != "GET" {
				body = strings.NewReader(test.reqBody)
			}
			r := httptest.NewRequest(test.method, test.path, body)
			h.ServeHTTP(w, r)

			if e, g := test.expectedCode, w.Code; e != g {
				t.Errorf("expected %d, got %d", e, g)
			}
			if e, g := test.expectedResBody, strings.TrimSpace(w.Body.String()); e != g {
				t.Fatalf("expected %s, got %s", e, g)
			}
		})
	}
}

type HookedUser struct {
	Name   string
	Called map[string]int
}

var globalCalled = map[string]int{}

func (u *HookedUser) recordCall(name string) {
	if u.Called == nil {
		u.Called = make(map[string]int)
	}
	u.Called[name]++
}

func (u *HookedUser) BeforeCreate(ctx context.Context) error {
	u.recordCall("BeforeCreate")
	return nil
}

func (u *HookedUser) AfterCreate(ctx context.Context) error {
	u.recordCall("AfterCreate")
	return nil
}

func (u *HookedUser) BeforeRead(ctx context.Context, pkey uint64, q *SearchQuery) error {
	// *u is a zero value resource
	globalCalled["BeforeRead"]++
	return nil
}

func (u *HookedUser) AfterRead(ctx context.Context, pkey uint64, q *SearchQuery) error {
	u.recordCall("AfterRead")
	return nil
}

func (u *HookedUser) BeforeUpdate(ctx context.Context, pkey uint64) error {
	u.recordCall("BeforeUpdate")
	return nil
}

func (u *HookedUser) AfterUpdate(ctx context.Context, pkey uint64) error {
	u.recordCall("AfterUpdate")
	return nil
}

func (u *HookedUser) BeforeDelete(ctx context.Context, pkey uint64) error {
	// *u is a zero value resource
	globalCalled["BeforeDelete"]++
	return nil
}

func (u *HookedUser) AfterDelete(ctx context.Context, pkey uint64) error {
	// *u is a zero value resource
	globalCalled["AfterDelete"]++
	return nil
}

func (u *HookedUser) BeforeList(ctx context.Context, q *SearchQuery) error {
	globalCalled["BeforeList"]++
	return nil
}

func (u *HookedUser) AfterList(ctx context.Context, q *SearchQuery, rr []HookedUser) error {
	globalCalled["AfterList"]++
	return nil
}

func TestHook(t *testing.T) {
	store := ghost.NewMapStore(HookedUser{}, SearchQuery{}, uint64(0))
	g := ghost.New(store)

	tests := []struct {
		name, method, path, reqBody string
		expectedCode                int
		expectedResBody             string
	}{
		{
			name:            "POST /",
			method:          "POST",
			path:            "/",
			reqBody:         `{"Name":"John"}`,
			expectedCode:    201,
			expectedResBody: `{"Name":"John","Called":{"AfterCreate":1,"BeforeCreate":1}}`,
		}, {
			name:            "GET /1",
			method:          "GET",
			path:            "/1",
			expectedCode:    200,
			expectedResBody: `{"Name":"John","Called":{"AfterCreate":1,"AfterRead":1,"BeforeCreate":1}}`,
		}, {
			name:            "PUT /1",
			method:          "PUT",
			path:            "/1",
			reqBody:         `{"Name":"Bob"}`,
			expectedCode:    200,
			expectedResBody: `{"Name":"Bob","Called":{"AfterUpdate":1,"BeforeUpdate":1}}`,
		}, {
			name:            "GET /",
			method:          "GET",
			path:            "/",
			expectedCode:    200,
			expectedResBody: `[{"Name":"Bob","Called":{"AfterUpdate":1,"BeforeUpdate":1}}]`,
		}, {
			name:            "DELETE /1",
			method:          "DELETE",
			path:            "/1",
			expectedCode:    204,
			expectedResBody: `{}`,
		}, {
			name:            "PATCH /1",
			method:          "PATCH",
			path:            "/1",
			expectedCode:    405,
			expectedResBody: `{"error":"Method Not Allowed"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			var body io.Reader
			if test.method != "GET" {
				body = strings.NewReader(test.reqBody)
			}
			r := httptest.NewRequest(test.method, test.path, body)
			g.ServeHTTP(w, r)

			if e, g := test.expectedCode, w.Code; e != g {
				t.Errorf("expected %d, got %d", e, g)
			}
			if e, g := test.expectedResBody, strings.TrimSpace(w.Body.String()); e != g {
				t.Fatalf("expected %s, got %s", e, g)
			}
		})
	}
	if diff := cmp.Diff(map[string]int{
		"BeforeRead":   1,
		"BeforeDelete": 1,
		"AfterDelete":  1,
		"BeforeList":   1,
		"AfterList":    1,
	}, globalCalled); diff != "" {
		t.Errorf("unexpected calls to hooks (-want +got):\n%s", diff)
	}
}

type ValidateUser struct {
	Name string `validate:"required"`
}

type ValidateSearchQuery struct {
	Name string `validate:"required"`
}

func TestValidate(t *testing.T) {
	store := ghost.NewMapStore(ValidateUser{}, ValidateSearchQuery{}, uint64(0))
	validator := validator.New()
	store = v.NewStore(store, validator)
	g := ghost.New(store)

	tests := []struct {
		name, method, path, reqBody string
		expectedCode                int
		expectedResBody             string
	}{
		{
			name:            "POST /",
			method:          "POST",
			path:            "/",
			reqBody:         `{"Name":""}`,
			expectedCode:    400,
			expectedResBody: `{"error":"Key: 'ValidateUser.Name' Error:Field validation for 'Name' failed on the 'required' tag"}`,
		}, {
			name:            "PUT /1",
			method:          "PUT",
			path:            "/1",
			reqBody:         `{"Name":""}`,
			expectedCode:    400,
			expectedResBody: `{"error":"Key: 'ValidateUser.Name' Error:Field validation for 'Name' failed on the 'required' tag"}`,
		}, {
			name:            "GET /",
			method:          "GET",
			path:            "/",
			reqBody:         `{"Name":""}`,
			expectedCode:    400,
			expectedResBody: `{"error":"Key: 'ValidateSearchQuery.Name' Error:Field validation for 'Name' failed on the 'required' tag"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			var body io.Reader
			if test.method != "GET" {
				body = strings.NewReader(test.reqBody)
			}
			r := httptest.NewRequest(test.method, test.path, body)
			g.ServeHTTP(w, r)

			if e, g := test.expectedCode, w.Code; e != g {
				t.Errorf("expected %d, got %d", e, g)
			}
			if e, g := test.expectedResBody, strings.TrimSpace(w.Body.String()); e != g {
				t.Fatalf("expected %s, got %s", e, g)
			}
		})
	}
}
