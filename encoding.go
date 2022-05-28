package ghost

import (
	"encoding/json"
	"net/http"
)

type Encoding[R Resource] interface {
	Encode(http.ResponseWriter, R, int) error
	EncodeList(http.ResponseWriter, []R, int) error
	EncodeEmpty(http.ResponseWriter, int) error
	Decode(*http.Request) (R, error)
}

// JSON is an Encoding.
type JSON[R Resource] struct{}

func (j JSON[R]) Encode(w http.ResponseWriter, r R, code int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	return enc.Encode(r)
}

func (j JSON[R]) EncodeList(w http.ResponseWriter, rs []R, code int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	return enc.Encode(rs)
}

func (j JSON[R]) EncodeEmpty(w http.ResponseWriter, code int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	return enc.Encode(map[string]interface{}{})
}

func (j JSON[R]) Decode(r *http.Request) (R, error) {
	var rr R
	err := json.NewDecoder(r.Body).Decode(&rr)
	return rr, err
}
