package ghost

import (
	"net/http"
	"path"
	"strconv"
)

type PKey uint64

type Identifier interface {
	PKeys(*http.Request) ([]PKey, error)
}

// PathIdentifier is an Identifier which extracts the PKey from the request URL path.
type PathIdentifier struct{}

func (PathIdentifier) PKeys(r *http.Request) ([]PKey, error) {
	_, lastpath := path.Split(r.URL.Path)
	if lastpath != "" {
		i, err := strconv.ParseInt(lastpath, 10, 64)
		if err != nil {
			return nil, err
		}
		return []PKey{PKey(i)}, nil
	}
	return nil, errNotFound
}
