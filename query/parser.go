package query

import (
	turtle "github.com/gtfierro/hod/goraptor"
	"io"
)

type Query struct {
	Select SelectClause
	Where  WhereClause
}

func (q Query) Copy() Query {
	newq := Query{
		Select: q.Select,
		Where: WhereClause{
			Filters: make([]Filter, len(q.Where.Filters)),
		},
	}
	for i, v := range q.Where.Filters {
		newq.Where.Filters[i] = v.Copy()
	}
	copy(newq.Where.Ors, q.Where.Ors)
	return newq
}

type SelectClause struct {
	Variables []SelectVar
	Distinct  bool
	Count     bool
	Partial   bool
	HasLinks  bool
	Limit     int
}

type SelectVar struct {
	Var      turtle.URI
	AllLinks bool
	Links    []Link
}

func (v SelectVar) Copy() SelectVar {
	sv := SelectVar{
		Var:      v.Var,
		AllLinks: v.AllLinks,
	}
	copy(sv.Links, v.Links)
	return sv
}

type Link struct {
	Name string
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

func (f Filter) Copy() Filter {
	ff := Filter{
		Subject: f.Subject,
		Object:  f.Object,
		Path:    make([]PathPattern, len(f.Path)),
	}
	for i, p := range f.Path {
		ff.Path[i] = p
	}
	return ff
}

func (f Filter) Equals(f2 Filter) bool {
	return f.Subject == f2.Subject &&
		f.Object == f2.Object &&
		comparePathSliceAsSet(f.Path, f2.Path)
}

func (f Filter) NumVars() int {
	num := 0
	if f.Subject.IsVariable() {
		num++
	}
	if f.Path[0].Predicate.IsVariable() {
		num++
	}
	if f.Object.IsVariable() {
		num++
	}
	return num
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

// returns a flat list of triples that comes from expanding out the OrClause tree
func (oc OrClause) Flatten() [][]Filter {
	var res [][]Filter
	// we split on left/right
	leftbase := append(oc.Terms, oc.LeftTerms...)
	if len(oc.LeftOr) == 0 {
		res = append(res, leftbase)
	} else {
		for _, loc := range oc.LeftOr {
			for _, termBlock := range loc.Flatten() {
				res = append(res, append(leftbase, termBlock...))
			}
		}
	}
	rightbase := append(oc.Terms, oc.RightTerms...)
	if len(oc.RightOr) == 0 {
		res = append(res, rightbase)
	} else {
		for _, roc := range oc.RightOr {
			for _, termBlock := range roc.Flatten() {
				res = append(res, append(rightbase, termBlock...))
			}
		}
	}
	return res
}

func FlattenOrClauseList(oclist []OrClause) [][]Filter {
	var allOrTerms [][]Filter
	for _, orclause := range oclist {
		if len(allOrTerms) == 0 {
			allOrTerms = append(allOrTerms, orclause.Flatten()...)
		} else {
			var newAllOrTerms [][]Filter
			for _, termblock := range orclause.Flatten() {
				for _, prefixTerm := range allOrTerms {
					newAllOrTerms = append(newAllOrTerms, append(termblock, prefixTerm...))
				}
			}
			allOrTerms = newAllOrTerms
		}
	}
	// eliminate duplicates
	if len(allOrTerms) > 0 {
		var keepTerms [][]Filter
	orterm:
		for _, flist := range allOrTerms {
			for _, kt := range keepTerms {
				if compareFilterSliceAsSet(flist, kt) {
					continue orterm
				}
			}
			keepTerms = append(keepTerms, flist)
		}
		allOrTerms = keepTerms
	}
	return allOrTerms
}

func FilterListToOrClause(filters []Filter) OrClause {
	orc := OrClause{}
	if len(filters) == 1 {
		orc.Terms = filters
		return orc
	}
	orc.LeftTerms = []Filter{filters[0]}
	orc.RightOr = []OrClause{FilterListToOrClause(filters[1:])}
	return orc
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
	defer lexerpool.Put(l)
	yyParse(l)
	if l.error != nil {
		return Query{}, l.error
	}
	q := Query{}
	q.Select = SelectClause{Variables: l.varlist, Distinct: l.distinct, Count: l.count, Partial: l.partial, Limit: int(l.limit)}
	for _, selectvar := range l.varlist {
		if len(selectvar.Links) > 0 || selectvar.AllLinks {
			q.Select.HasLinks = true
			break
		}
	}
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
