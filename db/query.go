package db

import (
	"fmt"
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

func (db *DB) RunQuery(q Query) {
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
