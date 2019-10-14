package storage

import (
	"github.com/gtfierro/hod/config"
	"github.com/gtfierro/hod/turtle"
	"github.com/pkg/errors"
	logrus "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

type MemoryStorageProvider struct {
	versions   map[Version]*MemoryGraph
	namespaces map[string]string
	inverse    map[turtle.URI]turtle.URI
	vm         *VersionManager
	versionDir string
	sync.RWMutex
}

// provides configuration information to the storage provider.
// When this method returns, the storage provider is allowed to be used
func (msp *MemoryStorageProvider) Initialize(cfg *config.Config) error {
	msp.versions = make(map[Version]*MemoryGraph)
	msp.namespaces = make(map[string]string)
	msp.inverse = make(map[turtle.URI]turtle.URI)

	// version manager dir
	dir, err := ioutil.TempDir("", "msp")
	if err != nil {
		return err
	}
	msp.versionDir = dir
	vm, err := CreateVersionManager(msp.versionDir)
	if err != nil {
		return err
	}
	msp.vm = vm
	return nil
}

// Closes the storage provider. Further calls to the storage provider
// should return an error
func (msp *MemoryStorageProvider) Close() error {
	closeErr := msp.vm.db.Close()
	if closeErr != nil {
		return closeErr
	}
	os.RemoveAll(msp.versionDir)
	return nil
}

// Adds a new graph to the storage provider under the given name,
// Returns the version of the graph and a boolean 'exists' value
// that is true if the database already existed.
func (msp *MemoryStorageProvider) AddGraph(name string) (version Version, exists bool, err error) {
	logrus.Info("Add graph", name)

	msp.Lock()
	defer msp.Unlock()
	current, err := msp.CurrentVersion(name)
	//if err != nil {
	//	logrus.Error("// ", err)
	//	return
	//}
	//_, found := msp.versions[current]
	//if current.Empty() || !found {
	//	//create new version bc doesn't exist
	//	newversion, verr := msp.vm.NewVersion(name)
	//	if verr != nil {
	//		err = verr
	//		return
	//	}
	//	msp.versions[newversion] = nil // TODO: copy memorygraph
	//	return newversion, false, nil
	//}
	return current, true, nil
}

// creates and returns a new writable version of the graph with the given name.
// This version will not be available until it is committed
func (msp *MemoryStorageProvider) CreateVersion(name string) (tx Transaction, err error) {
	msp.Lock()
	defer msp.Unlock()
	var (
		db *MemoryGraph
	)
	current, err := msp.CurrentVersion(name)
	if err != nil {
		return nil, err
	}
	db = msp.versions[current]
	newversion, err := msp.vm.NewVersion(name)
	if err != nil {
		logrus.Error(errors.Wrap(err, "could not create new version"))
		return
	}
	msp.versions[newversion] = newMemoryGraph(db, newversion, false)

	return msp.versions[newversion], nil
}

// returns the given version of the graph with the given name; returns an error if the version doesn't exist
// The returned transaction should be read-only
func (msp *MemoryStorageProvider) OpenVersion(ver Version) (Transaction, error) {
	var (
		db    *MemoryGraph
		found bool
	)
	if db, found = msp.versions[ver]; !found {
		return nil, ErrGraphNotFound
	}
	return db, nil
}

// lists versions of the graph with the given name
func (msp *MemoryStorageProvider) ListVersions(name string) ([]Version, error) {
	return msp.vm.ListVersions(name)
}

func (msp *MemoryStorageProvider) Names() ([]string, error) {
	return msp.vm.Names()
}

// returns the latest version of the given graph
func (msp *MemoryStorageProvider) CurrentVersion(name string) (Version, error) {
	return msp.vm.GetLatestVersion(name)
}

// returns the version of the given graph at the given timestamp
func (msp *MemoryStorageProvider) VersionAt(name string, timestamp time.Time) (Version, error) {
	return msp.vm.GetVersionAt(name, timestamp)
}

// returns the version active before the one active at the given timestamp
func (msp *MemoryStorageProvider) VersionsBefore(name string, timestamp time.Time, limit int) ([]Version, error) {
	return msp.vm.GetVersionsBefore(name, timestamp, limit)
}

// returns the version active after the one active at the given timestamp
func (msp *MemoryStorageProvider) VersionsAfter(name string, timestamp time.Time, limit int) ([]Version, error) {
	return msp.vm.GetVersionsAfter(name, timestamp, limit)
}

// list all stored versions
func (msp *MemoryStorageProvider) Graphs() ([]Version, error) {
	return msp.vm.Graphs()
}

// return the set of saved abbreviation -> namespace URI mappings
func (msp *MemoryStorageProvider) GetNamespaces() (mapping map[string]string, err error) {
	return msp.namespaces, nil
}

// save a new abbreviation -> namespace URI mapping (e.g. brick -> https://brickschema.org/schema/1.0.3/Brick#)
func (msp *MemoryStorageProvider) SaveNamespace(abbreviation string, uri string) error {
	msp.namespaces[abbreviation] = uri
	return nil
}

// Transaction struct for MemoryStorageProvider
type MemoryGraph struct {
	version      Version
	entityHash   map[turtle.URI]HashKey
	uri          map[HashKey]turtle.URI
	entityObject map[HashKey]Entity
	entityIndex  map[HashKey]EntityExtendedIndex
	pred         map[HashKey]PredicateEntity
	readonly     bool
	inverse      map[turtle.URI]turtle.URI
}

func newMemoryGraph(from *MemoryGraph, version Version, readonly bool) *MemoryGraph {
	newm := &MemoryGraph{
		entityHash:   make(map[turtle.URI]HashKey),
		entityObject: make(map[HashKey]Entity),
		entityIndex:  make(map[HashKey]EntityExtendedIndex),
		uri:          make(map[HashKey]turtle.URI),
		pred:         make(map[HashKey]PredicateEntity),
		inverse:      make(map[turtle.URI]turtle.URI),
		readonly:     readonly,
		version:      version,
	}
	if from != nil {
		for k, v := range from.inverse {
			newm.inverse[k] = v
		}
		for k, v := range from.entityHash {
			newm.entityHash[k] = v
		}
		for k, v := range from.entityObject {
			newm.entityObject[k] = v.Copy()
		}
		for k, v := range from.entityIndex {
			newm.entityIndex[k] = v.Copy()
		}
		for k, v := range from.uri {
			newm.uri[k] = v
		}
		for k, v := range from.pred {
			newm.pred[k] = v.Copy()
		}
	}

	return newm
}

// Commit the transaction
func (mg *MemoryGraph) Commit() error {
	return nil
}

// Release the transaction (read-only) or discard the transaction (rw)
func (mg *MemoryGraph) Release() {
}

// return the current Version of the Transaction
func (mg *MemoryGraph) Version() Version {
	return mg.version
}

// retrive the HashKey for the given URI
func (mg *MemoryGraph) GetHash(uri turtle.URI) (HashKey, error) {
	hash, found := mg.entityHash[uri]
	if !found {
		return hash, ErrNotFound
	}
	return hash, nil
}

// retrive the URI for the given HashKey
func (mg *MemoryGraph) GetURI(hash HashKey) (turtle.URI, error) {
	uri, found := mg.uri[hash]
	if !found {
		return uri, ErrNotFound
	}
	return uri, nil
}

// retrive the Entity object for the given HashKey
func (mg *MemoryGraph) GetEntity(hash HashKey) (Entity, error) {
	ent, found := mg.entityObject[hash]
	if !found {
		return ent, ErrNotFound
	}
	return ent, nil
}

// retrive the ExtendedIndex object for the given HashKey
func (mg *MemoryGraph) GetExtendedIndex(hash HashKey) (EntityExtendedIndex, error) {
	ent, found := mg.entityIndex[hash]
	if !found {
		return ent, ErrNotFound
	}
	return ent, nil
}

// retrive the PredicateEntity object for the given HashKey
func (mg *MemoryGraph) GetPredicate(hash HashKey) (PredicateEntity, error) {
	pred, found := mg.pred[hash]
	if !found {
		return pred, ErrNotFound
	}
	return pred, nil
}

// get the inverse edge for the given URI, if it exists
func (mg *MemoryGraph) GetReversePredicate(rel turtle.URI) (turtle.URI, bool) {
	inverse, found := mg.inverse[rel]
	return inverse, found
}

// call the provided function for each Entity object in the graph
func (mg *MemoryGraph) IterateAllEntities(cb func(HashKey, Entity) bool) error {
	for hash, ent := range mg.entityObject {
		cb(hash, ent)
	}
	return nil
}

// store the URI and return the HashKey
func (mg *MemoryGraph) PutURI(uri turtle.URI) (hash HashKey, rerr error) {
	var found bool
	if hash, found = mg.entityHash[uri]; found {
		return
	} else {
		var salt = uint64(0)
		hashURI(uri, &hash, salt)
		for {
			if _, found := mg.uri[hash]; found {
				salt++
				hashURI(uri, &hash, salt)
			} else {
				break
			}
		}
		mg.uri[hash] = uri
		mg.entityHash[uri] = hash
	}

	if _, found := mg.entityObject[hash]; !found {
		ent := NewEntity(hash)
		rerr = mg.PutEntity(ent)
	}

	return hash, nil
}

// store the Entity object, mapped by HashKey (Entity.Key())
func (mg *MemoryGraph) PutEntity(ent Entity) error {
	if !mg.readonly {
		mg.entityObject[ent.Key()] = ent
	}
	return nil
}

// store the ExtendedIndex object, mapped by HashKey (ExtendedIndex.Key())
func (mg *MemoryGraph) PutExtendedIndex(ent EntityExtendedIndex) error {
	if !mg.readonly {
		mg.entityIndex[ent.Key()] = ent
	}
	return nil
}

// store the PredicateEntity object, mapped by HashKey (PredicateEntity.Key())
func (mg *MemoryGraph) PutPredicate(pred PredicateEntity) error {
	if !mg.readonly {
		mg.pred[pred.Key()] = pred
	}
	return nil
}

// store the two URIs as inverses of each other
func (mg *MemoryGraph) PutReversePredicate(a, b turtle.URI) error {
	mg.inverse[a] = b
	mg.inverse[b] = a
	return nil
}
