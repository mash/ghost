package ghost

import "net/http"

type PKey uint64

type PKeyer interface {
	PKeys(*http.Request) ([]PKey, error)
}
