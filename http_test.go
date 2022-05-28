package ghost_test

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mash/ghost"
)

type User struct {
	Name string
}

type SearchQuery struct {
	Name string
}

func TestHttp(t *testing.T) {
	store := ghost.NewMapStore(&User{}, SearchQuery{})
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
			expectedResBody: `{"error":"Not Allowed"}`,
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
