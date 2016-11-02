package db

import (
	"encoding/binary"
	"strings"

	turtle "github.com/gtfierro/hod/goraptor"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/zhangxinngang/murmur"
)

type DB struct {
	// store []byte(entity URI) => primary key
	entityDB *leveldb.DB
	pkDB     *leveldb.DB
}

// type DataSet struct {
//	Namespaces  map[string]string
//	Triples     []Triple
// }
// type Triple struct {
// 	Subject   URI
// 	Predicate URI
// 	Object    URI
// }
// type URI struct {
// 	Namespace string
// 	Value     string
// }

func NewDB(path string) (*DB, error) {
	path = strings.TrimSuffix(path, "/")

	// set up entity, pk databases
	entityDBPath := path + "/db-entities"
	entityDB, err := leveldb.OpenFile(entityDBPath, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not open entityDB file %s", entityDBPath)
	}

	pkDBPath := path + "/db-pk"
	pkDB, err := leveldb.OpenFile(pkDBPath, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not open pkDB file %s", pkDBPath)
	}

	db := &DB{
		entityDB: entityDB,
		pkDB:     pkDB,
	}

	return db, nil
}

// hashes the given URI into the byte array
func (db *DB) hashURI(u turtle.URI, dest []byte, salt uint32) {
	var hash uint32
	if len(dest) < 4 {
		dest = make([]byte, 4)
	}
	if salt > 0 {
		saltbytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(saltbytes, salt)
		hash = murmur.Murmur3(append(u.Bytes(), saltbytes...))
	} else {
		hash = murmur.Murmur3(u.Bytes())
	}
	binary.LittleEndian.PutUint32(dest, hash)
}

// for each part of the triple (subject, predicate, object), we check if its already in the entity database.
// If it is, we can skip it. If not, we generate a murmur3 hash for the entity, and then
// 0. check if we've already inserted the entity (skip if we already have)
// 1. check if the hash is unique (check membership in pk db) - if it isn't then we add a salt and check again
// 2. insert hash => []byte(entity) into pk db
// 3. insert []byte(entity) => hash into entity db
func (db *DB) insertEntity(entity turtle.URI, hashdest []byte, enttx, pktx *leveldb.Transaction) error {
	// check if we've inserted Subject already
	if exists, err := enttx.Has(entity.Bytes(), nil); err == nil && exists {
		return nil
	} else if err != nil {
		return errors.Wrapf(err, "Error checking db membership for %s", entity.String())
	}
	// generate the hash
	var salt = uint32(0)
	db.hashURI(entity, hashdest, salt)
	for {
		if exists, err := pktx.Has(hashdest, nil); err == nil && exists {
			salt += 1
			db.hashURI(entity, hashdest, salt)
		} else if err != nil {
			return errors.Wrapf(err, "Error checking db membership for %v", hashdest)
		} else {
			break
		}
	}

	// insert the hash into the entity and prefix dbs
	if err := enttx.Put(entity.Bytes(), hashdest, nil); err != nil {
		return errors.Wrapf(err, "Error inserting entity %s", entity.String())
	}
	if err := pktx.Put(hashdest, entity.Bytes(), nil); err != nil {
		return errors.Wrapf(err, "Error inserting pk %s", hashdest)
	}
	return nil
}

func (db *DB) LoadDataset(dataset turtle.DataSet) error {
	// start transactions
	enttx, err := db.entityDB.OpenTransaction()
	if err != nil {
		return errors.Wrap(err, "Could not open transaction on entity dataset")
	}
	pktx, err := db.pkDB.OpenTransaction()
	if err != nil {
		return errors.Wrap(err, "Could not open transaction on pk dataset")
	}

	var hashdest = make([]byte, 4)
	for _, triple := range dataset.Triples {
		if err := db.insertEntity(triple.Subject, hashdest, enttx, pktx); err != nil {
			return err
		}
		if err := db.insertEntity(triple.Predicate, hashdest, enttx, pktx); err != nil {
			return err
		}
		if err := db.insertEntity(triple.Object, hashdest, enttx, pktx); err != nil {
			return err
		}

	}

	if err := enttx.Commit(); err != nil {
		return errors.Wrap(err, "Could not commit transaction")
	}
	if err := pktx.Commit(); err != nil {
		return errors.Wrap(err, "Could not commit transaction")
	}

	return nil
}
