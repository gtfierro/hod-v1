package lang

import (
	turtle "github.com/gtfierro/hod/goraptor"
	"io"
)

type Query struct {
	Select SelectClause
	Where  WhereClause
}

type SelectClause struct {
	Variables []turtle.URI
	Distinct  bool
	Count     bool
}

type WhereClause struct {
	Filters []Filter
	Ors     []OrClause
}

type Filter struct {
	Subject turtle.URI
	Path    []PathPattern
	Object  turtle.URI
}

type OrClause struct {
	// a component of an OR clause
	Terms []Filter
	// pointer to the left/right of OR clause
	// These are nestable
	LeftOr     []OrClause
	LeftTerms  []Filter
	RightOr    []OrClause
	RightTerms []Filter
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

func (p Pattern) String() string {
	switch p {
	case PATTERN_SINGLE:
		return ""
	case PATTERN_ZERO_ONE:
		return "?"
	case PATTERN_ONE_PLUS:
		return "+"
	case PATTERN_ZERO_PLUS:
		return "*"
	}
	return "unknown"
}

func Parse(r io.Reader) (Query, error) {
	l := newlexer(r)
	yyParse(l)
	if l.error != nil {
		return Query{}, l.error
	}
	q := Query{}
	q.Select = SelectClause{Variables: l.varlist, Distinct: l.distinct, Count: l.count}
	q.Where = WhereClause{
		Filters: []Filter{},
		Ors:     []OrClause{},
	}
	if len(l.triples) > 0 {
		q.Where.Filters = l.triples
	}
	if len(l.orclauses) > 0 {
		q.Where.Ors = l.orclauses
	}

	return q, nil
}
