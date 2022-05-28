package ghost_test

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mash/ghost"
)

type User struct {
	ID   uint64
	Name string
}

func (u *User) PKeys() []ghost.PKey {
	return []ghost.PKey{ghost.PKey(u.ID)}
}

func (u *User) SetPKeys(pkeys []ghost.PKey) {
	u.ID = uint64(pkeys[0])
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
			expectedResBody: `{"ID":1,"Name":"John"}`,
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
