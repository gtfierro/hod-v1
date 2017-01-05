package db

import (
	"sort"
)

// need operator types that go into the query plan
// Types:
//  SELECT: given a 2/3 triple, it resolves the 3rd item
//  FILTER: given a 1/3 triple, it restricts the other 2 items

// the old "queryplan" file is really a dependency graph for the query: it is NOT
// the queryplanner. What we should do now is take that dependency graph and turn
// it into a query plan

type queryPlan struct {
	operations []operation
	selectVars []string
	varOrder   *variableStateMap
	dg         *dependencyGraph
}

func (qp *queryPlan) findVarDepth(target string) int {
	var depth = 0
	start := qp.varOrder.vars[target]
	for start != RESOLVED {
		start = qp.varOrder.vars[start]
		depth += 1
	}
	return depth
}

func (db *DB) formQueryPlan(dg *dependencyGraph) *queryPlan {
	qp := new(queryPlan)
	qp.dg = dg
	qp.selectVars = dg.selectVars
	qp.varOrder = newVariableStateMap()

	for term := range dg.iter() {
		var (
			subjectIsVariable = term.Subject.IsVariable()
			objectIsVariable  = term.Object.IsVariable()
			// for now just look at first item in path
			predicateIsVariable  = term.Path[0].Predicate.IsVariable()
			subjectVar           = term.Subject.String()
			objectVar            = term.Object.String()
			predicateVar         = term.Path[0].Predicate.String()
			hasResolvedSubject   bool
			hasResolvedObject    bool
			hasResolvedPredicate bool
			newop                operation
		)
		hasResolvedSubject = qp.varOrder.hasVar(subjectVar)
		hasResolvedObject = qp.varOrder.hasVar(objectVar)
		hasResolvedPredicate = qp.varOrder.hasVar(predicateVar)

		switch {
		case subjectIsVariable && objectIsVariable && predicateIsVariable:
			// Cases:
			// NONE resolved: enumerate all triples in the store
			// subject, pred resolved:
			// object, pred resolved:
			// subject, object resolved:
			// subject resolved:
			// object resolved:
			// pred resolved:
			switch {
			case !hasResolvedSubject && !hasResolvedObject && !hasResolvedPredicate:
				log.Fatal("?x ?y ?z queries not supported yet")
			case hasResolvedSubject && !hasResolvedObject && hasResolvedPredicate:

			}
		case subjectIsVariable && objectIsVariable && !predicateIsVariable:
			switch {
			case hasResolvedSubject && hasResolvedObject:
				// if we have both subject and object, we filter
				rso := &restrictSubjectObjectByPredicate{term: term}
				subDepth := qp.findVarDepth(subjectVar)
				objDepth := qp.findVarDepth(objectVar)
				if subDepth > objDepth {
					qp.varOrder.addLink(subjectVar, objectVar)
					rso.parentVar = subjectVar
					rso.childVar = objectVar
				} else if objDepth > subDepth {
					qp.varOrder.addLink(objectVar, subjectVar)
					rso.parentVar = objectVar
					rso.childVar = subjectVar
				} else if qp.varOrder.varIsChild(subjectVar) {
					qp.varOrder.addLink(subjectVar, objectVar)
					rso.parentVar = subjectVar
					rso.childVar = objectVar
				} else if qp.varOrder.varIsChild(objectVar) {
					qp.varOrder.addLink(objectVar, subjectVar)
					rso.parentVar = objectVar
					rso.childVar = subjectVar
				} else if qp.varOrder.varIsTop(subjectVar) {
					qp.varOrder.addLink(subjectVar, objectVar)
					rso.parentVar = subjectVar
					rso.childVar = objectVar
				} else if qp.varOrder.varIsTop(objectVar) {
					qp.varOrder.addLink(objectVar, subjectVar)
					rso.parentVar = objectVar
					rso.childVar = subjectVar
				}
				newop = rso
			case hasResolvedObject:
				newop = &resolveSubjectFromVarObject{term: term}
				qp.varOrder.addLink(objectVar, subjectVar)
			case hasResolvedSubject:
				newop = &resolveObjectFromVarSubject{term: term}
				qp.varOrder.addLink(subjectVar, objectVar)
			default:
				panic("HERE")
			}
		case !subjectIsVariable && !objectIsVariable && predicateIsVariable:
			newop = &resolvePredicate{term: term}
			if !qp.varOrder.varIsChild(predicateVar) {
				qp.varOrder.addTopLevel(predicateVar)
			}
			//log.Fatal("x ?y z query not supported yet")
		case subjectIsVariable && !objectIsVariable && predicateIsVariable:
			log.Fatal("?x ?y z query not supported yet")
		case !subjectIsVariable && objectIsVariable && predicateIsVariable:
			log.Fatal("x ?y ?z query not supported yet")
		case subjectIsVariable:
			newop = &resolveSubject{term: term}
			if !qp.varOrder.varIsChild(subjectVar) {
				qp.varOrder.addTopLevel(subjectVar)
			}
		case objectIsVariable:
			newop = &resolveObject{term: term}
			if !qp.varOrder.varIsChild(objectVar) {
				qp.varOrder.addTopLevel(objectVar)
			}
		default:
			log.Fatal("Nothing chosen for", term)
		}
		qp.operations = append(qp.operations, newop)
	}
	// sort operations
	sort.Sort(qp)
	return qp
}

func (qp *queryPlan) Len() int {
	return len(qp.operations)
}
func (qp *queryPlan) Swap(i, j int) {
	qp.operations[i], qp.operations[j] = qp.operations[j], qp.operations[i]
}
func (qp *queryPlan) Less(i, j int) bool {
	iDepth := qp.findVarDepth(qp.operations[i].SortKey())
	jDepth := qp.findVarDepth(qp.operations[j].SortKey())
	return iDepth < jDepth
}

const (
	RESOLVED   = "RESOLVED"
	UNRESOLVED = ""
)

type variableStateMap struct {
	vars map[string]string
}

func newVariableStateMap() *variableStateMap {
	return &variableStateMap{
		vars: make(map[string]string),
	}
}

func (vsm *variableStateMap) hasVar(variable string) bool {
	return vsm.vars[variable] != UNRESOLVED
}

func (vsm *variableStateMap) varIsChild(variable string) bool {
	return vsm.hasVar(variable) && vsm.vars[variable] != RESOLVED
}

func (vsm *variableStateMap) varIsTop(variable string) bool {
	return vsm.hasVar(variable) && vsm.vars[variable] == RESOLVED
}

func (vsm *variableStateMap) addTopLevel(variable string) {
	vsm.vars[variable] = RESOLVED
}

func (vsm *variableStateMap) addLink(parent, child string) {
	vsm.vars[child] = parent
}
