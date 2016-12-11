package db

import (
	turtle "github.com/gtfierro/hod/goraptor"
	"github.com/pkg/errors"
)

// make sure to call this after we've populated the entity and publickey databases
// via db.LoadDataset
// This function builds the graph structure inside another leveldb kv store
// This is done in several passes (which we can optimize later):
//
// First pass:
//  - loop through all the triples and add the entities to the graph kv
//  - during this, we:
//	  - make a small local cache of predicateBytes => uint32 hash
//    - allocate an entity for both the subject AND object of a triple and add those
//      if they are not already added.
//		Make sure to use the entity/pk databases to look up their hashes (db.GetHash)
// Second pass:
//  - fill in all of the edges in the graph
func (db *DB) buildGraph(dataset turtle.DataSet) error {
	var predicates = make(map[string][4]byte)
	var subjAdded = 0
	var objAdded = 0
	graphtx, err := db.graphDB.OpenTransaction()
	if err != nil {
		return errors.Wrap(err, "Could not open transaction on graph dataset")
	}
	// first pass
	for _, triple := range dataset.Triples {
		// populate predicate cache
		if _, found := predicates[triple.Predicate.String()]; !found {
			predHash, err := db.GetHash(triple.Predicate)
			if err != nil {
				return err
			}
			predicates[triple.Predicate.String()] = predHash
		}
		if reversePredicate, hasReverse := db.relationships[triple.Predicate]; hasReverse {
			if _, found := predicates[reversePredicate.String()]; !found {
				predHash, err := db.GetHash(reversePredicate)
				if err != nil {
					return err
				}
				predicates[reversePredicate.String()] = predHash
			}
		}

		// make subject entity
		subjHash, err := db.GetHash(triple.Subject)
		if err != nil {
			return err
		}
		// check if entity exists
		if exists, err := graphtx.Has(subjHash[:], nil); err == nil && !exists {
			// if not exists, create a new entity and insert it
			subjAdded += 1
			subEnt := NewEntity()
			subEnt.PK = subjHash
			bytes, err := subEnt.MarshalMsg(nil)
			if err != nil {
				return err
			}
			if err := graphtx.Put(subjHash[:], bytes, nil); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		// make object entity
		objHash, err := db.GetHash(triple.Object)
		if err != nil {
			return err
		}
		// check if entity exists
		if exists, err := graphtx.Has(objHash[:], nil); err == nil && !exists {
			// if not exists, create a new entity and insert it
			objAdded += 1
			objEnt := NewEntity()
			objEnt.PK = objHash
			bytes, err := objEnt.MarshalMsg(nil)
			if err != nil {
				return err
			}
			if err := graphtx.Put(objHash[:], bytes, nil); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	if err := graphtx.Commit(); err != nil {
		return errors.Wrap(err, "Could not commit transaction")
	}

	log.Errorf("ADDED subjects %d, objects %d", subjAdded, objAdded)

	graphtx, err = db.graphDB.OpenTransaction()
	if err != nil {
		return errors.Wrap(err, "Could not open transaction on graph dataset")
	}

	// second pass
	for _, triple := range dataset.Triples {
		subject, err := db.GetEntityTx(graphtx, triple.Subject)
		if err != nil {
			return err
		}
		object, err := db.GetEntityTx(graphtx, triple.Object)
		if err != nil {
			return err
		}

		// add the forward edge
		predHash := predicates[triple.Predicate.String()]
		subject.AddOutEdge(predHash, object.PK)
		object.AddInEdge(predHash, subject.PK)

		// find the inverse edge
		reverseEdge, hasReverseEdge := db.relationships[triple.Predicate]
		// if an inverse edge exists, then we add it to the object
		if hasReverseEdge {
			reverseEdgeHash := predicates[reverseEdge.String()]
			object.AddOutEdge(reverseEdgeHash, subject.PK)
			subject.AddInEdge(reverseEdgeHash, object.PK)
		}

		// re-put in graph
		bytes, err := subject.MarshalMsg(nil)
		if err != nil {
			return err
		}
		if err := graphtx.Put(subject.PK[:], bytes, nil); err != nil {
			return err
		}

		bytes, err = object.MarshalMsg(nil)
		if err != nil {
			return err
		}
		if err := graphtx.Put(object.PK[:], bytes, nil); err != nil {
			return err
		}
	}
	if err = graphtx.Commit(); err != nil {
		return errors.Wrap(err, "Could not commit transaction")
	}

	return nil
}
