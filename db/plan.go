package db

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/btree"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
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
			subjectIsVariable  = strings.HasPrefix(term.Subject.Value, "?")
			objectIsVariable   = strings.HasPrefix(term.Object.Value, "?")
			subjectVar         = term.Subject.String()
			objectVar          = term.Object.String()
			hasResolvedSubject bool
			hasResolvedObject  bool
			newop              operation
		)
		hasResolvedSubject = qp.varOrder.hasVar(subjectVar)
		hasResolvedObject = qp.varOrder.hasVar(objectVar)

		switch {
		case subjectIsVariable && objectIsVariable:
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
				log.Warningf("%s depends on %s", rso.childVar, rso.parentVar)
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

type operation interface {
	run(db *DB, varOrder *variableStateMap, rm *resultMap) (*resultMap, error)
	String() string
	SortKey() string
	GetTerm() *queryTerm
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

// ?subject predicate object
// i.e. subject is the var
type resolveSubject struct {
	term *queryTerm
}

func (rs *resolveSubject) String() string {
	return fmt.Sprintf("[resolveSubject %s]", rs.term)
}

func (rs *resolveSubject) SortKey() string {
	return rs.term.Subject.String()
}

func (rs *resolveSubject) GetTerm() *queryTerm {
	return rs.term
}

func (rs *resolveSubject) run(db *DB, varOrder *variableStateMap, rm *resultMap) (*resultMap, error) {
	// fetch the object from the graph
	object, err := db.GetEntity(rs.term.Object)
	if err != nil && err != leveldb.ErrNotFound {
		return rm, errors.Wrap(err, fmt.Sprintf("%+v", rs.term))
	} else if err == leveldb.ErrNotFound {
		return rm, nil
	}
	subjectVar := rs.term.Subject.String()
	// get all subjects reachable from the given object along the path
	subjects := db.getSubjectFromPredObject(object.PK, rs.term.Path)

	// need to restrict if we are child. Else, just add definition
	if varOrder.varIsChild(subjectVar) {
		entSubjects := hashTreeToEntityTree(subjects)
		for _, subject := range rm.iterVariable(subjectVar) {
			if !entSubjects.Has(subject) {
				subject.PK = emptyHash
			}
		}
	} else {
		rm.addVariable(subjectVar, subjects)
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

func (ro *resolveObject) SortKey() string {
	return ro.term.Object.String()
}

func (ro *resolveObject) GetTerm() *queryTerm {
	return ro.term
}

func (ro *resolveObject) run(db *DB, varOrder *variableStateMap, rm *resultMap) (*resultMap, error) {
	// fetch the subject from the graph
	subject, err := db.GetEntity(ro.term.Subject)
	if err != nil && err != leveldb.ErrNotFound {
		return rm, errors.Wrap(err, fmt.Sprintf("%+v", ro.term))
	} else if err == leveldb.ErrNotFound {
		return rm, nil
	}
	objectVar := ro.term.Object.String()
	// get all objects reachable from the given subject along the path
	objects := db.getObjectFromSubjectPred(subject.PK, ro.term.Path)
	rm.addVariable(objectVar, objects)
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

func (rso *restrictSubjectObjectByPredicate) SortKey() string {
	return rso.parentVar
}

func (rso *restrictSubjectObjectByPredicate) GetTerm() *queryTerm {
	return rso.term
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
	var (
		subjectVar = rso.term.Subject.String()
		objectVar  = rso.term.Object.String()
		subTree    = rm.getVar(subjectVar)
		objTree    = rm.getVar(objectVar)
	)
	// we add the objects on to each subject
	if rso.parentVar == subjectVar {
		// iterate through current subjects
		for _, subject := range rm.iterVariable(subjectVar) {
			objects := hashTreeToEntityTree(db.getObjectFromSubjectPred(subject.PK, rso.term.Path))
			if objects.Len() > 0 {
				if objTree == nil {
					subject.Next[objectVar] = objects
				} else {
					subject.Next[objectVar] = intersectTrees(objects, objTree)
				}
			}
			//if len(subject.Next) > 0 {
			//	rm.replaceEntity(subjectVar, subject)
			//}
		}
	} else if rso.parentVar == objectVar {
		for _, object := range rm.iterVariable(objectVar) {
			subjects := hashTreeToEntityTree(db.getSubjectFromPredObject(object.PK, rso.term.Path))
			if subjects.Len() > 0 {
				if subTree == nil {
					object.Next[subjectVar] = subjects
				} else {
					object.Next[subjectVar] = intersectTrees(subjects, subTree)
				}
			}
			//if len(object.Next) > 0 {
			//	rm.replaceEntity(objectVar, object)
			//}
		}
	} else {
		log.Fatal("unfamiliar situation")
	}

	return rm, nil
}

type resolveSubjectFromVarObject struct {
	term *queryTerm
}

func (rsv *resolveSubjectFromVarObject) String() string {
	return fmt.Sprintf("[resolveSubFromVarObj %s]", rsv.term)
}

func (rsv *resolveSubjectFromVarObject) SortKey() string {
	return rsv.term.Object.String()
}

func (rsv *resolveSubjectFromVarObject) GetTerm() *queryTerm {
	return rsv.term
}

// Use this when we have subject and object variables, but only object has been filled in
func (rsv *resolveSubjectFromVarObject) run(db *DB, varOrder *variableStateMap, rm *resultMap) (*resultMap, error) {
	var (
		objectVar  = rsv.term.Object.String()
		subjectVar = rsv.term.Subject.String()
	)
	for _, object := range rm.iterVariable(objectVar) {
		subjects := hashTreeToEntityTree(db.getSubjectFromPredObject(object.PK, rsv.term.Path))
		if _, found := object.Next[subjectVar]; found {
			mergeTrees(object.Next[subjectVar], subjects)
		} else {
			object.Next[subjectVar] = subjects
		}
	}
	return rm, nil
}

type resolveObjectFromVarSubject struct {
	term *queryTerm
}

func (rov *resolveObjectFromVarSubject) String() string {
	return fmt.Sprintf("[resolveObjFromVarSub %s]", rov.term)
}

func (rov *resolveObjectFromVarSubject) SortKey() string {
	return rov.term.Subject.String()
}

func (rov *resolveObjectFromVarSubject) GetTerm() *queryTerm {
	return rov.term
}

func (rov *resolveObjectFromVarSubject) run(db *DB, varOrder *variableStateMap, rm *resultMap) (*resultMap, error) {
	var (
		objectVar  = rov.term.Object.String()
		subjectVar = rov.term.Subject.String()
	)
	for _, subject := range rm.iterVariable(subjectVar) {
		objects := hashTreeToEntityTree(db.getObjectFromSubjectPred(subject.PK, rov.term.Path))
		if _, found := subject.Next[objectVar]; found {
			mergeTrees(subject.Next[objectVar], objects)
		} else {
			subject.Next[objectVar] = objects
		}
	}
	return rm, nil
}

type resolveSubjectObjectFromPred struct {
	term *queryTerm
}

func (rso *resolveSubjectObjectFromPred) run(db *DB, varOrder *variableStateMap, rm *resultMap) (*resultMap, error) {
	subsobjs := db.getSubjectObjectFromPred(rso.term.Path)
	subjectVar := rso.term.Subject.String()
	objectVar := rso.term.Object.String()
	subjects := btree.New(3)
	objects := btree.New(3)
	for _, sopair := range subsobjs {
		subjects.ReplaceOrInsert(Item(sopair[0]))
		objects.ReplaceOrInsert(Item(sopair[1]))
	}
	rm.addVariable(subjectVar, subjects)
	rm.addVariable(objectVar, objects)
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
