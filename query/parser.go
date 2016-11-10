package lang

import (
	turtle "github.com/gtfierro/hod/goraptor"
	"io"
)

type Query struct {
	Select SelectClause
	Where  []Filter
}

type SelectClause struct {
	Variables []turtle.URI
}

type Filter struct {
	Subject turtle.URI
	Path    []PathPattern
	Object  turtle.URI
}

type PathPattern struct {
	Predicate turtle.URI
	Pattern   Pattern
}

type Pattern uint

const (
	PATTERN_SINGLE = iota + 1
	PATTERN_ZERO_ONE
	PATTERN_ONE_PLUS
	PATTERN_ZERO_PLUS
)

func Parse(r io.Reader) (Query, error) {
	l := newlexer(r)
	yyParse(l)
	if l.error != nil {
		return Query{}, l.error
	}
	q := Query{}
	q.Select = SelectClause{Variables: l.varlist}
	q.Where = []Filter{}
	for _, filter := range l.triples {
		q.Where = append(q.Where, filter)
	}

	return q, nil
}
