package db

import (
	"container/list"
	"fmt"

	"github.com/gtfierro/hod/query"

	"github.com/mitghi/btree"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

// takes the inverse of every relationship. If no inverse exists, returns nil
func (db *DB) reversePathPattern(path []query.PathPattern) []query.PathPattern {
	var reverse = make([]query.PathPattern, len(path))
	for idx, pred := range path {
		if inverse, found := db.relationships[pred.Predicate]; found {
			pred.Predicate = inverse
			reverse[idx] = pred
		} else {
			return nil
		}
	}
	reversePath(reverse)
	return reverse
}

// follow the pattern from the given object's InEdges, placing the results in the btree
func (db *DB) followPathFromObject(object *Entity, results *btree.BTree, searchstack *list.List, pattern query.PathPattern) {
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
		case query.PATTERN_SINGLE:
			// [found] indicates whether or not we have any edges with the given pattern
			edges, found := entity.InEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range edges {
				results.ReplaceOrInsert(entityHash)
			}
			// because this is one hop, we don't add any new entities to the stack
		case query.PATTERN_ZERO_ONE:
			results.ReplaceOrInsert(entity.PK)
			endpoints, found := entity.InEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range endpoints {
				results.ReplaceOrInsert(entityHash)
			}
			// because this is one hop, we don't add any new entities to the stack
		case query.PATTERN_ZERO_PLUS:
			results.ReplaceOrInsert(entity.PK)
			// faster index
			if !db.loading {
				index := db.MustGetEntityIndexFromHash(entity.PK)
				if index != nil {
					if endpoints, found := index.InPlusEdges[string(predHash[:])]; found {
						for _, entityHash := range endpoints {
							results.ReplaceOrInsert(entityHash)
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
				results.ReplaceOrInsert(entityHash)
				stack.PushBack(nextEntity)
			}
		case query.PATTERN_ONE_PLUS:
			// faster index
			if !db.loading {
				index := db.MustGetEntityIndexFromHash(entity.PK)
				if index != nil {
					if endpoints, found := index.InPlusEdges[string(predHash[:])]; found {
						for _, entityHash := range endpoints {
							results.ReplaceOrInsert(entityHash)
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
				results.ReplaceOrInsert(nextEntity.PK)
				searchstack.PushBack(nextEntity)
				// also make sure to add this to the stack so we can search
				stack.PushBack(nextEntity)
			}
		}
	}
}

// follow the pattern from the given subject's OutEdges, placing the results in the btree
func (db *DB) followPathFromSubject(subject *Entity, results *btree.BTree, searchstack *list.List, pattern query.PathPattern) {
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
		case query.PATTERN_SINGLE:
			// [found] indicates whether or not we have any edges with the given pattern
			endpoints, found := entity.OutEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range endpoints {
				results.ReplaceOrInsert(entityHash)
			}
			// because this is one hop, we don't add any new entities to the stack
		case query.PATTERN_ZERO_ONE:
			// this does not require the pattern to exist, so we add the current entity plus any
			// connected by the appropriate edge
			results.ReplaceOrInsert(entity.PK)
			endpoints, found := entity.OutEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range endpoints {
				results.ReplaceOrInsert(entityHash)
			}
			// because this is one hop, we don't add any new entities to the stack
		case query.PATTERN_ZERO_PLUS:
			results.ReplaceOrInsert(entity.PK)
			// faster index
			if !db.loading {
				index := db.MustGetEntityIndexFromHash(entity.PK)
				if index != nil {
					if endpoints, found := index.OutPlusEdges[string(predHash[:])]; found {
						for _, entityHash := range endpoints {
							results.ReplaceOrInsert(entityHash)
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
				results.ReplaceOrInsert(entityHash)
				stack.PushBack(nextEntity)
			}
		case query.PATTERN_ONE_PLUS:
			// faster index
			if !db.loading {
				index := db.MustGetEntityIndexFromHash(entity.PK)
				if index != nil {
					if endpoints, found := index.OutPlusEdges[string(predHash[:])]; found {
						for _, entityHash := range endpoints {
							results.ReplaceOrInsert(entityHash)
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
				results.ReplaceOrInsert(nextEntity.PK)
				searchstack.PushBack(nextEntity)
				// also make sure to add this to the stack so we can search
				stack.PushBack(nextEntity)
			}
		}
	}
}

func (db *DB) getSubjectFromPredObject(objectHash Key, path []query.PathPattern) *btree.BTree {
	// first get the initial object entity from the db
	// then we're going to conduct a BFS search starting from this entity looking for all entities
	// that have the required path sequence. We place the results in a BTree to maintain uniqueness

	// So how does this traversal actually work?
	// At each 'step', we are looking at an entity and some offset into the path.

	// get the object, look in its "in" edges for the path pattern
	objEntity, err := db.GetEntityFromHash(objectHash)
	if err != nil {
		log.Error(errors.Wrapf(err, "Not found: %v", objectHash))
		return nil
	}

	stack := list.New()
	stack.PushFront(objEntity)

	var traversed = traversedBTreePool.Get()
	defer traversedBTreePool.Put(traversed)
	// reverse the path because we are getting from the object
	reversePath(path)

	for idx, segment := range path {
		// clear out the tree
		for traversed.Max() != nil {
			traversed.DeleteMax()
		}
		reachable := btree.New(BTREE_DEGREE, "")
		for stack.Len() > 0 {
			entity := stack.Remove(stack.Front()).(*Entity)
			// if we have already traversed this entity, skip it
			if traversed.Has(entity.PK) {
				continue
			}
			// mark this entity as traversed
			traversed.ReplaceOrInsert(entity.PK)
			db.followPathFromObject(entity, reachable, stack, segment)
		}

		// if we aren't done, then we push these items onto the stack
		if idx < len(path)-1 {
			max := reachable.Max()
			iter := func(i btree.Item) bool {
				ent, err := db.GetEntityFromHash(i.(Key))
				if err != nil {
					log.Error(err)
					return false
				}
				stack.PushBack(ent)
				return i != max
			}
			reachable.Ascend(iter)
		} else {
			return reachable
		}
	}
	return btree.New(BTREE_DEGREE, "")
}

// Given object and predicate, get all subjects
func (db *DB) getObjectFromSubjectPred(subjectHash Key, path []query.PathPattern) *btree.BTree {
	subEntity, err := db.GetEntityFromHash(subjectHash)
	if err != nil {
		log.Error(errors.Wrapf(err, "Not found: %v", subjectHash))
		return nil
	}

	// stack of entities to search
	stack := list.New()
	stack.PushFront(subEntity)
	var traversed = traversedBTreePool.Get()
	defer traversedBTreePool.Put(traversed)

	// we have our starting entity; follow the first segment of the path and save everything we can reach from there.
	// Then, from that set, search the second segment of the path, etc. We save the last reachable set
	for idx, segment := range path {
		// clear out the tree
		for traversed.Max() != nil {
			traversed.DeleteMax()
		}
		reachable := btree.New(BTREE_DEGREE, "")
		for stack.Len() > 0 {
			entity := stack.Remove(stack.Front()).(*Entity)
			// if we have already traversed this entity, skip it
			if traversed.Has(entity.PK) {
				continue
			}
			// mark this entity as traversed
			traversed.ReplaceOrInsert(entity.PK)
			db.followPathFromSubject(entity, reachable, stack, segment)
		}

		// if we aren't done, then we push these items onto the stack
		if idx < len(path)-1 {
			max := reachable.Max()
			iter := func(i btree.Item) bool {
				ent, err := db.GetEntityFromHash(i.(Key))
				if err != nil {
					log.Error(err)
					return false
				}
				stack.PushBack(ent)
				return i != max
			}
			reachable.Ascend(iter)
		} else {
			return reachable
		}
	}
	return btree.New(BTREE_DEGREE, "")
}

// Given a predicate, it returns pairs of (subject, object) that are connected by that relationship
func (db *DB) getSubjectObjectFromPred(path []query.PathPattern) (soPair [][]Key) {
	pe, found := db.predIndex[path[0].Predicate]
	if !found {
		log.Errorf("Can't find predicate: %v", path[0].Predicate)
		return
	}
	for subject, objectMap := range pe.Subjects {
		for object := range objectMap {
			var sh, oh Key
			sh.FromSlice([]byte(subject))
			oh.FromSlice([]byte(object))
			soPair = append(soPair, []Key{sh, oh})
		}
	}
	return soPair
}

func (db *DB) getPredicateFromSubjectObject(subject, object *Entity) *btree.BTree {
	reachable := btree.New(BTREE_DEGREE, "")

	for edge, objects := range subject.InEdges {
		for _, edgeObject := range objects {
			if edgeObject == object.PK {
				// matches!
				var edgepk Key
				edgepk.FromSlice([]byte(edge))
				reachable.ReplaceOrInsert(edgepk)
			}
		}
	}
	for edge, objects := range subject.OutEdges {
		for _, edgeObject := range objects {
			if edgeObject == object.PK {
				// matches!
				var edgepk Key
				edgepk.FromSlice([]byte(edge))
				reachable.ReplaceOrInsert(edgepk)
			}
		}
	}

	return reachable
}

func (db *DB) getPredicatesFromObject(object *Entity) *btree.BTree {
	reachable := btree.New(BTREE_DEGREE, "")
	var edgepk Key
	for edge := range object.InEdges {
		edgepk.FromSlice([]byte(edge))
		reachable.ReplaceOrInsert(edgepk)
	}

	return reachable
}

func (db *DB) getPredicatesFromSubject(subject *Entity) *btree.BTree {
	reachable := btree.New(BTREE_DEGREE, "")
	var edgepk Key
	for edge := range subject.OutEdges {
		edgepk.FromSlice([]byte(edge))
		reachable.ReplaceOrInsert(edgepk)
	}

	return reachable
}
