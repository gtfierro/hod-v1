package db

import (
	"container/list"
	"fmt"
	"reflect"
	"strings"

	turtle "github.com/gtfierro/hod/goraptor"
)

// make a set of structs that capture what these queries want to do

// So, how do queries work?
// We have a list of filters, each of which has a subject, list of predicate things, and
// an object. Any of these might be variables, which we can distinguish by having a "?"
// in front of the value.
//
// First we "clean" these by making sure that they have their full
// namespaces rather than the prefix

type Query struct {
	Select SelectClause
	Where  []Filter
}

type SelectClause struct {
	Variables []string
}

type Filter struct {
	Subject turtle.URI
	Path    []PathPattern
	Object  turtle.URI
}

type PathPattern struct {
	Predicate turtle.URI
}

// struct to hold the graph of the query plan
type queryPlan struct {
	roots []*queryTerm
	// map of variable name -> resolved?
	variables map[string]bool
}

// initializes the query plan struct
func makeQueryPlan() *queryPlan {
	return &queryPlan{
		roots:     []*queryTerm{},
		variables: make(map[string]bool),
	}
}

// returns true of the query plan or any of its children
// already includes the given query term
func (qp *queryPlan) hasChild(qt *queryTerm) bool {
	for _, r := range qp.roots {
		if r.equals(qt) {
			return true
		}
		if r.hasChild(qt) {
			return true
		}
	}
	return false
}

// adds the query term to the root set if it is
// not already there
func (qp *queryPlan) addRootTerm(qt *queryTerm) {
	if !qp.hasChild(qt) {
		qp.roots = append(qp.roots, qt)
	}
}

func (qp *queryPlan) dump() {
	for _, r := range qp.roots {
		r.dump(0)
	}
}

// Firstly, if qt is already in the plan, we return
// iterate through in a breadth first search for any node
// [qt] shares a variable with. We attach qt as a child of that
// term
// Returns true if the node was added
func (qp *queryPlan) addChild(qt *queryTerm) bool {
	if qp.hasChild(qt) {
		fmt.Println("qp already has", qt.String())
		return false
	}
	stack := list.New()
	// push the roots onto the stack
	for _, r := range qp.roots {
		stack.PushFront(r)
	}
	for stack.Len() > 0 {
		node := stack.Remove(stack.Front()).(*queryTerm)
		// if depends on, attach and return
		if qt.dependsOn(node) {
			//fmt.Println("node", qt.String(), "depends on", node.String())
			node.children = append(node.children, qt)
			return true
		}
		// add node children to back of stack
		for _, c := range node.children {
			stack.PushBack(c)
		}
	}
	return false
}

// stores the state/variables for a particular triple
// from a SPARQL query
type queryTerm struct {
	Filter
	children  []*queryTerm
	qp        *queryPlan
	variables []string
}

// initializes a queryTerm from a given Filter
func (qp *queryPlan) makeQueryTerm(f Filter) *queryTerm {
	qt := &queryTerm{
		f,
		[]*queryTerm{},
		qp,
		[]string{},
	}
	// TODO: handle the predicates
	if qt.Subject.IsVariable() {
		qt.qp.variables[qt.Subject.String()] = false
		qt.variables = append(qt.variables, qt.Subject.String())
	}
	if qt.Object.IsVariable() {
		qt.qp.variables[qt.Object.String()] = false
		qt.variables = append(qt.variables, qt.Object.String())
	}
	return qt
}

// returns the number of unresolved variables in the term
func (qt *queryTerm) numUnresolved() int {
	num := 0
	for _, v := range qt.variables {
		if !qt.qp.variables[v] {
			num++
		}
	}
	return num
}

// returns true if the term or any of its children has
// the given child
func (qt *queryTerm) hasChild(child *queryTerm) bool {
	for _, c := range qt.children {
		if c.equals(child) {
			return true
		}
		if c.hasChild(child) {
			return true
		}
	}
	return false
}

// returns true if two query terms are equal
func (qt *queryTerm) equals(qt2 *queryTerm) bool {
	return qt.Subject == qt2.Subject &&
		qt.Object == qt2.Object &&
		reflect.DeepEqual(qt.Path, qt2.Path)
}

func (qt *queryTerm) String() string {
	return fmt.Sprintf("<%s %s %s>", qt.Subject, qt.Path, qt.Object)
}

func (qt *queryTerm) dump(indent int) {
	fmt.Println(strings.Repeat("  ", indent), qt.String())
	for _, c := range qt.children {
		c.dump(indent + 1)
	}
}

func (qt *queryTerm) dependsOn(other *queryTerm) bool {
	for _, v := range qt.variables {
		for _, vv := range other.variables {
			if vv == v {
				return true
			}
		}
	}
	return false
}

// removes all terms in the removeList from removeFrom and returns
// the result
func filterTermList(removeFrom, removeList []*queryTerm) []*queryTerm {
	var ret = []*queryTerm{}
	for _, a := range removeFrom {
		keep := true
		for _, b := range removeList {
			if a.equals(b) {
				keep = false
				break
			}
		}
		if keep {
			ret = append(ret, a)
		}
	}
	return ret
}

func buildQueryPlan(filters []Filter) {
	qp := makeQueryPlan()
	terms := make([]*queryTerm, len(filters))
	for i, f := range filters {
		terms[i] = qp.makeQueryTerm(f)
	}

	// TODO: remember to put this in the loop
	for len(terms) > 0 {
		// first find all the terms with 0 or 1 unresolved variable terms
		var added = []*queryTerm{}
		for _, term := range terms {
			if term.numUnresolved() < 2 {
				qp.addRootTerm(term)
				added = append(added, term)
			}
		}
		// remove the terms that we added to the root set
		terms = filterTermList(terms, added)
		added = []*queryTerm{}
		for _, term := range terms {
			if qp.addChild(term) {
				added = append(added, term)
			}
		}
		terms = filterTermList(terms, added)
	}
	qp.dump()
}

func (db *DB) RunQuery(q Query) {
	// TODO: FINISH build the query plan
	fmt.Println("-------------- start query plan -------------")
	buildQueryPlan(q.Where)
	fmt.Println("-------------- end query plan -------------")

	// "clean" the query by expanding out the prefixes
	for idx, filter := range q.Where {
		if !strings.HasPrefix(filter.Subject.Value, "?") {
			if full, found := db.namespaces[filter.Subject.Namespace]; found {
				filter.Subject.Namespace = full
			}
			q.Where[idx] = filter
		}
		if !strings.HasPrefix(filter.Object.Value, "?") {
			if full, found := db.namespaces[filter.Object.Namespace]; found {
				filter.Object.Namespace = full
			}
			q.Where[idx] = filter
		}
		for idx2, pred := range filter.Path {
			if !strings.HasPrefix(pred.Predicate.Value, "?") {
				if full, found := db.namespaces[pred.Predicate.Namespace]; found {
					pred.Predicate.Namespace = full
				}
				filter.Path[idx2] = pred
			}
		}
		q.Where[idx] = filter
	}

	for _, filter := range q.Where {
		db.runFilter(filter)
	}
}

// We need an execution plan for the list of filters contained in a query. How do we do this?
func (db *DB) formExecutionPlan(list []Filter) {
}

func (db *DB) runFilter(f Filter) error {
	var (
		subjectIsVariable = strings.HasPrefix(f.Subject.Value, "?")
		objectIsVariable  = strings.HasPrefix(f.Object.Value, "?")
	)
	// right now this only handles the first path predicate
	if !subjectIsVariable && !objectIsVariable {
		log.Noticef("S/O anchored: S: %s, O: %s", f.Subject.String(), f.Object.String())
		results := db.getSubjectObjectFromPred(f.Path[0])
		log.Infof("Got %d results", len(results))
	} else if !subjectIsVariable {
		log.Noticef("S anchored: S: %s, O: %s", f.Subject.String(), f.Object.String())
	} else if !objectIsVariable {
		log.Noticef("O anchored: S: %s, O: %s", f.Subject.String(), f.Object.String())
		entity, err := db.GetEntity(f.Object)
		if err != nil {
			return err
		}
		results := db.getSubjectFromPredObject(entity.PK, f.Path[0])
		log.Infof("Got %d results", len(results))
	} else {
		log.Noticef("not anchored!: S: %s, O: %s", f.Subject.String(), f.Object.String())
	}
	return nil
}

//// takes the inverse of every relationship. If no inverse exists, returns nil
func (db *DB) reversePathPattern(path []PathPattern) []PathPattern {
	var reverse = make([]PathPattern, len(path))
	for idx, pred := range path {
		if inverse, found := db.relationships[pred.Predicate]; found {
			pred.Predicate = inverse
			reverse[idx] = pred
		} else {
			return nil
		}
	}
	return reverse
}

// retrieve set of entities reachable starting from the given entity as the 'object'
// what if entity is a variable?
func (db *DB) followPredicateChainFromEnd(entityHash [4]byte, path []PathPattern) [][4]byte {
	// we first check if we have a reversible path
	reversePath := db.reversePathPattern(path)
	if reversePath != nil {
		// begin traversal
		log.Debug("found reverse path", reversePath)
	}
	log.Debug("no reverse path found!")
	// now we consult the predicate index
	var subjectHashes [][4]byte
	// TODO: this is wrong because it doesn't traverse from the last results of subjects
	// TODO NEXT: follow a path of predicates
	for _, pattern := range path {
		subjectHashes = append(subjectHashes, db.getSubjectFromPredObject(entityHash, pattern)...)
	}
	return subjectHashes
}

// Given object and predicate, get all subjects
func (db *DB) getSubjectFromPredObject(objectHash [4]byte, pattern PathPattern) [][4]byte {
	// get the object, look in its "in" edges for the path pattern
	objEntity, err := db.GetEntityFromHash(objectHash)
	if err != nil {
		panic(err)
	}
	// get predicate hash
	predHash, err := db.GetHash(pattern.Predicate)
	if err != nil {
		panic(err)
	}
	return objEntity.InEdges[string(predHash[:])]
}

// Given object and predicate, get all subjects
func (db *DB) getObjectFromSubjectPred(subjectHash [4]byte, pattern PathPattern) [][4]byte {
	// get the object, look in its "out" edges for the path pattern
	subEntity, err := db.GetEntityFromHash(subjectHash)
	if err != nil {
		panic(err)
	}
	// get predicate hash
	predHash, err := db.GetHash(pattern.Predicate)
	if err != nil {
		panic(err)
	}
	return subEntity.InEdges[string(predHash[:])]
}

// Given a predicate, it returns pairs of (subject, object) that are connected by that relationship
func (db *DB) getSubjectObjectFromPred(pattern PathPattern) (soPair [][][4]byte) {
	pe, found := db.predIndex[pattern.Predicate]
	if !found {
		panic(fmt.Sprintf("Cannot find predicate %s", pattern.Predicate))
	}
	for subject, objectMap := range pe.Subjects {
		for object := range objectMap {
			var sh, oh [4]byte
			copy(sh[:], subject)
			copy(oh[:], object)
			soPair = append(soPair, [][4]byte{sh, oh})
		}
	}
	return soPair
}

// Given subject and predicate, get all objects
