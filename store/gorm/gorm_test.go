package gorm_test

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mash/ghost"
	ggorm "github.com/mash/ghost/store/gorm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name string
}

type SearchQuery struct {
	Name string
}

func TestHttp(t *testing.T) {
	_ = os.Remove("test.db")
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// create the table
	db.AutoMigrate(&User{})

	store := ggorm.NewStore(User{}, SearchQuery{}, db)
	g := ghost.New(store)

	ignore := cmpopts.IgnoreFields(User{}, "Model")

	testUser := func(t *testing.T, expected User, r io.Reader) {
		t.Helper()

		g := User{}
		if err := json.NewDecoder(r).Decode(&g); err != nil {
			t.Errorf("failed to decode json body: %v", err)
		}
		if diff := cmp.Diff(expected, g, ignore); diff != "" {
			t.Errorf("unexpected response body (-expected +got):\n%s", diff)
		}
	}

	tests := []struct {
		name, method, path, reqBody string
		expectedCode                int
		testResBody                 func(t *testing.T, resBody io.Reader)
	}{
		{
			name:         "POST /",
			method:       "POST",
			path:         "/",
			reqBody:      `{"Name":"John"}`,
			expectedCode: 201,
			testResBody: func(t *testing.T, resBody io.Reader) {
				e := User{
					Name: "John",
				}
				testUser(t, e, resBody)
			},
		}, {
			name:         "Read /1",
			method:       "GET",
			path:         "/1",
			expectedCode: 200,
			testResBody: func(t *testing.T, resBody io.Reader) {
				e := User{
					Name: "John",
				}
				testUser(t, e, resBody)
			},
		}, {
			name:         "PUT /1",
			method:       "PUT",
			path:         "/1",
			reqBody:      `{"Name":"Bob"}`,
			expectedCode: 200,
			testResBody: func(t *testing.T, resBody io.Reader) {
				e := User{
					Name: "Bob",
				}
				testUser(t, e, resBody)
			},
		}, {
			name:         "GET /",
			method:       "GET",
			path:         "/",
			expectedCode: 200,
			testResBody: func(t *testing.T, resBody io.Reader) {
				e := []User{
					{
						Name: "Bob",
					},
				}
				g := []User{}
				if err := json.NewDecoder(resBody).Decode(&g); err != nil {
					t.Errorf("failed to decode json body: %v", err)
				}
				if diff := cmp.Diff(e, g, ignore); diff != "" {
					t.Errorf("unexpected response body (-expected +got):\n%s", diff)
				}
			},
		}, {
			name:         "DELETE /1",
			method:       "DELETE",
			path:         "/1",
			expectedCode: 204,
			testResBody: func(t *testing.T, resBody io.Reader) {
				e := map[string]any{}
				g := map[string]any{}
				if err := json.NewDecoder(resBody).Decode(&g); err != nil {
					t.Errorf("failed to decode json body: %v", err)
				}
				if diff := cmp.Diff(e, g); diff != "" {
					t.Errorf("unexpected response body (-expected +got):\n%s", diff)
				}
			},
		}, {
			name:         "PATCH /1",
			method:       "PATCH",
			path:         "/1",
			expectedCode: 405,
			testResBody: func(t *testing.T, resBody io.Reader) {
				e := map[string]any{
					"error": "Method Not Allowed",
				}
				g := map[string]any{}
				if err := json.NewDecoder(resBody).Decode(&g); err != nil {
					t.Errorf("failed to decode json body: %v", err)
				}
				if diff := cmp.Diff(e, g); diff != "" {
					t.Errorf("unexpected response body (-expected +got):\n%s", diff)
				}
			},
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
				t.Errorf("expected %d, got %d, body: %s", e, g, w.Body.String())
			}
			test.testResBody(t, w.Body)
		})
	}
}
