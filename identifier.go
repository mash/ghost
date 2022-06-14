package ghost

import (
	"net/http"
	"path"
	"strconv"
)

type PKey interface {
	comparable
}

type PUintKey interface {
	comparable
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type PStrKey interface {
	comparable
	string
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

func UintPath[P PUintKey](s string) (P, error) {
	i, err := strconv.ParseUint(s, 10, 64)
	return P(i), err
}

func StrPath[P PStrKey](s string) (P, error) {
	return P(s), nil
}
