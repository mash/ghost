package ghost

import "net/http"

type Query any

type Querier[Q Query] interface {
	Query(*http.Request) (Q, error)
}
