package ghost

import (
	"net/http"

	"github.com/gorilla/schema"
)

type Query any

type Querier[Q Query] interface {
	Query(*http.Request) (Q, error)
}

// QueryParser is a Querier which maps the URL query parameters to a Query.
type QueryParser[Q Query] struct {
	decoder *schema.Decoder
}

func NewQueryParser[Q Query]() QueryParser[Q] {
	return QueryParser[Q]{
		decoder: schema.NewDecoder(),
	}
}

func (qp QueryParser[Q]) Query(r *http.Request) (Q, error) {
	var q Q
	err := qp.decoder.Decode(&q, r.URL.Query())
	return q, err
}
