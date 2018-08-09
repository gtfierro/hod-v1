package db

import (
	"time"

	"container/list"
	sparql "github.com/gtfierro/hod/lang/ast"
	"github.com/gtfierro/hod/storage"
	"github.com/gtfierro/hod/turtle"
	"github.com/pkg/errors"
	logrus "github.com/sirupsen/logrus"
)

const (
	//RDF_NAMESPACE = "http://www.w3.org/1999/02/22-rdf-syntax-ns"
	owlNamespace = "http://www.w3.org/2002/07/owl"
)

var (
	inverseOf = turtle.URI{
		Namespace: owlNamespace,
		Value:     "inverseOf",
	}
)

type transaction struct {
	snapshot             storage.Transaction
	triplesAdded         int
	hashes               map[turtle.URI]storage.HashKey
	inverseRelationships map[storage.HashKey]storage.HashKey
	cache                *dbcache
}

func (hod *HodDB) openTransaction(name string) (tx *transaction, err error) {
	hod.Lock()
	defer hod.Unlock()
	tx = &transaction{
		hashes:               make(map[turtle.URI]storage.HashKey),
		inverseRelationships: make(map[storage.HashKey]storage.HashKey),
		cache:                newCache(1),
	}
	tx.snapshot, err = hod.storage.CreateVersion(name)
	hod.loaded_versions[tx.snapshot.Version()] = tx
	logrus.Info("Using new version ", tx.snapshot.Version())
	return
}

func (hod *HodDB) openVersion(ver storage.Version) (tx *transaction, err error) {
	var found bool
	hod.RLock()
	if tx, found = hod.loaded_versions[ver]; found {
		defer hod.RUnlock()
		logrus.Info("Using existing version ", ver)
		err = nil
		return
	}
	hod.RUnlock()
	hod.Lock()
	defer hod.Unlock()
	for loadedVer, tx := range hod.loaded_versions {
		if loadedVer.Name != ver.Name {
			logrus.Info("Discarding old version ", loadedVer)
			tx.discard()
			delete(hod.loaded_versions, loadedVer)
		}
	}
	tx = &transaction{
		hashes:               make(map[turtle.URI]storage.HashKey),
		inverseRelationships: make(map[storage.HashKey]storage.HashKey),
		cache:                newCache(1),
	}
	tx.snapshot, err = hod.storage.OpenVersion(ver)
	hod.loaded_versions[tx.snapshot.Version()] = tx
	logrus.Info("Using newer version ", tx.snapshot.Version())
	return
}

func (tx *transaction) discard() {
	tx.snapshot.Release()
}

func (tx *transaction) commit() error {
	return tx.snapshot.Commit()
}

func (tx *transaction) addTriples(dataset turtle.DataSet) error {
	var newPredicates = make(map[storage.HashKey]struct{})

	addStart := time.Now()
	// add all URIs to the database
	for _, triple := range dataset.Triples {
		if err := tx.addTriple(triple); err != nil {
			return errors.Wrapf(err, "Could not load triple (%s)", triple)
		}
		newPredicates[tx.hashes[triple.Predicate]] = struct{}{} // mark new predicate

		// if triple defines an inverseOf relationship, then track the subject/object of that
		// triple so we can populate the graph later
		if triple.Predicate.Namespace == owlNamespace && triple.Predicate.Value == "inverseOf" {
			subjectHash := tx.hashes[triple.Subject]
			objectHash := tx.hashes[triple.Object]
			tx.inverseRelationships[subjectHash] = objectHash
			tx.inverseRelationships[objectHash] = subjectHash
			if _, err := tx.getPredicateByURI(triple.Subject); err != nil {
				return err
			}
			if _, err := tx.getPredicateByURI(triple.Object); err != nil {
				return err
			}
		}
		tx.triplesAdded++
	}
	addEnd := time.Now()

	// pull out all of the inverse edges from the database and add to inverseRelationships
	reverseEdgeFindStart := time.Now()
	var predicatesAdded int
	pred, err := tx.getPredicateByURI(inverseOf)
	if err != nil && err != storage.ErrNotFound {
		logrus.WithError(err).Error("Could not load inverseOf pred")
	} else if err == nil {
		for _, subject := range pred.GetAllSubjects() {
			for _, object := range pred.GetObjects(subject) {
				tx.inverseRelationships[subject] = object
				tx.inverseRelationships[object] = subject
				predicatesAdded++
			}
		}
	}

	reverseEdgeFindEnd := time.Now()

	// add the inverse edges to the graph index
	reverseEdgeBuildStart := time.Now()
	for predicate, reversePredicate := range tx.inverseRelationships {
		pred, err := tx.snapshot.GetPredicate(predicate)
		_uri, _ := tx.snapshot.GetURI(predicate)
		if err != nil {
			return errors.Wrapf(err, "Could not load predicate %s", _uri)
		}
		revPred, err := tx.snapshot.GetPredicate(reversePredicate)
		_uri, _ = tx.snapshot.GetURI(reversePredicate)
		if err != nil {
			return errors.Wrapf(err, "Could not load reverse predicate %s", _uri)
		}

		for _, subject := range pred.GetAllSubjects() {
			subjectEnt, err := tx.snapshot.GetEntity(subject)
			if err != nil {
				return errors.Wrap(err, "Could not load subject")
			}

			for _, object := range pred.GetObjects(subject) {
				objectEnt, err := tx.snapshot.GetEntity(object)
				if err != nil {
					return errors.Wrap(err, "Could not load object")
				}
				subjectEnt.AddInEdge(reversePredicate, object)
				subjectEnt.AddOutEdge(predicate, object)
				objectEnt.AddOutEdge(reversePredicate, subject)
				objectEnt.AddInEdge(predicate, subject)
				if err = tx.snapshot.PutEntity(objectEnt); err != nil {
					return err
				}

				revPred.AddSubjectObject(object, subject)
			}

			if err = tx.snapshot.PutEntity(subjectEnt); err != nil {
				return err
			}

		}
		if err := tx.snapshot.PutPredicate(revPred); err != nil {
			return err
		}

	}
	reverseEdgeBuildEnd := time.Now()

	extendedBuildStart := time.Now()
	// for all *new* predicates, roll the edges forward for all entities in the transaction.
	for predicateHash := range newPredicates {
		tx.rollupPredicate(predicateHash)
		if reversePredicate, found := tx.inverseRelationships[predicateHash]; found {
			//fmt.Println(reversePredicate)
			// for all entities
			// add the roll-forward index
			tx.rollupPredicate(reversePredicate)
		}
	}
	extendedBuildEnd := time.Now()

	logrus.WithFields(logrus.Fields{
		"EdgeBuild":          reverseEdgeBuildEnd.Sub(reverseEdgeBuildStart),
		"AddTriples":         addEnd.Sub(addStart),
		"EdgeFind":           reverseEdgeFindEnd.Sub(reverseEdgeFindStart),
		"ExtendedIndexBuild": extendedBuildEnd.Sub(extendedBuildStart),
		"Triples":            tx.triplesAdded,
		"Predicates":         predicatesAdded,
	}).Info("Insert")

	return nil
}

func (tx *transaction) addURI(uri turtle.URI) error {
	hash, err := tx.snapshot.PutURI(uri)
	if err != nil {
		return err
	}
	tx.hashes[uri] = hash
	return nil
}

func (tx *transaction) addTriple(triple turtle.Triple) error {
	if triple.Subject.IsEmpty() || triple.Predicate.IsEmpty() || triple.Object.IsEmpty() {
		return nil
	}

	// insert subject, predicate and object
	if err := tx.addURI(triple.Subject); err != nil {
		return err
	}
	if err := tx.addURI(triple.Predicate); err != nil {
		return err
	}
	if err := tx.addURI(triple.Object); err != nil {
		return err
	}

	// add the "1 or more" edge for the extended index
	rev := triple.Predicate
	rev.Value += "+"
	if err := tx.addURI(rev); err != nil {
		return err
	}

	// populate subject, predicate and object in graph index with forward/inverse edges
	var (
		subjectHash   = tx.hashes[triple.Subject]
		predicateHash = tx.hashes[triple.Predicate]
		objectHash    = tx.hashes[triple.Object]
	)

	pred, err := tx.getPredicateByURI(triple.Predicate)
	if err != nil {
		return errors.Wrap(err, "could not get pred")
	}
	if pred.AddSubjectObject(subjectHash, objectHash) {
		if err := tx.snapshot.PutPredicate(pred); err != nil {
			return err
		}
	}

	subject, err := tx.getEntityByURI(triple.Subject)
	if err != nil {
		return errors.Wrap(err, "could not get subject")
	}
	object, err := tx.getEntityByURI(triple.Object)
	if err != nil {
		return errors.Wrap(err, "could not get object")
	}

	if subject.AddOutEdge(predicateHash, object.Key()) {
		if err = tx.snapshot.PutEntity(subject); err != nil {
			return errors.Wrap(err, "could not set out edge")
		}
	}
	if object.AddInEdge(predicateHash, subject.Key()) {
		if err = tx.snapshot.PutEntity(object); err != nil {
			return errors.Wrap(err, "could not set in edge")
		}
	}

	//tx.cache.evict(subjectHash)
	//tx.cache.evict(objectHash)
	//tx.cache.evict(predicateHash)

	return nil
}

func (tx *transaction) rollupPredicate(predicateHash storage.HashKey) error {
	var err error
	forwardPath := sparql.PathPattern{Pattern: sparql.PATTERN_ONE_PLUS}
	results := newKeymap()
	forwardPath.Predicate, err = tx.snapshot.GetURI(predicateHash)
	if err != nil {
		return err
	}
	predicate, err := tx.snapshot.GetPredicate(predicateHash)
	if err != nil {
		return err
	}
	for _, subjectHash := range predicate.GetAllSubjects() {
		subjectIndex, err := tx.snapshot.GetExtendedIndex(subjectHash)
		if err == storage.ErrNotFound {
			subjectIndex = storage.NewEntityExtendedIndex(subjectHash)
			//if err := tx.snapshot.PutExtendedIndex(subjectIndex); err != nil {
			//	return err
			//}
		} else if err != nil {
			return err
		}

		subject, err := tx.snapshot.GetEntity(subjectHash)
		if err != nil {
			return err
		}

		stack := list.New()
		tx.followPathFromSubject(subject, results, stack, forwardPath)
		for results.Len() > 0 {
			objectIndex, err := tx.getExtendedIndexByHash(results.Max())
			if err != nil {
				return err
			}
			subjectIndex.AddOutPlusEdge(predicateHash, results.DeleteMax())
			objectIndex.AddInPlusEdge(predicateHash, subjectHash)
			if err := tx.snapshot.PutExtendedIndex(objectIndex); err != nil {
				return err
			}
		}
		if err := tx.snapshot.PutExtendedIndex(subjectIndex); err != nil {
			return err
		}
	}

	for _, objectHash := range predicate.GetAllObjects() {
		objectIndex, err := tx.snapshot.GetExtendedIndex(objectHash)
		if err == storage.ErrNotFound {
			objectIndex = storage.NewEntityExtendedIndex(objectHash)
			//if err := tx.snapshot.PutExtendedIndex(objectIndex); err != nil {
			//	return err
			//}
		} else if err != nil {
			return err
		}

		object, err := tx.snapshot.GetEntity(objectHash)
		if err != nil {
			return err
		}

		stack := list.New()
		if tx.followPathFromObject(object, results, stack, forwardPath); err != nil {
			logrus.Error(err)
		}
		for results.Len() > 0 {
			subjectIndex, err := tx.getExtendedIndexByHash(results.Max())
			if err != nil {
				return err
			}
			objectIndex.AddInPlusEdge(predicateHash, results.DeleteMax())
			subjectIndex.AddOutPlusEdge(predicateHash, objectHash)
			if err := tx.snapshot.PutExtendedIndex(subjectIndex); err != nil {
				return err
			}
		}
		if err := tx.snapshot.PutExtendedIndex(objectIndex); err != nil {
			return err
		}
	}
	return nil
}

func (tx *transaction) getHash(uri turtle.URI) (storage.HashKey, error) {
	if tx.cache == nil {
		return tx.snapshot.GetHash(uri)
	}
	hash, found := tx.cache.getHash(uri)
	if !found {
		hash, err := tx.snapshot.GetHash(uri)
		if err == nil {
			tx.cache.setHash(uri, hash)
		}
		return hash, err
	}
	return hash, nil
}

func (tx *transaction) getURI(hash storage.HashKey) (turtle.URI, error) {
	if tx.cache == nil {
		return tx.snapshot.GetURI(hash)
	}
	uri, found := tx.cache.getURI(hash)
	if !found {
		uri, err := tx.snapshot.GetURI(hash)
		if err == nil {
			tx.cache.setURI(hash, uri)
		}
		return uri, err
	}
	return uri, nil
}

func (tx *transaction) getEntityByURI(uri turtle.URI) (storage.Entity, error) {
	hash, err := tx.getHash(uri)
	if err != nil {
		return nil, err
	}
	return tx.getEntityByHash(hash)
}

func (tx *transaction) getEntityByHash(hash storage.HashKey) (storage.Entity, error) {
	if tx.cache == nil {
		return tx.snapshot.GetEntity(hash)
	}
	ent, found := tx.cache.getEntityByHash(hash)
	if !found {
		ent, err := tx.snapshot.GetEntity(hash)
		if err == nil {
			tx.cache.setEntityByHash(hash, ent)
		}
		return ent, err
	}
	return ent, nil
}

func (tx *transaction) getExtendedIndexByURI(uri turtle.URI) (storage.EntityExtendedIndex, error) {
	hash, err := tx.getHash(uri)
	if err != nil {
		return nil, err
	}
	return tx.getExtendedIndexByHash(hash)
}

func (tx *transaction) getExtendedIndexByHash(hash storage.HashKey) (storage.EntityExtendedIndex, error) {
	if tx.cache == nil {
		return tx.snapshot.GetExtendedIndex(hash)
	}
	ext, found := tx.cache.getExtendedIndexByHash(hash)
	if !found {
		ext, err := tx.snapshot.GetExtendedIndex(hash)
		if err == nil {
			tx.cache.setExtendedIndexByHash(hash, ext)
		}
		return ext, err
	}
	return ext, nil
}

func (tx *transaction) getPredicateByURI(uri turtle.URI) (storage.PredicateEntity, error) {
	hash, err := tx.getHash(uri)
	if err != nil {
		return nil, err
	}
	return tx.getPredicateByHash(hash)
}

func (tx *transaction) getPredicateByHash(hash storage.HashKey) (pred storage.PredicateEntity, err error) {
	if tx.cache == nil {
		pred, err = tx.snapshot.GetPredicate(hash)
		if err == storage.ErrNotFound {
			pred = storage.NewPredicateEntity(hash)
			err = tx.snapshot.PutPredicate(pred)
		}
		return pred, err
	}

	var found bool
	pred, found = tx.cache.getPredicateByHash(hash)
	if !found {
		pred, err = tx.snapshot.GetPredicate(hash)
		if err == nil {
			tx.cache.setPredicateByHash(hash, pred)
		} else if err == storage.ErrNotFound {
			pred = storage.NewPredicateEntity(hash)
			err = tx.snapshot.PutPredicate(pred)
		}
		return pred, err
	}
	return pred, nil
}

// takes the inverse of every relationship. If no inverse exists, returns nil
func (tx *transaction) reversePathPattern(path []sparql.PathPattern) []sparql.PathPattern {
	var reverse = make([]sparql.PathPattern, len(path))
	for idx, pred := range path {
		if inverse, found := tx.snapshot.GetReversePredicate(pred.Predicate); found {
			pred.Predicate = inverse
			reverse[idx] = pred
		} else {
			return nil
		}
	}
	return reversePath(reverse)
}

// follow the pattern from the given object's InEdges, placing the results in the btree
func (tx *transaction) followPathFromObject(object storage.Entity, results *keymap, searchstack *list.List, pattern sparql.PathPattern) error {
	stack := list.New()
	stack.PushFront(object)

	predHash, err := tx.getHash(pattern.Predicate)
	if err != nil && err == storage.ErrNotFound {
		logrus.Infof("Adding unseen predicate %s", pattern.Predicate)
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
		entity := stack.Remove(stack.Front()).(storage.Entity)
		if traversed.Has(entity.Key()) {
			continue
		}
		traversed.ReplaceOrInsert(entity.Key())
		switch pattern.Pattern {
		case sparql.PATTERN_SINGLE:
			// because this is one hop, we don't add any new entities to the stack
			// here, these entities are all connected by the required predicate
			for _, entityHash := range entity.ListInEndpoints(predHash) {
				results.Add(entityHash)
			}
		case sparql.PATTERN_ZERO_ONE:
			// because this is one hop, we don't add any new entities to the stack
			results.Add(entity.Key())
			// here, these entities are all connected by the required predicate
			for _, entityHash := range entity.ListInEndpoints(predHash) {
				results.Add(entityHash)
			}
		case sparql.PATTERN_ZERO_PLUS:
			results.Add(entity.Key())
			fallthrough

		case sparql.PATTERN_ONE_PLUS:
			index, err := tx.snapshot.GetExtendedIndex(entity.Key())
			if err != nil {
				return err
			}
			for _, entityHash := range index.ListInPlusEndpoints(predHash) {
				results.Add(entityHash)
			}

			// here, these entities are all connected by the required predicate
			for _, entityHash := range entity.ListInEndpoints(predHash) {
				nextEntity, err := tx.snapshot.GetEntity(entityHash)
				if err != nil {
					return err
				}
				results.Add(nextEntity.Key())
				searchstack.PushBack(nextEntity)
				// also make sure to add this to the stack so we can search
				stack.PushBack(nextEntity)
			}
		}
	}
	return nil
}

// follow the pattern from the given subject's OutEdges, placing the results in the btree
func (tx *transaction) followPathFromSubject(subject storage.Entity, results *keymap, searchstack *list.List, pattern sparql.PathPattern) error {
	stack := list.New()
	stack.PushFront(subject)

	predHash, err := tx.getHash(pattern.Predicate)
	if err != nil {
		return errors.Wrapf(err, "Not found: %v", pattern.Predicate)
	}

	var traversed = traversedBTreePool.Get()
	defer traversedBTreePool.Put(traversed)

	for stack.Len() > 0 {
		entity := stack.Remove(stack.Front()).(storage.Entity)
		if traversed.Has(entity.Key()) {
			continue
		}
		traversed.ReplaceOrInsert(entity.Key())
		switch pattern.Pattern {
		case sparql.PATTERN_SINGLE:
			// here, these entities are all connected by the required predicate
			for _, entityHash := range entity.ListOutEndpoints(predHash) {
				results.Add(entityHash)
			}
			// because this is one hop, we don't add any new entities to the stack
		case sparql.PATTERN_ZERO_ONE:
			// this does not require the pattern to exist, so we add the current entity plus any
			// connected by the appropriate edge
			results.Add(entity.Key())
			// here, these entities are all connected by the required predicate
			for _, entityHash := range entity.ListOutEndpoints(predHash) {
				results.Add(entityHash)
			}
			// because this is one hop, we don't add any new entities to the stack
		case sparql.PATTERN_ZERO_PLUS:
			results.Add(entity.Key())
			fallthrough

		case sparql.PATTERN_ONE_PLUS:
			index, err := tx.snapshot.GetExtendedIndex(entity.Key())
			if err != nil {
				return err
			}
			for _, entityHash := range index.ListOutPlusEndpoints(predHash) {
				results.Add(entityHash)
			}

			// here, these entities are all connected by the required predicate
			for _, entityHash := range entity.ListOutEndpoints(predHash) {
				nextEntity, err := tx.snapshot.GetEntity(entityHash)
				if err != nil {
					return err
				}
				if !results.Has(nextEntity.Key()) {
					searchstack.PushBack(nextEntity)
				}
				results.Add(entityHash)
				stack.PushBack(nextEntity)
			}
		}
	}
	return nil
}

func (tx *transaction) getSubjectFromPredObject(objectHash storage.HashKey, path []sparql.PathPattern) (*keymap, error) {
	// first get the initial object entity from the db
	// then we're going to conduct a BFS search starting from this entity looking for all entities
	// that have the required path sequence. We place the results in a BTree to maintain uniqueness

	// So how does this traversal actually work?
	// At each 'step', we are looking at an entity and some offset into the path.

	// get the object, look in its "in" edges for the path pattern
	objEntity, err := tx.getEntityByHash(objectHash)
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
			entity := stack.Remove(stack.Front()).(storage.Entity)
			// if we have already traversed this entity, skip it
			if traversed.Has(entity.Key()) {
				continue
			}
			// mark this entity as traversed
			traversed.ReplaceOrInsert(entity.Key())
			tx.followPathFromObject(entity, reachable, stack, segment)
		}

		// if we aren't done, then we push these items onto the stack
		if idx < len(path)-1 {
			reachable.Iter(func(key storage.HashKey) {
				ent, err := tx.getEntityByHash(key)
				if err != nil {
					logrus.Error(err)
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
func (tx *transaction) getObjectFromSubjectPred(subjectHash storage.HashKey, path []sparql.PathPattern) *keymap {
	subEntity, err := tx.getEntityByHash(subjectHash)
	if err != nil {
		logrus.Error(errors.Wrapf(err, "Not found: %v", subjectHash))
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
			entity := stack.Remove(stack.Front()).(storage.Entity)
			// if we have already traversed this entity, skip it
			if traversed.Has(entity.Key()) {
				continue
			}
			// mark this entity as traversed
			traversed.ReplaceOrInsert(entity.Key())
			tx.followPathFromSubject(entity, reachable, stack, segment)
		}

		// if we aren't done, then we push these items onto the stack
		if idx < len(path)-1 {
			reachable.Iter(func(key storage.HashKey) {
				ent, err := tx.getEntityByHash(key)
				if err != nil {
					logrus.Error(err)
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
func (tx *transaction) getSubjectObjectFromPred(path []sparql.PathPattern) (soPair [][]storage.HashKey, err error) {
	var pe storage.PredicateEntity
	logrus.Debug(path[0])
	pe, err = tx.getPredicateByURI(path[0].Predicate)
	if err != nil {
		err = errors.Wrapf(err, "Can't find predicate %v", path[0].Predicate)
		return
	}

	for _, subject := range pe.GetAllSubjects() {
		for _, object := range pe.GetObjects(subject) {
			soPair = append(soPair, []storage.HashKey{subject, object})
		}
	}
	return soPair, nil
}

func (tx *transaction) getPredicateFromSubjectObject(subject, object storage.Entity) *keymap {
	reachable := newKeymap()

	for _, pred := range subject.GetAllPredicates() {
		for _, edgeObject := range subject.ListInEndpoints(pred) {
			if edgeObject == object.Key() {
				reachable.Add(pred)
			}
		}
	}
	for _, pred := range subject.GetAllPredicates() {
		for _, edgeObject := range subject.ListOutEndpoints(pred) {
			if edgeObject == object.Key() {
				reachable.Add(pred)
			}
		}
	}

	return reachable
}

func (tx *transaction) getPredicatesFromObject(object storage.Entity) *keymap {
	reachable := newKeymap()

	for _, pred := range object.GetAllPredicates() {
		reachable.Add(pred)
	}

	return reachable
}

func (tx *transaction) getPredicatesFromSubject(subject storage.Entity) *keymap {
	reachable := newKeymap()
	for _, pred := range subject.GetAllPredicates() {
		reachable.Add(pred)
	}

	return reachable
}

func (tx *transaction) iterAllEntities(F func(storage.HashKey, storage.Entity) bool) error {
	return tx.snapshot.IterateAllEntities(F)
}
