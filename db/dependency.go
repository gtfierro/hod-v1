package db

import (
	"fmt"
	sparql "github.com/gtfierro/hod/lang/ast"
	//"reflect"
	"strings"
)

const (
	varRESOLVED   = "RESOLVED"
	varUNRESOLVED = ""
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
	dg.selectVars = append(dg.selectVars, q.Select.Vars...)
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
		var idx int
		var term *queryTerm
		for idx, term = range dg.terms {
			if term.overlap(next) > 0 {
				break
			}
		}
		next = term
		dg.plan = append(dg.plan, *next)
		dg.terms = append(dg.terms[:idx], dg.terms[idx+1:]...)

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
//func (qt *queryTerm) equals(qt2 *queryTerm) bool {
//	return qt.Subject == qt2.Subject &&
//		qt.Object == qt2.Object &&
//		reflect.DeepEqual(qt.Predicates, qt2.Predicates)
//}

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

// need operator types that go into the query plan
// Types:
//  SELECT: given a 2/3 triple, it resolves the 3rd item
//  FILTER: given a 1/3 triple, it restricts the other 2 items

// the old "queryplan" file is really a dependency graph for the query: it is NOT
// the queryplanner. What we should do now is take that dependency graph and turn
// it into a query plan

func formQueryPlan(dg *dependencyGraph, q *sparql.Query) (*queryPlan, error) {
	plan := newQueryPlan(dg, q)

	for _, term := range dg.plan {
		var (
			subjectIsVariable = term.Subject.IsVariable()
			objectIsVariable  = term.Object.IsVariable()
			// for now just look at first item in path
			predicateIsVariable  = term.Predicates[0].Predicate.IsVariable()
			subjectVar           = term.Subject.String()
			objectVar            = term.Object.String()
			predicateVar         = term.Predicates[0].Predicate.String()
			hasResolvedSubject   bool
			hasResolvedObject    bool
			hasResolvedPredicate bool
			newop                operation
			numvars              = len(term.variables)
		)
		hasResolvedSubject = plan.hasVar(subjectVar)
		hasResolvedObject = plan.hasVar(objectVar)
		hasResolvedPredicate = plan.hasVar(predicateVar)

		switch {
		// definitions: do these first
		case numvars == 1 && subjectIsVariable:
			newop = &resolveSubject{term: term}
			if !plan.varIsChild(subjectVar) {
				plan.addTopLevel(subjectVar)
			}
		case numvars == 1 && objectIsVariable:
			// s p ?o
			newop = &resolveObject{term: term}
			if !plan.varIsChild(objectVar) {
				plan.addTopLevel(objectVar)
			}
		case numvars == 1 && predicateIsVariable:
			// s ?p o
			newop = &resolvePredicate{term: term}
			if !plan.varIsChild(predicateVar) {
				plan.addTopLevel(predicateVar)
			}
		// terms with 3 variables
		case subjectIsVariable && objectIsVariable && predicateIsVariable:
			switch {
			case hasResolvedSubject:
				newop = &resolveVarTripleFromSubject{term: term}
			case hasResolvedObject:
				newop = &resolveVarTripleFromObject{term: term}
			case hasResolvedPredicate:
				newop = &resolveVarTripleFromPredicate{term: term}
			default: // all are vars
				newop = &resolveVarTripleAll{term: term}
			}
		// subject/object variable terms
		case subjectIsVariable && objectIsVariable && !predicateIsVariable:
			switch {
			case hasResolvedSubject && hasResolvedObject:
				// if we have both subject and object, we filter
				rso := &restrictSubjectObjectByPredicate{term: term}
				if plan.varIsChild(subjectVar) {
					plan.addLink(subjectVar, objectVar)
					rso.parentVar = subjectVar
					rso.childVar = objectVar
				} else if plan.varIsChild(objectVar) {
					plan.addLink(objectVar, subjectVar)
					rso.parentVar = objectVar
					rso.childVar = subjectVar
				} else if plan.varIsTop(subjectVar) {
					plan.addLink(subjectVar, objectVar)
					rso.parentVar = subjectVar
					rso.childVar = objectVar
				} else if plan.varIsTop(objectVar) {
					plan.addLink(objectVar, subjectVar)
					rso.parentVar = objectVar
					rso.childVar = subjectVar
				}
				newop = rso
			case hasResolvedObject:
				newop = &resolveSubjectFromVarObject{term: term}
				plan.addLink(objectVar, subjectVar)
			case hasResolvedSubject:
				newop = &resolveObjectFromVarSubject{term: term}
				plan.addLink(subjectVar, objectVar)
			default:
				newop = &resolveSubjectObjectFromPred{term: term}
				plan.addLink(subjectVar, objectVar)
			}
		case !subjectIsVariable && !objectIsVariable && predicateIsVariable:
			newop = &resolvePredicate{term: term}
			if !plan.varIsChild(predicateVar) {
				plan.addTopLevel(predicateVar)
			}
		case subjectIsVariable && !objectIsVariable && predicateIsVariable:
			// ?s ?p o
			newop = &resolveSubjectPredFromObject{term: term}
			plan.addLink(subjectVar, predicateVar)
		case !subjectIsVariable && objectIsVariable && predicateIsVariable:
			// s ?p ?o
			newop = &resolvePredObjectFromSubject{term: term}
			plan.addLink(objectVar, predicateVar)
		case subjectIsVariable:
			// ?s p o
			newop = &resolveSubject{term: term}
			if !plan.varIsChild(subjectVar) {
				plan.addTopLevel(subjectVar)
			}
		case objectIsVariable:
			// s p ?o
			newop = &resolveObject{term: term}
			if !plan.varIsChild(objectVar) {
				plan.addTopLevel(objectVar)
			}
		default:
			return plan, fmt.Errorf("Nothing chosen for %s. This shouldn't happen", term)
		}
		plan.operations = append(plan.operations, newop)
	}
	return plan, nil
}

// contains all useful state information for executing a query
type queryPlan struct {
	operations []operation
	selectVars []string
	dg         *dependencyGraph
	query      *sparql.Query
	vars       map[string]string
}

func newQueryPlan(dg *dependencyGraph, q *sparql.Query) *queryPlan {
	plan := &queryPlan{
		selectVars: dg.selectVars,
		dg:         dg,
		query:      q,
		vars:       make(map[string]string),
	}
	return plan
}

//func (plan *queryPlan) dumpVarchain() {
//	for k, v := range plan.vars {
//		fmt.Println(k, "=>", v)
//	}
//}

func (plan *queryPlan) hasVar(variable string) bool {
	return plan.vars[variable] != varUNRESOLVED
}

func (plan *queryPlan) varIsChild(variable string) bool {
	return plan.hasVar(variable) && plan.vars[variable] != varRESOLVED
}

func (plan *queryPlan) varIsTop(variable string) bool {
	return plan.hasVar(variable) && plan.vars[variable] == varRESOLVED
}

func (plan *queryPlan) addTopLevel(variable string) {
	plan.vars[variable] = varRESOLVED
}

func (plan *queryPlan) addLink(parent, child string) {
	plan.vars[child] = parent
}
