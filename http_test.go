package ghost_test

import (
	"net/http/httptest"
	"testing"

	"github.com/mash/ghost"
)

type User struct {
	ID   uint64
	Name string
}

func (u User) PKeys() []ghost.PKey {
	return []ghost.PKey{ghost.PKey(u.ID)}
}

func (u User) SetPKeys(pkeys []ghost.PKey) {
	u.ID = uint64(pkeys[0])
}

type SearchQuery struct {
	Name string
}

func TestHttp(t *testing.T) {

	store := ghost.NewMapStore(User{}, SearchQuery{})
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
			reqBody:         `{"Name":"John", "Age":30}`,
			expectedCode:    201,
			expectedResBody: `{"ID":1, "Name":"John", "Age":30}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(test.method, test.path, nil)
			g.ServeHTTP(w, r)

			if e, g := test.expectedCode, w.Code; e != g {
				t.Fatalf("expected %d, got %d", e, g)
			}
			if e, g := test.expectedResBody, w.Body.String(); e != g {
				t.Fatalf("expected %s, got %s", e, g)
			}
		})
	}
}
