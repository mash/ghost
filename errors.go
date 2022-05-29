package ghost

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type Error struct {
	Code int   `json:"-"`
	Err  error `json:"error"`
}

func (e Error) Error() string {
	return fmt.Sprintf("code=%d, err=%s", e.Code, e.Err.Error())
}

func (e Error) MarshalJSON() ([]byte, error) {
	b := []byte(`{"error":`)
	be, err := json.Marshal(e.Err.Error())
	if err != nil {
		return nil, err
	}
	b = append(b, be...)
	b = append(b, '}')
	return b, nil
}

var ErrNotFound = Error{
	Code: http.StatusNotFound,
	Err:  errors.New(http.StatusText(http.StatusNotFound)),
}

var ErrMethodNotAllowed = Error{
	Code: http.StatusMethodNotAllowed,
	Err:  errors.New(http.StatusText(http.StatusMethodNotAllowed)),
}

func DefaultErrorHandler(encoding Encoding[Error]) func(err error) http.Handler {
	return func(err error) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var e Error
			if errors.As(err, &e) {
				_ = encoding.Encode(w, e, e.Code)
				return
			}
			_ = encoding.Encode(w, Error{Err: err}, http.StatusInternalServerError)
		})
	}
}
