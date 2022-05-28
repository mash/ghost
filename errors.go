package ghost

import (
	"errors"
	"net/http"
)

var errNotFound = errors.New("not found")

func DefaultErrorHandler(err error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	})
}
