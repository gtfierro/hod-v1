package db

import (
	"fmt"
	sparql "github.com/gtfierro/hod/lang/ast"
	"reflect"
	"strings"
)

// struct to hold the graph of the query plan
type dependencyGraph struct {
	selectVars []string
	variables  map[string]bool
	terms      []*queryTerm
	plan       []queryTerm
}

func makeDependencyGraph(q *sparql.Query) *dependencyGraph {
	dg := &dependencyGraph{
		selectVars: []string{},
		variables:  make(map[string]bool),
		terms:      make([]*queryTerm, len(q.Where.Terms)),
	}
	for _, v := range q.Select.Vars {
		dg.selectVars = append(dg.selectVars, v)
	}
	for i, term := range q.Where.Terms {
		dg.terms[i] = dg.makeQueryTerm(term)
	}

	// find term with fewest variables
	var next *queryTerm
rootLoop:
	for numvars := 1; numvars <= 3; numvars++ {
		for idx, term := range dg.terms {
			if len(term.variables) == numvars {
				next = term
				dg.terms = append(dg.terms[:idx], dg.terms[idx+1:]...)
				break rootLoop
			}
		}
	}
	dg.plan = append(dg.plan, *next)

	for len(dg.terms) > 0 {
		for idx, term := range dg.terms {
			if term.overlap(next) > 0 {
				next = term
				dg.plan = append(dg.plan, *next)
				dg.terms = append(dg.terms[:idx], dg.terms[idx+1:]...)
				break
			}
		}
	}
	return dg
}

func (dg *dependencyGraph) dump() {
	for _, r := range dg.terms {
		r.dump(0)
		//fmt.Println(r)
	}
}

// stores the state/variables for a particular triple
// from a SPARQL query
type queryTerm struct {
	sparql.Triple
	dependencies []*queryTerm
	variables    []string
}

// initializes a queryTerm from a given Filter
func (dg *dependencyGraph) makeQueryTerm(t sparql.Triple) *queryTerm {
	qt := &queryTerm{
		t,
		[]*queryTerm{},
		[]string{},
	}
	if qt.Subject.IsVariable() {
		dg.variables[qt.Subject.String()] = false
		qt.variables = append(qt.variables, qt.Subject.String())
	}
	if qt.Predicates[0].Predicate.IsVariable() {
		dg.variables[qt.Predicates[0].Predicate.String()] = false
		qt.variables = append(qt.variables, qt.Predicates[0].Predicate.String())
	}
	if qt.Object.IsVariable() {
		dg.variables[qt.Object.String()] = false
		qt.variables = append(qt.variables, qt.Object.String())
	}
	return qt
}

// returns true if two query terms are equal
func (qt *queryTerm) equals(qt2 *queryTerm) bool {
	return qt.Subject == qt2.Subject &&
		qt.Object == qt2.Object &&
		reflect.DeepEqual(qt.Predicates, qt2.Predicates)
}

func (qt *queryTerm) String() string {
	return fmt.Sprintf("<%s %s %s>", qt.Subject, qt.Predicates, qt.Object)
}

func (qt *queryTerm) dump(indent int) {
	fmt.Println(strings.Repeat("  ", indent), qt.String())
	for _, c := range qt.dependencies {
		c.dump(indent + 1)
	}
}

func (qt *queryTerm) overlap(other *queryTerm) int {
	count := 0
	for _, v := range qt.variables {
		for _, vv := range other.variables {
			if vv == v {
				count++
			}
		}
	}
	return count
}

type queryTermList []*queryTerm

func (list queryTermList) Len() int {
	return len(list)
}
func (list queryTermList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}
func (list queryTermList) Less(i, j int) bool {
	if len(list[i].variables) == 1 {
		return true
	} else if len(list[j].variables) == 1 {
		return false
	}
	i_overlap := 0
	for idx := 0; idx < i; idx++ {
		if idx == j {
			continue
		}
		i_overlap += list[i].overlap(list[idx])
	}
	j_overlap := 0
	for idx := 0; idx < j; idx++ {
		if idx == i {
			continue
		}
		j_overlap += list[j].overlap(list[idx])
	}
	return i_overlap > j_overlap

}
