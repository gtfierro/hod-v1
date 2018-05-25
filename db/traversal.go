package db

import (
	"container/list"
	sparql "github.com/gtfierro/hod/lang/ast"
	"github.com/gtfierro/hod/turtle"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

type traversable interface {
	getHash(turtle.URI) (Key, error)
	getURI(Key) (turtle.URI, error)
	getEntityByURI(turtle.URI) (*Entity, error)
	getEntityByHash(Key) (*Entity, error)
	getExtendedIndexByURI(turtle.URI) (*EntityExtendedIndex, error)
	getExtendedIndexByHash(Key) (*EntityExtendedIndex, error)
	getPredicateByURI(turtle.URI) (*PredicateEntity, error)
	getPredicateByHash(Key) (*PredicateEntity, error)
	iterAllEntities(func(Key, *Entity) bool) error

	getReverseRelationship(turtle.URI) (turtle.URI, bool)
	done() error
}

type traversal struct {
	under traversable
	cache *dbcache
}

func (t *traversal) getHash(uri turtle.URI) (Key, error) {
	if t.cache == nil {
		return t.under.getHash(uri)
	}
	if hash, found := t.cache.getHash(uri); !found {
		hash, err := t.under.getHash(uri)
		if err == nil {
			t.cache.setHash(uri, hash)
		}
		return hash, err
	} else {
		return hash, nil
	}
}

func (t *traversal) getURI(hash Key) (turtle.URI, error) {
	if t.cache == nil {
		return t.under.getURI(hash)
	}
	if uri, found := t.cache.getURI(hash); !found {
		uri, err := t.under.getURI(hash)
		if err == nil {
			t.cache.setURI(hash, uri)
		}
		return uri, err
	} else {
		return uri, nil
	}
}

func (t *traversal) getEntityByURI(uri turtle.URI) (*Entity, error) {
	if t.cache == nil {
		return t.under.getEntityByURI(uri)
	}
	hash, err := t.getHash(uri)
	if err != nil {
		return nil, err
	}
	return t.getEntityByHash(hash)
}

func (t *traversal) getEntityByHash(hash Key) (*Entity, error) {
	if t.cache == nil {
		return t.under.getEntityByHash(hash)
	}
	if ent, found := t.cache.getEntityByHash(hash); !found {
		ent, err := t.under.getEntityByHash(hash)
		if err == nil {
			t.cache.setEntityByHash(hash, ent)
		}
		return ent, err
	} else {
		return ent, nil
	}
}

func (t *traversal) getExtendedIndexByURI(uri turtle.URI) (*EntityExtendedIndex, error) {
	if t.cache == nil {
		return t.under.getExtendedIndexByURI(uri)
	}
	hash, err := t.getHash(uri)
	if err != nil {
		return nil, err
	}
	return t.getExtendedIndexByHash(hash)
}

func (t *traversal) getExtendedIndexByHash(hash Key) (*EntityExtendedIndex, error) {
	if t.cache == nil {
		return t.under.getExtendedIndexByHash(hash)
	}
	if ext, found := t.cache.getExtendedIndexByHash(hash); !found {
		ext, err := t.under.getExtendedIndexByHash(hash)
		if err == nil {
			t.cache.setExtendedIndexByHash(hash, ext)
		}
		return ext, err
	} else {
		return ext, nil
	}
}

func (t *traversal) getPredicateByURI(uri turtle.URI) (*PredicateEntity, error) {
	if t.cache == nil {
		return t.under.getPredicateByURI(uri)
	}
	hash, err := t.getHash(uri)
	if err != nil {
		return nil, err
	}
	return t.getPredicateByHash(hash)
}

func (t *traversal) getPredicateByHash(hash Key) (*PredicateEntity, error) {
	if t.cache == nil {
		return t.under.getPredicateByHash(hash)
	}
	if pred, found := t.cache.getPredicateByHash(hash); !found {
		pred, err := t.under.getPredicateByHash(hash)
		if err == nil {
			t.cache.setPredicateByHash(hash, pred)
		}
		return pred, err
	} else {
		return pred, nil
	}
}

// takes the inverse of every relationship. If no inverse exists, returns nil
func (t *traversal) reversePathPattern(path []sparql.PathPattern) []sparql.PathPattern {
	var reverse = make([]sparql.PathPattern, len(path))
	for idx, pred := range path {
		if inverse, found := t.under.getReverseRelationship(pred.Predicate); found {
			pred.Predicate = inverse
			reverse[idx] = pred
		} else {
			return nil
		}
	}
	return reversePath(reverse)
}

// follow the pattern from the given object's InEdges, placing the results in the btree
func (t *traversal) followPathFromObject(object *Entity, results *keymap, searchstack *list.List, pattern sparql.PathPattern) error {
	stack := list.New()
	stack.PushFront(object)

	predHash, err := t.getHash(pattern.Predicate)
	if err != nil && err == leveldb.ErrNotFound {
		log.Infof("Adding unseen predicate %s", pattern.Predicate)
		panic("GOT TO HERE")
		//var hashdest Key
		//if err := snap.db.insertEntity(pattern.Predicate, hashdest[:]); err != nil {
		//	panic(fmt.Errorf("Could not insert entity %s (%v)", pattern.Predicate, err))
		//}
	} else if err != nil {
		return errors.Wrapf(err, "Not found: %v", pattern.Predicate)
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
			if index, err := t.getExtendedIndexByHash(entity.PK); err != nil {
				return err
			} else if index != nil {
				if endpoints, found := index.InPlusEdges[string(predHash[:])]; found {
					for _, entityHash := range endpoints {
						results.Add(entityHash)
					}
					return nil
				}
			}
			endpoints, found := entity.InEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range endpoints {
				nextEntity, err := t.getEntityByHash(entityHash)
				if err != nil {
					return err
				}
				if !results.Has(nextEntity.PK) {
					searchstack.PushBack(nextEntity)
				}
				results.Add(entityHash)
				stack.PushBack(nextEntity)
			}
		case sparql.PATTERN_ONE_PLUS:
			// faster index
			index, err := t.getExtendedIndexByHash(entity.PK)
			if err != nil {
				return err
			}
			if index != nil {
				if endpoints, found := index.InPlusEdges[string(predHash[:])]; found {
					for _, entityHash := range endpoints {
						results.Add(entityHash)
					}
					return nil
				}
			}
			edges, found := entity.InEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range edges {
				nextEntity, err := t.getEntityByHash(entityHash)
				if err != nil {
					return err
				}
				results.Add(nextEntity.PK)
				searchstack.PushBack(nextEntity)
				// also make sure to add this to the stack so we can search
				stack.PushBack(nextEntity)
			}
		}
	}
	return nil
}

// follow the pattern from the given subject's OutEdges, placing the results in the btree
func (t *traversal) followPathFromSubject(subject *Entity, results *keymap, searchstack *list.List, pattern sparql.PathPattern) error {
	stack := list.New()
	stack.PushFront(subject)

	predHash, err := t.getHash(pattern.Predicate)
	if err != nil {
		return errors.Wrapf(err, "Not found: %v", pattern.Predicate)
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
			index, err := t.getExtendedIndexByHash(entity.PK)
			if err != nil {
				return err
			}
			if index != nil {
				if endpoints, found := index.OutPlusEdges[string(predHash[:])]; found {
					for _, entityHash := range endpoints {
						results.Add(entityHash)
					}
					return nil
				}
			}

			endpoints, found := entity.OutEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range endpoints {
				nextEntity, err := t.getEntityByHash(entityHash)
				if err != nil {
					return err
				}
				if !results.Has(nextEntity.PK) {
					searchstack.PushBack(nextEntity)
				}
				results.Add(entityHash)
				stack.PushBack(nextEntity)
			}
		case sparql.PATTERN_ONE_PLUS:
			// faster index
			index, err := t.getExtendedIndexByHash(entity.PK)
			if err != nil {
				return err
			}
			if index != nil {
				if endpoints, found := index.OutPlusEdges[string(predHash[:])]; found {
					for _, entityHash := range endpoints {
						results.Add(entityHash)
					}
					return nil
				}
			}
			edges, found := entity.OutEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range edges {
				nextEntity, err := t.getEntityByHash(entityHash)
				if err != nil {
					return err
				}
				results.Add(nextEntity.PK)
				searchstack.PushBack(nextEntity)
				// also make sure to add this to the stack so we can search
				stack.PushBack(nextEntity)
			}
		}
	}
	return nil
}

func (t *traversal) getSubjectFromPredObject(objectHash Key, path []sparql.PathPattern) (*keymap, error) {
	// first get the initial object entity from the db
	// then we're going to conduct a BFS search starting from this entity looking for all entities
	// that have the required path sequence. We place the results in a BTree to maintain uniqueness

	// So how does this traversal actually work?
	// At each 'step', we are looking at an entity and some offset into the path.

	// get the object, look in its "in" edges for the path pattern
	objEntity, err := t.getEntityByHash(objectHash)
	if err != nil {
		return nil, errors.Wrapf(err, "Not found: %v", objectHash)
	}

	stack := list.New()
	stack.PushFront(objEntity)

	var traversed = traversedBTreePool.Get()
	defer traversedBTreePool.Put(traversed)
	// reverse the path because we are getting from the object
	path = reversePath(path)

	for idx, segment := range path {
		// clear out the tree
		for traversed.Max() != nil {
			traversed.DeleteMax()
		}
		reachable := newKeymap()
		for stack.Len() > 0 {
			entity := stack.Remove(stack.Front()).(*Entity)
			// if we have already traversed this entity, skip it
			if traversed.Has(entity.PK) {
				continue
			}
			// mark this entity as traversed
			traversed.ReplaceOrInsert(entity.PK)
			t.followPathFromObject(entity, reachable, stack, segment)
		}

		// if we aren't done, then we push these items onto the stack
		if idx < len(path)-1 {
			reachable.Iter(func(key Key) {
				ent, err := t.getEntityByHash(key)
				if err != nil {
					log.Error(err)
					return
				}
				stack.PushBack(ent)
			})
		} else {
			return reachable, nil
		}
	}
	return newKeymap(), nil
}

// Given object and predicate, get all subjects
func (t *traversal) getObjectFromSubjectPred(subjectHash Key, path []sparql.PathPattern) *keymap {
	subEntity, err := t.getEntityByHash(subjectHash)
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
		reachable := newKeymap()
		for stack.Len() > 0 {
			entity := stack.Remove(stack.Front()).(*Entity)
			// if we have already traversed this entity, skip it
			if traversed.Has(entity.PK) {
				continue
			}
			// mark this entity as traversed
			traversed.ReplaceOrInsert(entity.PK)
			t.followPathFromSubject(entity, reachable, stack, segment)
		}

		// if we aren't done, then we push these items onto the stack
		if idx < len(path)-1 {
			reachable.Iter(func(key Key) {
				ent, err := t.getEntityByHash(key)
				if err != nil {
					log.Error(err)
					return
				}
				stack.PushBack(ent)
			})
		} else {
			return reachable
		}
	}
	return newKeymap()
}

// Given a predicate, it returns pairs of (subject, object) that are connected by that relationship
func (t *traversal) getSubjectObjectFromPred(path []sparql.PathPattern) (soPair [][]Key, err error) {
	var pe *PredicateEntity
	pe, err = t.getPredicateByURI(path[0].Predicate)
	if err != nil {
		err = errors.Wrapf(err, "Can't find predicate %v", path[0].Predicate)
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
	return soPair, nil
}

func (t *traversal) getPredicateFromSubjectObject(subject, object *Entity) *keymap {
	reachable := newKeymap()

	for edge, objects := range subject.InEdges {
		for _, edgeObject := range objects {
			if edgeObject == object.PK {
				// matches!
				var edgepk Key
				edgepk.FromSlice([]byte(edge))
				reachable.Add(edgepk)
			}
		}
	}
	for edge, objects := range subject.OutEdges {
		for _, edgeObject := range objects {
			if edgeObject == object.PK {
				// matches!
				var edgepk Key
				edgepk.FromSlice([]byte(edge))
				reachable.Add(edgepk)
			}
		}
	}

	return reachable
}

func (t *traversal) getPredicatesFromObject(object *Entity) *keymap {
	reachable := newKeymap()
	var edgepk Key
	for edge := range object.InEdges {
		edgepk.FromSlice([]byte(edge))
		reachable.Add(edgepk)
	}

	return reachable
}

func (t *traversal) getPredicatesFromSubject(subject *Entity) *keymap {
	reachable := newKeymap()
	var edgepk Key
	for edge := range subject.OutEdges {
		edgepk.FromSlice([]byte(edge))
		reachable.Add(edgepk)
	}

	return reachable
}

func (t *traversal) iterAllEntities(F func(Key, *Entity) bool) error {
	return t.under.iterAllEntities(F)
}
