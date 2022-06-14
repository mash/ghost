package gorm_test

import (
	"context"
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

	store := ggorm.NewStore(User{}, SearchQuery{}, uint64(0), db)
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

type HookedUser struct {
	gorm.Model
	Name   string
	Called map[string]int `gorm:"-"`
}

var globalCalled = map[string]int{}

func (u *HookedUser) recordCall(name string) {
	if u.Called == nil {
		u.Called = make(map[string]int)
	}
	u.Called[name]++
}

// implements ggorm.Create interface
func (h *HookedUser) Create(ctx context.Context, db *gorm.DB) error {
	h.recordCall("Create")
	return db.Create(h).Error
}

// implements ggorm.Read interface
func (h *HookedUser) Read(ctx context.Context, db *gorm.DB, pkey uint64, q *SearchQuery) (*HookedUser, error) {
	var r HookedUser
	r.recordCall("Read")

	result := db.First(&r, pkey)
	return &r, result.Error
}

// implements ggorm.Update interface
func (h *HookedUser) Update(ctx context.Context, db *gorm.DB, pkey uint64) error {
	h.recordCall("Update")

	var orig HookedUser
	result := db.Find(&orig, pkey)
	if result.Error != nil {
		return result.Error
	}

	result = db.Model(&orig).Updates(&h)
	return result.Error
}

// implements ggorm.Delete interface
func (h *HookedUser) Delete(ctx context.Context, db *gorm.DB, pkey uint64) error {
	globalCalled["Delete"]++
	var r HookedUser
	result := db.Delete(&r, pkey)
	return result.Error
}

// implements ggorm.List interface
func (h *HookedUser) List(ctx context.Context, db *gorm.DB, q *SearchQuery) ([]HookedUser, error) {
	globalCalled["List"]++
	var r []HookedUser
	result := db.Find(&r)
	return r, result.Error
}

func TestHook(t *testing.T) {
	_ = os.Remove("hook.db")
	db, err := gorm.Open(sqlite.Open("hook.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// create the table
	db.AutoMigrate(&HookedUser{})

	store := ggorm.NewStore(HookedUser{}, SearchQuery{}, uint64(0), db)
	g := ghost.New(store)

	ignore := cmpopts.IgnoreFields(HookedUser{}, "Model")

	testUser := func(t *testing.T, expected HookedUser, r io.Reader) {
		t.Helper()

		g := HookedUser{}
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
				e := HookedUser{
					Name: "John",
					Called: map[string]int{
						"Create": 1,
					},
				}
				testUser(t, e, resBody)
			},
		}, {
			name:         "Read /1",
			method:       "GET",
			path:         "/1",
			expectedCode: 200,
			testResBody: func(t *testing.T, resBody io.Reader) {
				e := HookedUser{
					Name: "John",
					Called: map[string]int{
						"Read": 1,
					},
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
				e := HookedUser{
					Name: "Bob",
					Called: map[string]int{
						"Update": 1,
					},
				}
				testUser(t, e, resBody)
			},
		}, {
			name:         "GET /",
			method:       "GET",
			path:         "/",
			expectedCode: 200,
			testResBody: func(t *testing.T, resBody io.Reader) {
				e := []HookedUser{
					{
						Name: "Bob",
					},
				}
				g := []HookedUser{}
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
	if diff := cmp.Diff(map[string]int{
		"Delete": 1,
		"List":   1,
	}, globalCalled); diff != "" {
		t.Errorf("unexpected calls to hooks (-want +got):\n%s", diff)
	}
}
