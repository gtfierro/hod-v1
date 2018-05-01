package db

import (
	"container/list"
	"fmt"

	sparql "github.com/gtfierro/hod/lang/ast"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

// takes the inverse of every relationship. If no inverse exists, returns nil
func (db *DB) reversePathPattern(path []sparql.PathPattern) []sparql.PathPattern {
	var reverse = make([]sparql.PathPattern, len(path))
	for idx, pred := range path {
		if inverse, found := db.relationships[pred.Predicate]; found {
			pred.Predicate = inverse
			reverse[idx] = pred
		} else {
			return nil
		}
	}
	return reversePath(reverse)
}

// follow the pattern from the given object's InEdges, placing the results in the btree
func (db *DB) followPathFromObject(object *Entity, results *keyTree, searchstack *list.List, pattern sparql.PathPattern) {
	stack := list.New()
	stack.PushFront(object)

	predHash, err := db.GetHash(pattern.Predicate)
	if err != nil && err == leveldb.ErrNotFound {
		log.Infof("Adding unseen predicate %s", pattern.Predicate)
		var hashdest Key
		if err := db.insertEntity(pattern.Predicate, hashdest[:]); err != nil {
			panic(fmt.Errorf("Could not insert entity %s (%v)", pattern.Predicate, err))
		}
	} else if err != nil {
		log.Error(errors.Wrapf(err, "Not found: %v", pattern.Predicate))
		return
	}

	var traversed = traversedBTreePool.Get()
	defer traversedBTreePool.Put(traversed)

	for stack.Len() > 0 {
		entity := stack.Remove(stack.Front()).(*Entity)
		if traversed.Has(entity.PK) {
			continue
		}
		traversed.ReplaceOrInsert(entity.PK)
		switch pattern.Pattern {
		case sparql.PATTERN_SINGLE:
			// [found] indicates whether or not we have any edges with the given pattern
			edges, found := entity.InEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range edges {
				results.Add(entityHash)
			}
			// because this is one hop, we don't add any new entities to the stack
		case sparql.PATTERN_ZERO_ONE:
			results.Add(entity.PK)
			endpoints, found := entity.InEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range endpoints {
				results.Add(entityHash)
			}
			// because this is one hop, we don't add any new entities to the stack
		case sparql.PATTERN_ZERO_PLUS:
			results.Add(entity.PK)
			// faster index
			if !db.loading {
				if index := db.MustGetEntityIndexFromHash(entity.PK); index != nil {
					if endpoints, found := index.InPlusEdges[string(predHash[:])]; found {
						for _, entityHash := range endpoints {
							results.Add(entityHash)
						}
						return
					}
				}
			}
			endpoints, found := entity.InEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range endpoints {
				nextEntity := db.MustGetEntityFromHash(entityHash)
				if !results.Has(nextEntity.PK) {
					searchstack.PushBack(nextEntity)
				}
				results.Add(entityHash)
				stack.PushBack(nextEntity)
			}
		case sparql.PATTERN_ONE_PLUS:
			// faster index
			if !db.loading {
				index := db.MustGetEntityIndexFromHash(entity.PK)
				if index != nil {
					if endpoints, found := index.InPlusEdges[string(predHash[:])]; found {
						for _, entityHash := range endpoints {
							results.Add(entityHash)
						}
						return
					}
				}
			}
			edges, found := entity.InEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range edges {
				nextEntity := db.MustGetEntityFromHash(entityHash)
				results.Add(nextEntity.PK)
				searchstack.PushBack(nextEntity)
				// also make sure to add this to the stack so we can search
				stack.PushBack(nextEntity)
			}
		}
	}
}

// follow the pattern from the given subject's OutEdges, placing the results in the btree
func (db *DB) followPathFromSubject(subject *Entity, results *keyTree, searchstack *list.List, pattern sparql.PathPattern) {
	stack := list.New()
	stack.PushFront(subject)

	predHash, err := db.GetHash(pattern.Predicate)
	if err != nil {
		log.Error(errors.Wrapf(err, "Not found: %v", pattern.Predicate))
		return
	}

	var traversed = traversedBTreePool.Get()
	defer traversedBTreePool.Put(traversed)

	for stack.Len() > 0 {
		entity := stack.Remove(stack.Front()).(*Entity)
		if traversed.Has(entity.PK) {
			continue
		}
		traversed.ReplaceOrInsert(entity.PK)
		switch pattern.Pattern {
		case sparql.PATTERN_SINGLE:
			// [found] indicates whether or not we have any edges with the given pattern
			endpoints, found := entity.OutEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range endpoints {
				results.Add(entityHash)
			}
			// because this is one hop, we don't add any new entities to the stack
		case sparql.PATTERN_ZERO_ONE:
			// this does not require the pattern to exist, so we add the current entity plus any
			// connected by the appropriate edge
			results.Add(entity.PK)
			endpoints, found := entity.OutEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range endpoints {
				results.Add(entityHash)
			}
			// because this is one hop, we don't add any new entities to the stack
		case sparql.PATTERN_ZERO_PLUS:
			results.Add(entity.PK)
			// faster index
			if !db.loading {
				index := db.MustGetEntityIndexFromHash(entity.PK)
				if index != nil {
					if endpoints, found := index.OutPlusEdges[string(predHash[:])]; found {
						for _, entityHash := range endpoints {
							results.Add(entityHash)
						}
						return
					}
				}
			}

			endpoints, found := entity.OutEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range endpoints {
				nextEntity := db.MustGetEntityFromHash(entityHash)
				if !results.Has(nextEntity.PK) {
					searchstack.PushBack(nextEntity)
				}
				results.Add(entityHash)
				stack.PushBack(nextEntity)
			}
		case sparql.PATTERN_ONE_PLUS:
			// faster index
			if !db.loading {
				index := db.MustGetEntityIndexFromHash(entity.PK)
				if index != nil {
					if endpoints, found := index.OutPlusEdges[string(predHash[:])]; found {
						for _, entityHash := range endpoints {
							results.Add(entityHash)
						}
						return
					}
				}
			}
			edges, found := entity.OutEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range edges {
				nextEntity := db.MustGetEntityFromHash(entityHash)
				results.Add(nextEntity.PK)
				searchstack.PushBack(nextEntity)
				// also make sure to add this to the stack so we can search
				stack.PushBack(nextEntity)
			}
		}
	}
}
