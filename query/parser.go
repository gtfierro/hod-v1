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
}

func Parse(r io.Reader) (Query, error) {
	l := newlexer(r)
	yyParse(l)
	if l.error != nil {
		return Query{}, l.error
	}
	q := Query{}
	q.Select = SelectClause{Variables: l.varlist}
	q.Where = []Filter{}
	for _, triple := range l.triples {
		t := Filter{Subject: triple.Subject, Object: triple.Object}
		pattern := PathPattern{Predicate: triple.Predicate}
		t.Path = []PathPattern{pattern}
		q.Where = append(q.Where, t)
	}

	return q, nil
}
