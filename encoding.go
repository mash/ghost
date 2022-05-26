package ghost

import "net/http"

type Encoding[R Resource, Q Query] interface {
	Encode(http.ResponseWriter, R) error
	EncodeList(http.ResponseWriter, []R) error
	EncodeEmpty(http.ResponseWriter) error
	Decode(*http.Request) (R, error)
}
