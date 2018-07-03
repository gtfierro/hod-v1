package db

import (
	"container/list"

	//"github.com/coocood/freecache"
	sparql "github.com/gtfierro/hod/lang/ast"
	"github.com/gtfierro/hod/storage"
	"github.com/gtfierro/hod/turtle"
	"github.com/pkg/errors"
)

type snapshot struct {
	db       *DB
	snapshot storage.Traversable
}

func (db *DB) snapshot() (snap *snapshot, err error) {
	snap = &snapshot{db: db}
	if snap.snapshot, err = db.backing.OpenSnapshot(); err != nil {
		return nil, err
	}
	return
}

func (snap *snapshot) Close() {
	snap.snapshot.Release()
}

func (snap *snapshot) done() error {
	snap.snapshot.Release()
	return nil
}

/*** Get URI methods ***/

func (snap *snapshot) getURI(hash Key) (turtle.URI, error) {
	val, err := snap.snapshot.Get(storage.PKBucket, hash[:])
	if err != nil {
		return turtle.URI{}, err
	}
	uri := turtle.ParseURI(string(val))
	return uri, nil
}

func (snap *snapshot) MustGetURI(hash Key) turtle.URI {
	if hash == emptyKey {
		return turtle.URI{}
	}
	uri, err := snap.getURI(hash)
	if err != nil {
		log.Error(errors.Wrapf(err, "Could not get URI for %v", hash))
		return turtle.URI{}
	}
	return uri
}

func (snap *snapshot) getPredicateByURI(uri turtle.URI) (*PredicateEntity, error) {
	hash, err := snap.getHash(uri)
	if err != nil {
		return nil, err
	}
	return snap.getPredicateByHash(hash)
}

func (snap *snapshot) getPredicateByHash(hash Key) (*PredicateEntity, error) {
	var pred = NewPredicateEntity()
	bytes, err := snap.snapshot.Get(storage.PredBucket, hash[:])
	if err != nil && err != storage.ErrNotFound {
		return nil, errors.Wrap(err, "Error getting predicate from transaction")
	} else if err == storage.ErrNotFound {
		pred.PK = hash
		return pred, nil
	} else {
		// load predicate entity from db
		_, err = pred.UnmarshalMsg(bytes)
		return pred, err
	}
}

/*** Get Hash methods ***/

func (snap *snapshot) getHash(entity turtle.URI) (Key, error) {
	var rethash Key
	val, err := snap.snapshot.Get(storage.EntityBucket, entity.Bytes())
	if err != nil {
		return emptyKey, errors.Wrapf(err, "Could not get Entity for %s", entity)
	}
	copy(rethash[:], val)
	if rethash == emptyKey {
		return emptyKey, errors.New("Got bad hash")
	}
	return rethash, nil
}

func (snap *snapshot) MustGetHash(entity turtle.URI) Key {
	val, err := snap.getHash(entity)
	if err != nil {
		log.Error(errors.Wrapf(err, "Could not get hash for %s", entity))
		return emptyKey
	}
	return val
}

/*** Get Entity methods ***/

func (snap *snapshot) getEntityByURI(uri turtle.URI) (*Entity, error) {
	hash, err := snap.getHash(uri)
	if err != nil {
		return nil, err
	}
	return snap.getEntityByHash(hash)
}

func (snap *snapshot) MustGetEntityFromHash(hash Key) *Entity {
	e, err := snap.getEntityByHash(hash)
	if err != nil {
		log.Error(errors.Wrap(err, "Could not get entity"))
		return nil
	}
	return e
}

func (snap *snapshot) getEntityByHash(hash Key) (*Entity, error) {
	bytes, err := snap.snapshot.Get(storage.GraphBucket, hash[:])
	if err != nil {
		return nil, errors.Wrapf(err, "Could not get Entity from graph for %s", snap.MustGetURI(hash))
	}
	ent := NewEntity()
	_, err = ent.UnmarshalMsg(bytes)
	return ent, err
}

func (snap *snapshot) getExtendedIndexByURI(uri turtle.URI) (*EntityExtendedIndex, error) {
	hash, err := snap.getHash(uri)
	if err != nil {
		return nil, err
	}
	return snap.getExtendedIndexByHash(hash)
}

/*** Entity Index methods ***/
func (snap *snapshot) getExtendedIndexByHash(hash Key) (*EntityExtendedIndex, error) {
	bytes, err := snap.snapshot.Get(storage.ExtendedBucket, hash[:])
	if err != nil && err != storage.ErrNotFound {
		return nil, errors.Wrapf(err, "Could not get EntityIndex from graph for %s", snap.MustGetURI(hash))
	} else if err == storage.ErrNotFound {
		return nil, nil
	}
	ent := NewEntityExtendedIndex()
	_, err = ent.UnmarshalMsg(bytes)
	return ent, err
}

func (snap *snapshot) MustGetEntityIndexFromHash(hash Key) *EntityExtendedIndex {
	e, err := snap.getExtendedIndexByHash(hash)
	if err != nil {
		log.Error(errors.Wrap(err, "Could not get entity index"))
		return nil
	}
	return e
}

// takes the inverse of every relationship. If no inverse exists, returns nil
func (snap *snapshot) reversePathPattern(path []sparql.PathPattern) []sparql.PathPattern {
	var reverse = make([]sparql.PathPattern, len(path))
	for idx, pred := range path {
		if inverse, found := snap.db.relationships[pred.Predicate]; found {
			pred.Predicate = inverse
			reverse[idx] = pred
		} else {
			return nil
		}
	}
	return reversePath(reverse)
}

func (snap *snapshot) getReverseRelationship(forward turtle.URI) (reverse turtle.URI, found bool) {
	reverse, found = snap.db.relationships[forward]
	return
}

// follow the pattern from the given object's InEdges, placing the results in the btree
func (snap *snapshot) followPathFromObject(object *Entity, results *keymap, searchstack *list.List, pattern sparql.PathPattern) {
	stack := list.New()
	stack.PushFront(object)

	predHash, err := snap.getHash(pattern.Predicate)
	if err != nil && err == storage.ErrNotFound {
		log.Infof("Adding unseen predicate %s", pattern.Predicate)
		panic("GOT TO HERE")
		//var hashdest Key
		//if err := snap.db.insertEntity(pattern.Predicate, hashdest[:]); err != nil {
		//	panic(fmt.Errorf("Could not insert entity %s (%v)", pattern.Predicate, err))
		//}
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
			if index := snap.MustGetEntityIndexFromHash(entity.PK); index != nil {
				if endpoints, found := index.InPlusEdges[string(predHash[:])]; found {
					for _, entityHash := range endpoints {
						results.Add(entityHash)
					}
					return
				}
			}
			endpoints, found := entity.InEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range endpoints {
				nextEntity := snap.MustGetEntityFromHash(entityHash)
				if !results.Has(nextEntity.PK) {
					searchstack.PushBack(nextEntity)
				}
				results.Add(entityHash)
				stack.PushBack(nextEntity)
			}
		case sparql.PATTERN_ONE_PLUS:
			// faster index
			index := snap.MustGetEntityIndexFromHash(entity.PK)
			if index != nil {
				if endpoints, found := index.InPlusEdges[string(predHash[:])]; found {
					for _, entityHash := range endpoints {
						results.Add(entityHash)
					}
					return
				}
			}
			edges, found := entity.InEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range edges {
				nextEntity := snap.MustGetEntityFromHash(entityHash)
				results.Add(nextEntity.PK)
				searchstack.PushBack(nextEntity)
				// also make sure to add this to the stack so we can search
				stack.PushBack(nextEntity)
			}
		}
	}
}

// follow the pattern from the given subject's OutEdges, placing the results in the btree
func (snap *snapshot) followPathFromSubject(subject *Entity, results *keymap, searchstack *list.List, pattern sparql.PathPattern) {
	stack := list.New()
	stack.PushFront(subject)

	predHash, err := snap.getHash(pattern.Predicate)
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
			index := snap.MustGetEntityIndexFromHash(entity.PK)
			if index != nil {
				if endpoints, found := index.OutPlusEdges[string(predHash[:])]; found {
					for _, entityHash := range endpoints {
						results.Add(entityHash)
					}
					return
				}
			}

			endpoints, found := entity.OutEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range endpoints {
				nextEntity := snap.MustGetEntityFromHash(entityHash)
				if !results.Has(nextEntity.PK) {
					searchstack.PushBack(nextEntity)
				}
				results.Add(entityHash)
				stack.PushBack(nextEntity)
			}
		case sparql.PATTERN_ONE_PLUS:
			// faster index
			index := snap.MustGetEntityIndexFromHash(entity.PK)
			if index != nil {
				if endpoints, found := index.OutPlusEdges[string(predHash[:])]; found {
					for _, entityHash := range endpoints {
						results.Add(entityHash)
					}
					return
				}
			}
			edges, found := entity.OutEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range edges {
				nextEntity := snap.MustGetEntityFromHash(entityHash)
				results.Add(nextEntity.PK)
				searchstack.PushBack(nextEntity)
				// also make sure to add this to the stack so we can search
				stack.PushBack(nextEntity)
			}
		}
	}
}

func (snap *snapshot) getSubjectFromPredObject(objectHash Key, path []sparql.PathPattern) *keymap {
	// first get the initial object entity from the db
	// then we're going to conduct a BFS search starting from this entity looking for all entities
	// that have the required path sequence. We place the results in a BTree to maintain uniqueness

	// So how does this traversal actually work?
	// At each 'step', we are looking at an entity and some offset into the path.

	// get the object, look in its "in" edges for the path pattern
	objEntity, err := snap.getEntityByHash(objectHash)
	if err != nil {
		log.Error(errors.Wrapf(err, "Not found: %v", objectHash))
		return nil
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
			snap.followPathFromObject(entity, reachable, stack, segment)
		}

		// if we aren't done, then we push these items onto the stack
		if idx < len(path)-1 {
			reachable.Iter(func(key Key) {
				ent, err := snap.getEntityByHash(key)
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

// Given object and predicate, get all subjects
func (snap *snapshot) getObjectFromSubjectPred(subjectHash Key, path []sparql.PathPattern) *keymap {
	subEntity, err := snap.getEntityByHash(subjectHash)
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
			snap.followPathFromSubject(entity, reachable, stack, segment)
		}

		// if we aren't done, then we push these items onto the stack
		if idx < len(path)-1 {
			reachable.Iter(func(key Key) {
				ent, err := snap.getEntityByHash(key)
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
func (snap *snapshot) getSubjectObjectFromPred(path []sparql.PathPattern) (soPair [][]Key) {
	pe, found := snap.db.predIndex[path[0].Predicate]
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

func (snap *snapshot) getPredicateFromSubjectObject(subject, object *Entity) *keymap {
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

func (snap *snapshot) getPredicatesFromObject(object *Entity) *keymap {
	reachable := newKeymap()
	var edgepk Key
	for edge := range object.InEdges {
		edgepk.FromSlice([]byte(edge))
		reachable.Add(edgepk)
	}

	return reachable
}

func (snap *snapshot) getPredicatesFromSubject(subject *Entity) *keymap {
	reachable := newKeymap()
	var edgepk Key
	for edge := range subject.OutEdges {
		edgepk.FromSlice([]byte(edge))
		reachable.Add(edgepk)
	}

	return reachable
}

func (snap *snapshot) iterAllEntities(F func(Key, *Entity) bool) error {
	iter := snap.snapshot.Iterate(storage.GraphBucket)
	for iter.Next() {
		var subjectHash Key
		entityHash := iter.Key()
		copy(subjectHash[:], entityHash[:8])
		var entity = NewEntity()
		_, err := entity.UnmarshalMsg(iter.Value())
		if err != nil {
			return err
		}
		if F(subjectHash, entity) {
			return nil
		}
	}
	return nil
}
