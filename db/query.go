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

func (db *DB) runFilter(f Filter) error {
	var (
		subjectIsVariable = strings.HasPrefix(f.Subject.Value, "?")
		objectIsVariable  = strings.HasPrefix(f.Object.Value, "?")
	)
	fmt.Println(f)
	// if the subject is a variable, then we need to anchor to something else,
	// so we skip this part. If the subject is *not* a variable, then we pull all
	// the triples it starts in (or create a function to do this)
	if !subjectIsVariable {
		entity, err := db.GetEntity(f.Subject)
		if err != nil {
			return err
		}
		fmt.Printf("%+v\n", entity)
	}
	if !objectIsVariable {
		entity, err := db.GetEntity(f.Object)
		if err != nil {
			return err
		}
		fmt.Printf("%+v\n", entity)
		results := db.followPredicateChainFromEnd(entity.PK, f.Path)
		for _, res := range results {
			uri, err := db.GetURI(res)
			if err != nil {
				return err
			}
			fmt.Println("=>", uri)
		}
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
	for _, pattern := range path {
		pe, found := db.predIndex[pattern.Predicate]
		if !found {
			panic(fmt.Sprintf("Cannot find predicate %s", pattern.Predicate))
		}
		// want to find all subjects that have pattern.Predicate relationship to us
		for idx, obj := range pe.Objects {
			if obj == entityHash {
				subjectHashes = append(subjectHashes, pe.Subjects[idx])
			}
		}
	}
	return subjectHashes
}
