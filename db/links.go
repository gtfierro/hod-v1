package db

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gtfierro/hod/config"
	"github.com/gtfierro/hod/goraptor"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// TODO: integrate this in with the db. Two pars:
// TODO: - add ability to set links from API call
// TODO: - hook up fetching links with query evaluation

var ErrKeyTooLong = errors.New("Key is more than 12 bytes")

const MaxKeyLength = 12
const MaxValLength = 128

type Link struct {
	URI    turtle.URI
	entity [4]byte
	Key    []byte
	Value  []byte
}

func (l Link) Valid() bool {
	return l.entity != emptyHash && len(l.Key) <= MaxKeyLength && len(l.Value) <= MaxValLength
}

// we deserialize from a structure that looks like:
//   {
//    ex:temp-sensor-1:
//       {
//           URI: ucberkeley/eecs/soda/sensors/etcetc/1,
//           UUID: abcdef,
//       },
//    ex:temp-sensor-2:
//       {
//           URI: ucberkeley/eecs/soda/sensors/etcetc/2,
//           UUID: ghijkl,
//       }
//   }
// To "delete" links, leave the key empty, e.g.
//   ex:temp-sensor-1: {
//     URI: "",
//   }
// To clear all links for an entity, leave its entry empty, e.g.
//   ex:temp-sensor-1: {}
type LinkUpdates struct {
	Adding   []*Link
	Removing []*Link
}

func (updates *LinkUpdates) UnmarshalJSON(b []byte) (err error) {
	// intermediate form for the JSON
	var intermediate map[string]interface{}
	if err := json.Unmarshal(b, &intermediate); err != nil {
		return errors.Wrap(err, "Could not unmarshal LinkUpdates json (was not a map?)")
	}
	for uri, res := range intermediate {
		if res == nil {
			updates.Removing = append(updates.Removing, &Link{URI: turtle.ParseURI(uri)})
			continue
		}
		uri_updates, ok := res.(map[string]interface{})
		if !ok {
			return errors.New("Could not unmarshal LinkUpdates json (invalid entry for URI)")
		}
		if len(uri_updates) == 0 {
			updates.Removing = append(updates.Removing, &Link{URI: turtle.ParseURI(uri)})
			continue
		}
		for key, val := range uri_updates {
			if val == nil {
				updates.Removing = append(updates.Removing,
					&Link{URI: turtle.ParseURI(uri),
						Key: []byte(key)})
			}
			str, ok := val.(string)
			if !ok {
				str = fmt.Sprintf("%s", val) // coerce to string
			}
			if len(str) == 0 {
				updates.Removing = append(updates.Removing,
					&Link{URI: turtle.ParseURI(uri),
						Key: []byte(key)})
			} else {
				updates.Adding = append(updates.Adding,
					&Link{URI: turtle.ParseURI(uri),
						Key:   []byte(key),
						Value: []byte(str),
					})
			}
		}
	}
	return nil
}

//TODO: need json deserialize for link updates so that they can be loaded in from the API (both
// command line loading and a POST via the server

// this database stores "links", which are key-value pairs that connect
// Brick nodes with other virtual resources. Links will be UUIDs for timeseries
// streams, URIs for BOSSWAVE subscription or web interfaces, timestamps for when the
// key was added, etc
// The structure of the database will be as follows:
//  - Key: [ 4-byte entity PK | key bytes ] (concatenation)
//      Keys can have a maximum length of 12 bytes
//  - Value: [ value bytes ]
//      Values have a maximum length of 128 bytes
type linkDB struct {
	hod *DB
	db  *leveldb.DB
}

func newLinkDB(hod *DB, cfg *config.Config) (*linkDB, error) {
	path := strings.TrimSuffix(cfg.DBPath, "/")
	linkDBPath := path + "/db-links"
	ldb, err := leveldb.OpenFile(linkDBPath, &opt.Options{
		Filter: filter.NewBloomFilter(32),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Could not open linkDB file %s", linkDBPath)
	}

	db := &linkDB{
		hod: hod,
		db:  ldb,
	}

	return db, nil
}

// copies the entity and key bytes into 'dest' to act as the access key for
// the underlying leveldb
func getlinkdbkey(entity [4]byte, key []byte, dest *[64]byte) {
	// TODO: switch this to uri/key, not entity. Use this method to retrieve entity PK
	copy(dest[:], entity[:])
	copy(dest[4:], key)
}

func (ldb *linkDB) getKey(l *Link, dest *[64]byte) {
	hash := ldb.hod.MustGetHash(l.URI)
	l.entity = hash
	copy(dest[:], hash[:])
	copy(dest[4:], l.Key)
}

func (ldb *linkDB) startTx() (*leveldb.Transaction, error) {
	return ldb.db.OpenTransaction()
}

func (ldb *linkDB) commitTx(tx *leveldb.Transaction) error {
	return tx.Commit()
}

func (ldb *linkDB) get(link *Link) (value []byte, err error) {
	var fetchKey [64]byte
	var exists bool
	if len(link.Key) > MaxKeyLength {
		err = ErrKeyTooLong
		return
	}
	ldb.getKey(link, &fetchKey)
	if exists, err = ldb.db.Has(fetchKey[:], nil); err == nil && exists {
		value, err = ldb.db.Get(fetchKey[:], nil)
		return
	} else if err != nil {
		return
	}
	return
}

func (ldb *linkDB) getAll(entity [4]byte) (keys [][]byte, values [][]byte, err error) {
	var start, limit [16]byte
	copy(start[:], entity[:])
	for i := 4; i < 16; i++ {
		start[i] = 0
	}
	copy(limit[:], entity[:])
	for i := 4; i < 16; i++ {
		limit[i] = 0xFF
	}
	keyrange := &util.Range{Start: start[:], Limit: limit[:]}
	iter := ldb.db.NewIterator(keyrange, nil)
	for iter.Next() {
		keys = append(keys, iter.Key())
		values = append(values, iter.Value())
	}
	iter.Release()
	err = iter.Error()
	return
}

func (ldb *linkDB) set(tx *leveldb.Transaction, link *Link) (err error) {
	var fetchKey [64]byte
	if len(link.Key) > MaxKeyLength {
		err = ErrKeyTooLong
		return
	}
	ldb.getKey(link, &fetchKey)
	err = tx.Put(fetchKey[:], link.Value, nil)
	return
}

func (ldb *linkDB) delete(tx *leveldb.Transaction, link *Link) (err error) {
	var fetchKey [64]byte
	if len(link.Key) > MaxKeyLength {
		err = ErrKeyTooLong
		return
	}
	ldb.getKey(link, &fetchKey)
	err = tx.Delete(fetchKey[:], nil)
	return
}
