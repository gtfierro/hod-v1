package db

import (
	turtle "github.com/gtfierro/hod/goraptor"
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

		// make subject entity
		subjHash, err := db.GetHash(triple.Subject)
		if err != nil {
			return err
		}
		// check if entity exists
		if exists, err := db.graphDB.Has(subjHash[:], nil); err == nil && !exists {
			// if not exists, create a new entity and insert it
			subEnt := NewEntity()
			subEnt.PK = subjHash
			bytes, err := subEnt.MarshalMsg(nil)
			if err != nil {
				return err
			}
			if err := db.graphDB.Put(subjHash[:], bytes, nil); err != nil {
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
		if exists, err := db.graphDB.Has(objHash[:], nil); err == nil && !exists {
			// if not exists, create a new entity and insert it
			objEnt := NewEntity()
			objEnt.PK = objHash
			bytes, err := objEnt.MarshalMsg(nil)
			if err != nil {
				return err
			}
			if err := db.graphDB.Put(objHash[:], bytes, nil); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	// second pass
	for _, triple := range dataset.Triples {
		subject, err := db.GetEntity(triple.Subject)
		if err != nil {
			return err
		}
		object, err := db.GetEntity(triple.Object)
		if err != nil {
			return err
		}

		// add the forward edge
		predHash := predicates[triple.Predicate.String()]
		subject.AddOutEdge(predHash, object.PK)
		object.AddInEdge(predHash, subject.PK)

		// find the inverse edge
		reverseEdge, found := db.relationships[triple.Predicate]
		// if an inverse edge exists, then we add it to the object
		if found {
			reverseEdgeHash := predicates[reverseEdge.String()]
			object.AddOutEdge(reverseEdgeHash, subject.PK)
			subject.AddInEdge(reverseEdgeHash, object.PK)
		}

		// re-put in graph
		bytes, err := subject.MarshalMsg(nil)
		if err != nil {
			return err
		}
		if err := db.graphDB.Put(subject.PK[:], bytes, nil); err != nil {
			return err
		}

		bytes, err = object.MarshalMsg(nil)
		if err != nil {
			return err
		}
		if err := db.graphDB.Put(object.PK[:], bytes, nil); err != nil {
			return err
		}

	}

	return nil
}
