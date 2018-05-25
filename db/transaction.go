package db

import (
	"container/list"
	"fmt"
	sparql "github.com/gtfierro/hod/lang/ast"
	"github.com/gtfierro/hod/turtle"
	"github.com/pkg/errors"
	logrus "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"time"
)

const (
	RDF_NAMESPACE = "http://www.w3.org/1999/02/22-rdf-syntax-ns"
	OWL_NAMESPACE = "http://www.w3.org/2002/07/owl"
)

var (
	INVERSEOF = turtle.URI{
		Namespace: OWL_NAMESPACE,
		Value:     "inverseOf",
	}
)

// wrapper around the internal k/v store transaction
type transaction struct {
	entity               *leveldb.Transaction
	pk                   *leveldb.Transaction
	graph                *leveldb.Transaction
	ext                  *leveldb.Transaction
	pred                 *leveldb.Transaction
	triplesAdded         int
	hashes               map[turtle.URI]Key
	inverseRelationships map[Key]Key
	t                    *traversal
	cache                *dbcache
}

func (db *DB) openTransaction() (tx *transaction, err error) {
	tx = &transaction{
		hashes:               make(map[turtle.URI]Key),
		inverseRelationships: make(map[Key]Key),
		cache:                db.cache,
	}
	t := &traversal{under: tx}
	tx.t = t
	getTransaction := func(db *leveldb.DB) (*leveldb.Transaction, error) {
		if ltx, err := db.OpenTransaction(); err != nil {
			if tx.entity != nil {
				tx.entity.Discard()
			}
			if tx.pk != nil {
				tx.pk.Discard()
			}
			if tx.graph != nil {
				tx.graph.Discard()
			}
			if tx.ext != nil {
				tx.ext.Discard()
			}
			if tx.pred != nil {
				tx.pred.Discard()
			}
			return nil, err
		} else {
			return ltx, err
		}
	}
	if tx.entity, err = getTransaction(db.entityDB); err != nil {
		return
	}
	if tx.pk, err = getTransaction(db.pkDB); err != nil {
		return
	}
	if tx.graph, err = getTransaction(db.graphDB); err != nil {
		return
	}
	if tx.ext, err = getTransaction(db.extendedDB); err != nil {
		return
	}
	if tx.pred, err = getTransaction(db.predDB); err != nil {
		return
	}
	return
}

func (tx *transaction) discard() {
	tx.entity.Discard()
	tx.pk.Discard()
	tx.graph.Discard()
	tx.ext.Discard()
	tx.pred.Discard()
}

func (tx *transaction) commit() error {
	if err := tx.entity.Commit(); err != nil {
		tx.discard()
		return err
	}
	if err := tx.pk.Commit(); err != nil {
		tx.discard()
		return err
	}
	if err := tx.graph.Commit(); err != nil {
		tx.discard()
		return err
	}
	if err := tx.ext.Commit(); err != nil {
		tx.discard()
		return err
	}
	if err := tx.pred.Commit(); err != nil {
		tx.discard()
		return err
	}
	return nil
}

func (tx *transaction) done() error {
	return tx.commit()
}

func (tx *transaction) getHash(uri turtle.URI) (Key, error) {
	var ret Key
	val, err := tx.entity.Get(uri.Bytes(), nil)
	if err != nil {
		return ret, fmt.Errorf("Got non-existent hash but it should exist for %s", uri)
	}
	copy(ret[:], val)
	return ret, nil
}

func (tx *transaction) getURI(hash Key) (turtle.URI, error) {
	if hash == emptyKey {
		return turtle.URI{}, nil
	}
	val, err := tx.pk.Get(hash[:], nil)
	if err != nil {
		return turtle.URI{}, errors.Wrapf(err, "Could not get URI for %v", hash)
	}
	uri := turtle.ParseURI(string(val))
	if err != nil {
		return turtle.URI{}, errors.Wrapf(err, "Could not get URI for %v", hash)
	}
	return uri, nil
}

func (tx *transaction) putEntity(ent *Entity) error {
	bytes, err := ent.MarshalMsg(nil)
	if err != nil {
		return err
	}
	return tx.graph.Put(ent.PK[:], bytes, nil)
}

func (tx *transaction) getEntityByURI(uri turtle.URI) (*Entity, error) {
	hash, err := tx.getHash(uri)
	if err != nil {
		return nil, err
	}
	return tx.getEntityByHash(hash)
}

func (tx *transaction) getEntityByHash(hash Key) (*Entity, error) {
	var entity = NewEntity()
	bytes, err := tx.graph.Get(hash[:], nil)
	if err != nil && err != leveldb.ErrNotFound {
		return nil, errors.Wrap(err, "Error getting entity from transaction")
	}
	_, err = entity.UnmarshalMsg(bytes)
	if err != nil {
		return nil, errors.Wrap(err, "Error deserializing entity from transaction")
	}
	return entity, nil
}

func (tx *transaction) getExtendedIndexByHash(hash Key) (*EntityExtendedIndex, error) {
	bytes, err := tx.ext.Get(hash[:], nil)
	if err != nil {
		return nil, err
	}
	ent := NewEntityExtendedIndex()
	_, err = ent.UnmarshalMsg(bytes)
	return ent, err
}

func (tx *transaction) getExtendedIndexByURI(uri turtle.URI) (*EntityExtendedIndex, error) {
	hash, err := tx.getHash(uri)
	if err != nil {
		return nil, err
	}
	return tx.getExtendedIndexByHash(hash)
}

func (tx *transaction) saveExtendedIndex(index *EntityExtendedIndex) error {
	if bytes, err := index.MarshalMsg(nil); err != nil {
		return errors.Wrap(err, "Error serializing extended index from transaction")
	} else if err := tx.ext.Put(index.PK[:], bytes, nil); err != nil {
		return errors.Wrap(err, "Error inserting extended index in transaction")
	}
	return nil
}

func (tx *transaction) getPredicateByURI(uri turtle.URI) (*PredicateEntity, error) {
	hash, err := tx.getHash(uri)
	if err != nil {
		return nil, err
	}
	return tx.getPredicateByHash(hash)
}

func (tx *transaction) getPredicateByHash(hash Key) (*PredicateEntity, error) {
	var pred = NewPredicateEntity()
	bytes, err := tx.pred.Get(hash[:], nil)
	if err != nil && err != leveldb.ErrNotFound {
		return nil, errors.Wrap(err, "Error getting predicate from transaction")
	} else if err == leveldb.ErrNotFound {
		// add predicate entity to predhash db
		pred.PK = hash
		return pred, tx.savePredicate(pred)
	} else {
		// load predicate entity from db
		_, err = pred.UnmarshalMsg(bytes)
		return pred, err
	}
}

func (tx *transaction) savePredicate(pred *PredicateEntity) error {
	if bytes, err := pred.MarshalMsg(nil); err != nil {
		return errors.Wrap(err, "Error serializing predicate from transaction")
	} else if err := tx.pred.Put(pred.PK[:], bytes, nil); err != nil {
		return errors.Wrap(err, "Error inserting predicate in transaction")
	}
	return nil
}

func (tx *transaction) addTriples(dataset turtle.DataSet) error {
	var newPredicates = make(map[Key]struct{})

	addStart := time.Now()
	// add all URIs to the database
	for _, triple := range dataset.Triples {
		if err := tx.addTriple(triple); err != nil {
			return errors.Wrapf(err, "Could not load triple (%s)", triple)
		}
		newPredicates[tx.hashes[triple.Predicate]] = struct{}{} // mark new predicate

		// if triple defines an inverseOf relationship, then track the subject/object of that
		// triple so we can populate the graph later
		if triple.Predicate.Namespace == OWL_NAMESPACE && triple.Predicate.Value == "inverseOf" {
			subjectHash := tx.hashes[triple.Subject]
			objectHash := tx.hashes[triple.Object]
			tx.inverseRelationships[subjectHash] = objectHash
			tx.inverseRelationships[objectHash] = subjectHash
			tx.getPredicateByURI(triple.Subject)
			tx.getPredicateByURI(triple.Object)
		}
		tx.triplesAdded += 1
	}
	addEnd := time.Now()

	// pull out all of the inverse edges from the database and add to inverseRelationships
	reverseEdgeFindStart := time.Now()
	var predicatesAdded int
	pred, err := tx.getPredicateByURI(INVERSEOF)
	if err != nil && err != leveldb.ErrNotFound {
		logrus.WithError(err).Error("Could not load INVERSEOF pred")
	} else if err == nil {
		for subject, objectMap := range pred.Subjects {
			for object := range objectMap {
				var sh, oh Key
				sh.FromSlice([]byte(subject))
				oh.FromSlice([]byte(object))
				tx.inverseRelationships[sh] = oh
				tx.inverseRelationships[oh] = sh
				predicatesAdded += 1
			}
		}
	}

	reverseEdgeFindEnd := time.Now()

	// add the inverse edges to the graph index
	reverseEdgeBuildStart := time.Now()
	var subject, object Key
	for predicate, reversePredicate := range tx.inverseRelationships {
		pred, err := tx.getPredicateByHash(predicate)
		_uri, _ := tx.getURI(predicate)
		if err != nil {
			return errors.Wrapf(err, "Could not load predicate %s", _uri)
		}

		for subjectStr, objectMap := range pred.Subjects {
			subject.FromSlice([]byte(subjectStr))
			for objectStr := range objectMap {
				object.FromSlice([]byte(objectStr))
				subjectEnt, err := tx.getEntityByHash(subject)
				if err != nil {
					return errors.Wrap(err, "Could not load subject")
				}
				objectEnt, err := tx.getEntityByHash(object)
				if err != nil {
					return errors.Wrap(err, "Could not load object")
				}
				subjectEnt.AddInEdge(reversePredicate, object)
				objectEnt.AddOutEdge(reversePredicate, subject)
				if err = tx.putEntity(subjectEnt); err != nil {
					return err
				}
				if err = tx.putEntity(objectEnt); err != nil {
					return err
				}
			}
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

	// TODO: for all *new* entities, roll the edges forward for all predicates
	//for _, turtle := range dataset.Triples {
	//}

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

// things to do:
// - check for inverseOf relationships (do as a second pass?) and mark these for reverse edges
// - track the namespaces we find
// - add the class name to the text index
// - add reverse edges to the graph
// - populate predicate index
//
// for each part of the triple (subject, predicate, object), we check if its already in the entity database.
// If it is, we can skip it. If not, we generate a murmur3 hash for the entity, and then
// 0. check if we've already inserted the entity (skip if we already have)
// 1. check if the hash is unique (check membership in pk db) - if it isn't then we add a salt and check again
// 2. insert hash => []byte(entity) into pk db
// 3. insert []byte(entity) => hash into entity db
func (tx *transaction) addTriple(triple turtle.Triple) error {
	// insert subject, predicate and object
	tx.addURI(triple.Subject)
	tx.addURI(triple.Predicate)
	tx.addURI(triple.Object)

	// add the "1 or more" edge for the extended index
	rev := triple.Predicate
	rev.Value += "+"
	tx.addURI(rev)

	// populate subject, predicate and object in graph index with forward/inverse edges
	var (
		subjectHash   = tx.hashes[triple.Subject]
		predicateHash = tx.hashes[triple.Predicate]
		objectHash    = tx.hashes[triple.Object]
	)

	pred, err := tx.getPredicateByURI(triple.Predicate)
	if err != nil {
		return err
	}
	if pred.AddSubjectObject(subjectHash, objectHash) {
		if err := tx.savePredicate(pred); err != nil {
			return err
		}
	}

	subject, err := tx.getEntityByURI(triple.Subject)
	if err != nil {
		return err
	}
	object, err := tx.getEntityByURI(triple.Object)
	if err != nil {
		return err
	}
	if subject.AddOutEdge(predicateHash, object.PK) {
		if err = tx.putEntity(subject); err != nil {
			return err
		}
	}
	if object.AddInEdge(predicateHash, subject.PK) {
		if err = tx.putEntity(object); err != nil {
			return err
		}
	}

	tx.cache.evict(subjectHash)
	tx.cache.evict(objectHash)
	tx.cache.evict(predicateHash)

	return nil
}

// add the URI to the transaction. This involves:
// - compute the hash of the entity
// - add the entity and its hash to the entity/pk dbs
// - initialize the entity's "neighbor table" in the graph db if it doesn't exist yet
func (tx *transaction) addURI(uri turtle.URI) error {
	var hashdest Key
	var found bool

	if hashdest, found = tx.hashes[uri]; !found {
		if _hashdest, err := tx.entity.Get(uri.Bytes(), nil); err != nil && err != leveldb.ErrNotFound {
			return errors.Wrap(err, "Could not check key existence")
		} else if err == nil {
			copy(hashdest[:], _hashdest)
		} else if err == leveldb.ErrNotFound {
			// else if not found, then generate it
			var salt = uint64(0)
			hashURI(uri, hashdest[:], salt)
			for {
				if exists, err := tx.pk.Has(hashdest[:], nil); err == nil && exists {
					log.Warning("hash exists", uri)
					salt += 1
					hashURI(uri, hashdest[:], salt)
				} else if err != nil {
					return errors.Wrapf(err, "Error checking db membership for %v", hashdest)
				} else {
					break
				}
			}
			// insert the hash into the entity and prefix dbs
			if err := tx.entity.Put(uri.Bytes(), hashdest[:], nil); err != nil {
				return errors.Wrapf(err, "Error inserting uri %s", uri.String())
			}
			if err := tx.pk.Put(hashdest[:], uri.Bytes(), nil); err != nil {
				return errors.Wrapf(err, "Error inserting pk %s", hashdest)
			}
		}
		tx.hashes[uri] = hashdest
	}

	// insert the hash into the graph index if it doesn't exist already
	if exists, err := tx.graph.Has(hashdest[:], nil); err == nil && !exists {
		ent := NewEntity()
		ent.PK = hashdest
		if bytes, err := ent.MarshalMsg(nil); err != nil {
			return err
		} else if err := tx.graph.Put(hashdest[:], bytes, nil); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

// follows all paths
func (tx *transaction) rollupPredicate(predicateHash Key) error {
	var err error
	forwardPath := sparql.PathPattern{Pattern: sparql.PATTERN_ONE_PLUS}
	results := newKeymap()
	forwardPath.Predicate, err = tx.getURI(predicateHash)
	if err != nil {
		return err
	}
	predicate, err := tx.getPredicateByHash(predicateHash)
	if err != nil {
		return err
	}
	for subjectStringHash := range predicate.Subjects {
		var subjectHash Key
		subjectHash.FromSlice([]byte(subjectStringHash))
		if exists, err := tx.ext.Has(subjectHash[:], nil); err == nil && !exists {
			subjectIndex := NewEntityExtendedIndex()
			subjectIndex.PK = subjectHash
			bytes, err := subjectIndex.MarshalMsg(nil)
			if err != nil {
				return err
			}
			if err := tx.ext.Put(subjectHash[:], bytes, nil); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		subjectIndex, err := tx.getExtendedIndexByHash(subjectHash)
		if err != nil {
			return err
		}
		subject, err := tx.getEntityByHash(subjectHash)
		if err != nil {
			return err
		}

		stack := list.New()
		tx.t.followPathFromSubject(subject, results, stack, forwardPath)
		for results.Len() > 0 {
			objectIndex, err := tx.getExtendedIndexByHash(results.Max())
			if err != nil {
				return err
			}
			subjectIndex.AddOutPlusEdge(predicateHash, results.DeleteMax())
			objectIndex.AddInPlusEdge(predicateHash, subjectHash)
			if err := tx.saveExtendedIndex(objectIndex); err != nil {
				return err
			}
		}
		if err := tx.saveExtendedIndex(subjectIndex); err != nil {
			return err
		}
	}

	for objectStringHash := range predicate.Objects {
		var objectHash Key
		objectHash.FromSlice([]byte(objectStringHash))
		if exists, err := tx.ext.Has(objectHash[:], nil); err == nil && !exists {
			objectIndex := NewEntityExtendedIndex()
			objectIndex.PK = objectHash
			if err := tx.saveExtendedIndex(objectIndex); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		objectIndex, err := tx.getExtendedIndexByHash(objectHash)
		if err != nil {
			return err
		}
		object, err := tx.getEntityByHash(objectHash)
		if err != nil {
			return err
		}

		stack := list.New()
		tx.t.followPathFromObject(object, results, stack, forwardPath)
		for results.Len() > 0 {
			subjectIndex, err := tx.getExtendedIndexByHash(results.Max())
			if err != nil {
				return err
			}
			objectIndex.AddInPlusEdge(predicateHash, results.DeleteMax())
			subjectIndex.AddOutPlusEdge(predicateHash, objectHash)
			if err := tx.saveExtendedIndex(subjectIndex); err != nil {
				return err
			}
		}
		if err := tx.saveExtendedIndex(objectIndex); err != nil {
			return err
		}
	}
	return nil
}

func (tx *transaction) getReverseRelationship(forward turtle.URI) (reverse turtle.URI, found bool) {
	var (
		forwardHash, reverseHash Key
		err                      error
	)
	forwardHash, err = tx.getHash(forward)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"predicate": forward,
			"err":       err,
		}).Error("transaction")
		found = false
		return
	}
	if reverseHash, found = tx.inverseRelationships[forwardHash]; !found {
		return
	} else if reverse, err = tx.getURI(reverseHash); err != nil {
		logrus.WithFields(logrus.Fields{
			"predicate": forward,
			"err":       err,
		}).Error("transaction")
		found = false
		return
	}
	return
}
