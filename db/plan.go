package db

import (
	"fmt"
	"strings"
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
	varOrder   *variableStateMap
}

func (db *DB) formQueryPlan(dg *dependencyGraph) *queryPlan {
	qp := new(queryPlan)
	statemap := newVariableStateMap()
	for term := range dg.iter() {
		var (
			subjectIsVariable  = strings.HasPrefix(term.Subject.Value, "?")
			objectIsVariable   = strings.HasPrefix(term.Object.Value, "?")
			subjectVar         = term.Subject.String()
			objectVar          = term.Object.String()
			hasResolvedSubject bool
			hasResolvedObject  bool
			newop              operation
		)
		hasResolvedSubject = statemap.hasVar(subjectVar)
		hasResolvedObject = statemap.hasVar(objectVar)

		switch {
		case subjectIsVariable && objectIsVariable:
			switch {
			case hasResolvedSubject && hasResolvedObject:
				// if we have both subject and object, we filter
				rso := &restrictSubjectObjectByPredicate{term: term}
				if statemap.varIsChild(subjectVar) {
					statemap.addLink(subjectVar, objectVar)
					rso.parentVar = subjectVar
					rso.childVar = objectVar
				} else if statemap.varIsChild(objectVar) {
					statemap.addLink(objectVar, subjectVar)
					rso.parentVar = objectVar
					rso.childVar = subjectVar
				} else if statemap.varIsTop(subjectVar) {
					statemap.addLink(subjectVar, objectVar)
					rso.parentVar = subjectVar
					rso.childVar = objectVar
				} else if statemap.varIsTop(objectVar) {
					statemap.addLink(objectVar, subjectVar)
					rso.parentVar = objectVar
					rso.childVar = subjectVar
				}
				newop = rso
			case hasResolvedObject:
				newop = &resolveSubjectFromVarObject{term: term}
				statemap.addLink(objectVar, subjectVar)
			case hasResolvedSubject:
				newop = &resolveObjectFromVarSubject{term: term}
				statemap.addLink(subjectVar, objectVar)
			}
		case subjectIsVariable:
			newop = &resolveSubject{term: term}
			if !statemap.varIsChild(subjectVar) {
				statemap.addTopLevel(subjectVar)
			}
		case objectIsVariable:
			newop = &resolveObject{term: term}
			if !statemap.varIsChild(objectVar) {
				statemap.addTopLevel(objectVar)
			}
		}
		qp.operations = append(qp.operations, newop)
		log.Debug(term)
	}
	for k, v := range statemap.vars {
		log.Debug("statemap", k, "hasparent", v)
	}
	qp.varOrder = statemap
	return qp
}

type operation interface {
	run(db *DB, varOrder *variableStateMap, rm *resultMap) (*resultMap, error)
	String() string
}

// ?subject predicate object
// i.e. subject is the var
type resolveSubject struct {
	term *queryTerm
}

func (rs *resolveSubject) String() string {
	return fmt.Sprintf("[resolveSubject %s]", rs.term)
}

func (rs *resolveSubject) run(db *DB, varOrder *variableStateMap, rm *resultMap) (*resultMap, error) {
	// fetch the object from the graph
	object, err := db.GetEntity(rs.term.Object)
	if err != nil {
		return rm, err
	}
	subjectVar := rs.term.Subject.String()
	// get all subjects reachable from the given object along the path
	subjects := db.getSubjectFromPredObject(object.PK, rs.term.Path)
	// if there is already an entry for the variable in the resultMap, then
	// we intersect the trees
	rm.filterByTree(subjectVar, subjects)
	if varOrder.varIsTop(subjectVar) {
		rm.tuples = subjects
	}
	return rm, nil
}

// subject predicate ?object
type resolveObject struct {
	term *queryTerm
}

func (ro *resolveObject) String() string {
	return fmt.Sprintf("[resolveObject %s]", ro.term)
}

func (ro *resolveObject) run(db *DB, varOrder *variableStateMap, rm *resultMap) (*resultMap, error) {
	// fetch the subject from the graph
	subject, err := db.GetEntity(ro.term.Subject)
	if err != nil {
		return rm, err
	}
	objectVar := ro.term.Object.String()
	// get all objects reachable from the given subject along the path
	objects := db.getObjectFromSubjectPred(subject.PK, ro.term.Path)
	// if there is already an entry for the variable in the resultMap, then
	// we intersect the trees
	rm.filterByTree(objectVar, objects)
	if varOrder.varIsTop(objectVar) {
		rm.tuples = objects
	}
	return rm, nil
}

// ?sub pred ?obj
type restrictSubjectObjectByPredicate struct {
	term                *queryTerm
	parentVar, childVar string
}

func (rso *restrictSubjectObjectByPredicate) String() string {
	return fmt.Sprintf("[restrictSubObjByPred %s]", rso.term)
}

// this forms a linking between the subject and object vars; for each
// subject, we want to have the set of objects that 'follow' from it.
// A variable can be in various states:
//  - unresolved (we don't know what the variable is)
//  - resolved, unconnected (we have proposal values for the variable, but they aren't
//      associated with any other variable)
//  - resolved, connected (we have proposal values for the variable, and they are linked
//      to another variable)
func (rso *restrictSubjectObjectByPredicate) run(db *DB, varOrder *variableStateMap, rm *resultMap) (*resultMap, error) {
	//var (
	//	subjectVar = rso.term.Subject.String()
	//	objectVar  = rso.term.Object.String()
	//	subTree    = rm.getVar(subjectVar)
	//	objTree    = rm.getVar(objectVar)
	//)

	//// we add the objects on to each subject
	//if rso.parentVar == subjectVar {
	//}

	//keepSubjects := btree.New(3)
	//keepObjects := btree.New(3)

	//// we iterate through the subjectVar tree
	//iter := func(i btree.Item) bool {
	//	subject, err := db.GetEntityFromHash(i.(Item))
	//	if err != nil {
	//		log.Error(err)
	//	}
	//	results := db.getObjectFromSubjectPred(subject.PK, rso.term.Path)
	//	if results.Len() > 0 {
	//		keepSubjects.ReplaceOrInsert(i)
	//		mergeTrees(keepObjects, results)
	//	}
	//	return i != subTree.Max()
	//}
	//subTree.Ascend(iter)
	return rm, nil
}

type resolveSubjectFromVarObject struct {
	term *queryTerm
}

func (rsv *resolveSubjectFromVarObject) String() string {
	return fmt.Sprintf("[resolveSubFromVarObj %s]", rsv.term)
}

func (rsv *resolveSubjectFromVarObject) run(db *DB, varOrder *variableStateMap, rm *resultMap) (*resultMap, error) {
	return rm, nil
}

type resolveObjectFromVarSubject struct {
	term *queryTerm
}

func (rov *resolveObjectFromVarSubject) String() string {
	return fmt.Sprintf("[resolveSubFromVarObj %s]", rov.term)
}

func (rov *resolveObjectFromVarSubject) run(db *DB, varOrder *variableStateMap, rm *resultMap) (*resultMap, error) {
	return rm, nil
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
