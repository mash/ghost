package ghost

import (
	"errors"
	"fmt"
	"net/http"
)

type Error struct {
	Code int    `json:"-"`
	Err  string `json:"error"`
}

func (e Error) Error() string {
	return fmt.Sprintf("code=%d, err=%s", e.Code, e.Err)
}

var ErrNotFound = Error{
	Code: http.StatusNotFound,
	Err:  "Not Found",
}

var ErrMethodNotAllowed = Error{
	Code: http.StatusMethodNotAllowed,
	Err:  "Not Allowed",
}

func DefaultErrorHandler(encoding Encoding[Error]) func(err error) http.Handler {
	return func(err error) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var e Error
			if errors.As(err, &e) {
				encoding.Encode(w, e, e.Code)
				return
			}
			encoding.Encode(w, Error{Err: err.Error()}, http.StatusInternalServerError)
		})
	}
}
