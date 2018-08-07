package storage

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/gtfierro/hod/config"
	"github.com/gtfierro/hod/turtle"
	"github.com/pkg/errors"
	logrus "github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, ForceColors: true})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
}

type BadgerStorageProvider struct {
	cfg     *config.Config
	basedir string
	dbs     map[Version]*badger.DB
	vm      *VersionManager
	sync.RWMutex
}

func (bsp *BadgerStorageProvider) Initialize(cfg *config.Config) error {
	dir := "badger" //, err := "badger" //ioutil.TempDir("", "badger")
	//	if err != nil {
	//		return err
	//	}
	bsp.basedir = dir
	if err := os.MkdirAll(bsp.basedir, 0700); err != nil {
		return err
	}
	bsp.dbs = make(map[Version]*badger.DB)
	if vm, err := CreateVersionManager(bsp.basedir); err != nil {
		return err
	} else {
		bsp.vm = vm
	}

	if versions, err := bsp.vm.Graphs(); err != nil {
		return err
	} else {
		for _, version := range versions {
			opts := badger.DefaultOptions
			dir := filepath.Join(bsp.basedir, version.Name, strconv.Itoa(int(version.Timestamp)))
			if err = os.MkdirAll(dir, 0700); err != nil {
				return err
			}
			opts.Dir = dir
			opts.ValueDir = dir
			if bsp.dbs[version], err = badger.Open(opts); err != nil {
				return err
			}
		}
	}
	return nil
}

func (bsp *BadgerStorageProvider) Close() error {
	bsp.Lock()
	defer bsp.Unlock()
	for _, db := range bsp.dbs {
		db.Close()
	}
	bsp.vm.db.Close()
	return os.RemoveAll(bsp.basedir)
}

func (bsp *BadgerStorageProvider) AddGraph(name string) (version Version, exists bool, err error) {
	logrus.Info("Add graph", name)
	exists = false
	bsp.RLock()
	for ver := range bsp.dbs {
		if ver.Name == name {
			exists = true
			version = ver
		}
	}
	bsp.RUnlock()

	if exists {
		logrus.Info(version)
		return
	}

	opts := badger.DefaultOptions
	dir := filepath.Join(bsp.basedir, name)
	if err = os.MkdirAll(dir, 0700); err != nil {
		return
	}

	bsp.Lock()
	defer bsp.Unlock()
	opts.Dir = dir
	opts.ValueDir = dir
	version = Version{uint64(time.Now().UnixNano()), name}
	if bsp.dbs[version], err = badger.Open(opts); err != nil {
		return
	}
	return version, exists, bsp.vm.AddVersion(version)
}

func (bsp *BadgerStorageProvider) CurrentVersion(name string) (version Version, err error) {
	return bsp.vm.GetLatestVersion(name)
}

func (bsp *BadgerStorageProvider) Graphs() ([]Version, error) {
	return bsp.vm.Graphs()
}

func (bsp *BadgerStorageProvider) CreateVersion(name string) (tx Transaction, err error) {
	// get current version for given name
	var (
		db    *badger.DB
		found bool
	)
	bsp.Lock()
	current, err := bsp.CurrentVersion(name)
	if err != nil {
		bsp.Unlock()
		return nil, err
	}
	if db, found = bsp.dbs[current]; !found {
		bsp.Unlock()
		return nil, ErrGraphNotFound
	}

	bsp.Unlock()
	tx = &BadgerGraphTx{
		&BadgerGraph{
			db:      db,
			tx:      db.NewTransaction(true),
			inverse: make(map[turtle.URI]turtle.URI),
		},
	}

	return
}
func (bsp *BadgerStorageProvider) OpenVersion(ver Version) (tx Transaction, err error) {
	// get current version for given name
	var (
		db    *badger.DB
		found bool
	)
	bsp.Lock()
	if db, found = bsp.dbs[ver]; !found {
		bsp.Unlock()
		err = ErrGraphNotFound
		return
	}
	bsp.Unlock()
	tx = &BadgerGraphTx{
		&BadgerGraph{
			version: ver,
			db:      db,
			tx:      db.NewTransaction(false), // read-only
			inverse: make(map[turtle.URI]turtle.URI),
		},
	}

	return
}

func (bsp *BadgerStorageProvider) ListVersions(name string) (versions []Version, err error) {
	return bsp.vm.ListVersions(name)
}

type BadgerGraph struct {
	version Version
	db      *badger.DB
	tx      *badger.Txn
	inverse map[turtle.URI]turtle.URI
}

func (bg *BadgerGraph) GetHash(uri turtle.URI) (hash HashKey, rerr error) {
	var uri_bytes = uri.Bytes()
	if item, err := bg.tx.Get(uri_bytes); err == badger.ErrKeyNotFound {
		rerr = ErrNotFound
		return
	} else if err != nil {
		rerr = err
		return
	} else {
		_, rerr = item.ValueCopy(hash[:])
		return
	}
}

func (bg *BadgerGraph) GetURI(hash HashKey) (turtle.URI, error) {
	hash = hash.AsType(PK)

	if item, err := bg.tx.Get(hash[:]); err == badger.ErrKeyNotFound {
		return turtle.URI{}, ErrNotFound
	} else if value, err := item.Value(); err != nil {
		return turtle.URI{}, err
	} else {
		return turtle.ParseURI(string(value)), err
	}
}

func (bg *BadgerGraph) Release() {
	bg.tx.Discard()
	return
}

func (bg *BadgerGraph) Commit() error {
	return nil
}

func (bg *BadgerGraph) Version() Version {
	return bg.version
}

func (bg *BadgerGraph) GetReversePredicate(rel turtle.URI) (inverse turtle.URI, found bool) {
	inverse, found = bg.inverse[rel]
	return
}

func (bg *BadgerGraph) GetEntity(hash HashKey) (ent Entity, rerr error) {
	hash = hash.AsType(ENTITY)
	item, err := bg.tx.Get(hash[:])
	if err == badger.ErrKeyNotFound {
		rerr = ErrNotFound
		return
	} else if err != nil {
		rerr = err
		return
	}
	bytes, err := item.Value()
	if err != nil {
		rerr = err
		return
	}
	ent = NewEntity(hash)
	rerr = ent.FromBytes(bytes)
	return
}

func (bg *BadgerGraph) GetExtendedIndex(hash HashKey) (ent EntityExtendedIndex, rerr error) {
	hash = hash.AsType(EXTENDED)
	item, err := bg.tx.Get(hash[:])
	if err == badger.ErrKeyNotFound {
		rerr = ErrNotFound
		return
	} else if err != nil {
		rerr = err
		return
	}
	bytes, err := item.Value()
	if err != nil {
		rerr = err
		return
	}
	ent = NewEntityExtendedIndex(hash)
	rerr = ent.FromBytes(bytes)
	return
}

func (bg *BadgerGraph) GetPredicate(hash HashKey) (ent PredicateEntity, rerr error) {
	hash = hash.AsType(PREDICATE)
	item, err := bg.tx.Get(hash[:])
	if err == badger.ErrKeyNotFound {
		rerr = ErrNotFound
		return
	} else if err != nil {
		rerr = err
		return
	}
	bytes, err := item.Value()
	if err != nil {
		rerr = err
		return
	}
	ent = NewPredicateEntity(hash)
	rerr = ent.FromBytes(bytes)
	return
}

func (bg *BadgerGraph) IterateAllEntities(f func(HashKey, Entity) bool) error {
	iter := bg.tx.NewIterator(badger.DefaultIteratorOptions)
	var start = []byte{0, 0, 0, 2, 0, 0, 0, 0}
	iter.Seek(start)
	defer iter.Close()
	for iter.Rewind(); iter.Valid(); iter.Next() {
		if !iter.ValidForPrefix([]byte{0, 0, 0, 2}) {
			iter.Close()
		}

		bytes, err := iter.Item().Value()
		if err != nil {
			return err
		}
		var ent Entity
		if err = ent.FromBytes(bytes); err != nil {
			return err
		}
		if f(ent.Key(), ent) {
			iter.Close()
		}
	}
	return nil
}

type BadgerGraphTx struct {
	*BadgerGraph
}

func (bgtx *BadgerGraphTx) Commit() error {
	return bgtx.tx.Commit(nil)
}

func (bgtx *BadgerGraphTx) Release() {
	bgtx.tx.Discard()
}

func (bgtx *BadgerGraphTx) Version() Version {
	return bgtx.version
}

func (bgtx *BadgerGraphTx) PutURI(uri turtle.URI) (hash HashKey, rerr error) {

	var uri_bytes = uri.Bytes()
	if len(uri_bytes) == 0 {
		return
	}

	if item, err := bgtx.tx.Get(uri_bytes); err != nil && err != badger.ErrKeyNotFound {
		rerr = errors.Wrap(err, "Could not check key existance")
		return
	} else if err == nil {
		_, rerr = item.ValueCopy(hash[:])
		return
	} else if err == badger.ErrKeyNotFound {
		var salt = uint64(0)
		hashURI(uri, &hash, salt)
		for {
			if _, err := bgtx.tx.Get(hash[:]); err == nil {
				salt += 1
				hashURI(uri, &hash, salt)
			} else if err != nil && err != badger.ErrKeyNotFound {
				rerr = errors.Wrapf(err, "Error checking db membership for %v", hash)
				return
			} else {
				break
			}
		}
		pkhash := hash.AsType(PK)
		bgtx.set(pkhash[:], uri_bytes)
		bgtx.set(uri_bytes, hash[:])
	}

	enthash := hash.AsType(ENTITY)
	_, rerr = bgtx.tx.Get(enthash[:])
	if rerr == badger.ErrKeyNotFound {
		ent := NewEntity(enthash)
		rerr = bgtx.PutEntity(ent)
		return
	}

	return
}

func (bgtx *BadgerGraphTx) PutEntity(ent Entity) error {
	hash := ent.Key().AsType(ENTITY)
	return bgtx.set(hash[:], ent.Bytes())
}

func (bgtx *BadgerGraphTx) PutExtendedIndex(ent EntityExtendedIndex) error {
	hash := ent.Key().AsType(EXTENDED)
	return bgtx.set(hash[:], ent.Bytes())
}

func (bgtx *BadgerGraphTx) PutPredicate(ent PredicateEntity) error {
	hash := ent.Key().AsType(PREDICATE)
	return bgtx.set(hash[:], ent.Bytes())
}

func (bgtx *BadgerGraphTx) PutReversePredicate(uri, inverse turtle.URI) error {
	bgtx.inverse[uri] = inverse
	return nil
}

// set key->value. If the transaction is too big, commit it, open a new transaction,
// and retry setting the key/value pair
func (bgtx *BadgerGraphTx) set(key, val []byte) error {
	if err := bgtx.tx.Set(key, val); err == badger.ErrTxnTooBig {
		if err = bgtx.tx.Commit(nil); err != nil {
			return err
		}
		bgtx.tx = bgtx.db.NewTransaction(true)
		return bgtx.set(key, val)
	} else {
		return err
	}
}
