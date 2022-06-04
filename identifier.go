package ghost

import (
	"net/http"
	"path"
	"strconv"
)

type PKey interface {
	comparable
}

type PIntKey interface {
	comparable
	~int | ~int32 | ~int64 | ~uint | ~uint32 | ~uint64
}

type Identifier[P PKey] interface {
	PKey(*http.Request) (P, error)
}

// PathIdentifier is an Identifier which extracts the PKey from the request URL path.
type PathIdentifier[P PKey] func(string) (P, error)

func (pi PathIdentifier[P]) PKey(r *http.Request) (P, error) {
	var p P
	_, lastpath := path.Split(r.URL.Path)
	if lastpath != "" {
		return pi(lastpath)
	}
	return p, ErrNotFound
}

func IntPath[P PIntKey](s string) (P, error) {
	i, err := strconv.ParseInt(s, 10, 64)
	return P(i), err
}
