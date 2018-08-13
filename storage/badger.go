package storage

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
	//"time"

	"github.com/dgraph-io/badger"
	"github.com/gtfierro/hod/config"
	"github.com/gtfierro/hod/turtle"
	"github.com/onrik/logrus/filename"
	"github.com/pkg/errors"
	logrus "github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, ForceColors: true})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
	fh := filename.NewHook()
	fh.Field = "src"
	logrus.AddHook(fh)
}

// BadgerStorageProvider provides a HodDB storage interface to the github.com/dgraph-io/badger key-value store
type BadgerStorageProvider struct {
	basedir string
	dbs     map[Version]*badger.DB
	vm      *VersionManager
	sync.RWMutex
}

// Initialize Badger-backed storage
// TODO: use configured storage directory
// TODO: return ErrGraphNotFound when version is non existant
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
	vm, err := CreateVersionManager(bsp.basedir)
	if err != nil {
		return err
	}
	bsp.vm = vm

	versions, err := bsp.vm.Graphs()
	if err != nil {
		return err
	}
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
	return nil
}

// Close closes all underlying storage media
// Further calls to the storage provider should return an error
func (bsp *BadgerStorageProvider) Close() error {
	bsp.Lock()
	defer bsp.Unlock()
	//TODO: close internal dbs?
	closeErr := bsp.vm.db.Close()
	rmErr := os.RemoveAll(bsp.basedir)
	if closeErr != nil {
		return closeErr
	}
	if rmErr != nil {
		return rmErr
	}
	return nil
}

// AddGraph adds a new graph to the storage provider under the given name,
// Returns the version of the graph and a boolean 'exists' value
// that is true if the database already existed.
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

// CurrentVersion returns the latest version of the given graph
func (bsp *BadgerStorageProvider) CurrentVersion(name string) (version Version, err error) {
	return bsp.vm.GetLatestVersion(name)
}

// Graphs returns the given version of the graph with the given name; returns an error if the version doesn't exist
func (bsp *BadgerStorageProvider) Graphs() ([]Version, error) {
	return bsp.vm.Graphs()
}

// CreateVersion creates and returns a new writable version of the graph with the given name.
// This version will not be available until it is committed
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
	tx = &BadgerGraph{
		db:      db,
		tx:      db.NewTransaction(true),
		inverse: make(map[turtle.URI]turtle.URI),
	}

	return
}

// OpenVersion returns the given version of the graph with the given name; returns an error if the version doesn't exist
// The returned transaction should be read-only
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
	tx = &BadgerGraph{
		version: ver,
		db:      db,
		tx:      db.NewTransaction(false), // read-only
		inverse: make(map[turtle.URI]turtle.URI),
	}

	return
}

// ListVersions lists all stored versions for the given graph
func (bsp *BadgerStorageProvider) ListVersions(name string) (versions []Version, err error) {
	return bsp.vm.ListVersions(name)
}

// BadgerGraph is a badger transaction representing a Version of a Brick graph
type BadgerGraph struct {
	version Version
	db      *badger.DB
	tx      *badger.Txn
	inverse map[turtle.URI]turtle.URI
}

// GetHash retrives the HashKey for the given URI
func (bg *BadgerGraph) GetHash(uri turtle.URI) (hash HashKey, rerr error) {
	var uriBytes = uri.Bytes()
	if item, err := bg.tx.Get(uriBytes); err == badger.ErrKeyNotFound {
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

// GetURI retrives the URI for the given HashKey
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

// GetReversePredicate gets the inverse edge for the given URI, if it exists
func (bg *BadgerGraph) GetReversePredicate(rel turtle.URI) (inverse turtle.URI, found bool) {
	inverse, found = bg.inverse[rel]
	return
}

// GetEntity retrives the Entity object for the given HashKey
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

// GetExtendedIndex retrives the ExtendedIndex object for the given HashKey
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

// GetPredicate retrives the PredicateEntity object for the given HashKey
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

// IterateAllEntities calls the provided function for each Entity object in the graph
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

// Commit the transaction
func (bg *BadgerGraph) Commit() error {
	return bg.tx.Commit(nil)
}

// Release the transaction (read-only) or discard the transaction (rw)
func (bg *BadgerGraph) Release() {
	bg.tx.Discard()
}

// Version returns the current Version of the Transaction
func (bg *BadgerGraph) Version() Version {
	return bg.version
}

// PutURI stores the URI and return the HashKey
func (bg *BadgerGraph) PutURI(uri turtle.URI) (hash HashKey, rerr error) {

	var uriBytes = uri.Bytes()
	if len(uriBytes) == 0 {
		return
	}

	if item, err := bg.tx.Get(uriBytes); err != nil && err != badger.ErrKeyNotFound {
		rerr = errors.Wrap(err, "Could not check key existance")
		return
	} else if err == nil {
		_, rerr = item.ValueCopy(hash[:])
		return
	} else if err == badger.ErrKeyNotFound {
		var salt = uint64(0)
		hashURI(uri, &hash, salt)
		for {
			if _, err := bg.tx.Get(hash[:]); err == nil {
				salt++
				hashURI(uri, &hash, salt)
			} else if err != nil && err != badger.ErrKeyNotFound {
				rerr = errors.Wrapf(err, "Error checking db membership for %v", hash)
				return
			} else {
				break
			}
		}
		pkhash := hash.AsType(PK)
		if err := bg.set(pkhash[:], uriBytes); err != nil {
			rerr = err
			return
		}
		if err := bg.set(uriBytes, hash[:]); err != nil {
			rerr = err
			return
		}
	}

	enthash := hash.AsType(ENTITY)
	_, rerr = bg.tx.Get(enthash[:])
	if rerr == badger.ErrKeyNotFound {
		ent := NewEntity(enthash)
		rerr = bg.PutEntity(ent)
		return
	}

	return
}

// PutEntity stores the Entity object, mapped by HashKey (Entity.Key())
func (bg *BadgerGraph) PutEntity(ent Entity) error {
	hash := ent.Key().AsType(ENTITY)
	return bg.set(hash[:], ent.Bytes())
}

// PutExtendedIndex stores the ExtendedIndex object, mapped by HashKey (ExtendedIndex.Key())
func (bg *BadgerGraph) PutExtendedIndex(ent EntityExtendedIndex) error {
	hash := ent.Key().AsType(EXTENDED)
	return bg.set(hash[:], ent.Bytes())
}

// PutPredicate stores the PredicateEntity object, mapped by HashKey (PredicateEntity.Key())
func (bg *BadgerGraph) PutPredicate(ent PredicateEntity) error {
	hash := ent.Key().AsType(PREDICATE)
	return bg.set(hash[:], ent.Bytes())
}

// PutReversePredicate stores the two URIs as inverses of each other
func (bg *BadgerGraph) PutReversePredicate(uri, inverse turtle.URI) error {
	bg.inverse[uri] = inverse
	return nil
}

// set key->value. If the transaction is too big, commit it, open a new transaction,
// and retry setting the key/value pair
func (bg *BadgerGraph) set(key, val []byte) error {
	err := bg.tx.Set(key, val)
	if err == badger.ErrTxnTooBig {
		if err = bg.tx.Commit(nil); err != nil {
			return err
		}
		bg.tx = bg.db.NewTransaction(true)
		return bg.set(key, val)
	}
	return err
}
