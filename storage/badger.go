package storage

import (
	"encoding/json"
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
//
// HodDB needs to store many different versions of a graph. Currently, the BadgerStorageProvider implements this
// by copying the full database to a new directory every time a new version is created. The storage provider also
// needs to be able to have multiple versions open at once. The protocol for this is as follows:
//
// BadgerStorageProvider has two directories under HOD_DIR (from configuration): HOD_DIR/versions contains the serialized backups
// of all versions of all databases (organized according to <graph name>/<version timestamp>. HOD_DIR/open contains actively opened databases; ONE of which
// will be loaded read-write, and the rest of which will be loaded read-only (historical versions)
//
// CreateVersion(name string): save the current version of the database using badger.Backup("HOD_DIR/versions/<graph name>/<version timestamp>", 0).
// then, open a new database under HOD_DIR/open/<graph name>/<version timestamp>.
//
// OpenVersion(version Version): first looks to see if this version is aready loaded using a storage provider cache. If the
// version is not already loaded, uses badger.Restore("HOD_DIR/versions/<graph name>/<version timestamp>") to load a version
// from the serialized backup into a badger instance opened at HOD_DIR/open/<graph name>/<version timestamp>

type BadgerStorageProvider struct {
	namespaceIndexPath   string
	basedir              string
	versionsDir, openDir string
	dbs                  map[Version]*badger.DB
	vm                   *VersionManager
	sync.RWMutex
}

// Initialize Badger-backed storage
// TODO: use configured storage directory
// TODO: return ErrGraphNotFound when version is non existant
func (bsp *BadgerStorageProvider) Initialize(cfg *config.Config) error {
	bsp.basedir = cfg.DBPath
	if err := os.MkdirAll(bsp.basedir, 0700); err != nil {
		return err
	}
	bsp.namespaceIndexPath = filepath.Join(bsp.basedir, "namespaces.json")

	bsp.versionsDir = filepath.Join(bsp.basedir, "versions")
	if err := os.MkdirAll(bsp.versionsDir, 0700); err != nil {
		return err
	}
	bsp.openDir = filepath.Join(bsp.basedir, "open")
	if err := os.MkdirAll(bsp.openDir, 0700); err != nil {
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
		openDir := filepath.Join(bsp.openDir, version.Name) //, strconv.Itoa(int(version.Timestamp)))
		logrus.Info("Opening existing version ", version, " at ", openDir)
		if err = os.MkdirAll(openDir, 0700); err != nil {
			return err
		}
		opts := badger.DefaultOptions
		opts.Dir = openDir
		opts.ValueDir = openDir
		if bsp.dbs[version], err = badger.Open(opts); err != nil {
			return err
		}
		backupFileLocation := filepath.Join(bsp.versionsDir, version.Name, strconv.Itoa(int(version.Timestamp)))
		backupFile, err := os.Open(backupFileLocation)
		defer backupFile.Close()
		if err != nil {
			return errors.Wrap(err, "Could not open backup file")
		}
		err = bsp.dbs[version].Load(backupFile)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bsp *BadgerStorageProvider) GetNamespaces() (map[string]string, error) {
	var m = make(map[string]string)
	f, err := os.Open(bsp.namespaceIndexPath)
	if os.IsNotExist(err) {
		return m, nil
	} else if err != nil {
		return m, errors.Wrap(err, "Could not open namespace index")
	}
	dec := json.NewDecoder(f)
	err = dec.Decode(&m)
	return m, errors.Wrap(err, "Could not decode namespace index")
}

func (bsp *BadgerStorageProvider) SaveNamespace(abbreviation string, uri string) error {
	namespaces, err := bsp.GetNamespaces()
	if err != nil {
		return err
	}
	namespaces[abbreviation] = uri
	f, err := os.Create(bsp.namespaceIndexPath)
	if err != nil {
		return errors.Wrap(err, "Could not open namespace index")
	}

	enc := json.NewEncoder(f)
	return errors.Wrap(enc.Encode(namespaces), "Could not save namespace index")
}

// Close closes all underlying storage media
// Further calls to the storage provider should return an error
func (bsp *BadgerStorageProvider) Close() error {
	bsp.Lock()
	defer bsp.Unlock()
	//TODO: close internal dbs?
	for _, db := range bsp.dbs {
		closeErr := db.Close()
		if closeErr != nil {
			return closeErr
		}
	}
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

	//exists = false
	//bsp.RLock()
	//for ver := range bsp.dbs {
	//	if ver.Name == name {
	//		exists = true
	//		version = ver
	//	}
	//}
	//bsp.RUnlock()

	//if exists {
	//	logrus.Info(version)
	//	return
	//}

	// make directories for the graph
	versionDir := filepath.Join(bsp.versionsDir, name)
	if err = os.MkdirAll(versionDir, 0700); err != nil {
		return
	}
	openDir := filepath.Join(bsp.openDir, name)
	if err = os.MkdirAll(openDir, 0700); err != nil {
		return
	}

	// create the new version if it doesn't already exist
	current, err := bsp.CurrentVersion(name)
	if err != nil {
		logrus.Error("// ", err)
		return
	}
	logrus.Warning("Current version in add graph -> ", current)
	_, found := bsp.dbs[current]
	if current.Empty() || !found {
		// create new version because it doesn't exist
		newversion, verr := bsp.vm.NewVersion(name)
		if verr != nil {
			err = verr
			return
		}
		opts := badger.DefaultOptions
		openCurrentVersionDir := filepath.Join(bsp.openDir, newversion.Name) //, strconv.Itoa(int(newversion.Timestamp)))
		logrus.Info("Creating new version: ", newversion, " at ", openCurrentVersionDir)
		opts.Dir = openCurrentVersionDir
		opts.ValueDir = openCurrentVersionDir
		bsp.dbs[newversion], err = badger.Open(opts)
		return newversion, false, err
	}
	return current, true, nil
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
	defer bsp.Unlock()
	current, err := bsp.CurrentVersion(name)
	if err != nil {
		return nil, err
	}
	if db, found = bsp.dbs[current]; !found {
		return nil, ErrGraphNotFound
	}
	logrus.Warning("Current version ", current)

	// backup current version
	//backupFileLocation := filepath.Join(bsp.versionsDir, current.Name, strconv.Itoa(int(current.Timestamp)))
	//backupFile, err := os.Create(backupFileLocation)
	//defer backupFile.Close()
	//if err != nil {
	//	err = errors.Wrap(err, "Could not open backup file1")
	//	return
	//}
	//_, err = db.Backup(backupFile, 0)
	//if err != nil {
	//	logrus.Error(errors.Wrap(err, "could not backup to file"))
	//	return
	//}
	//logrus.Info("Backup version ", current, " to ", backupFileLocation)

	// create new version
	newversion, err := bsp.vm.NewVersion(name)
	if err != nil {
		logrus.Error(errors.Wrap(err, "could not create new version"))
		return
	}
	delete(bsp.dbs, current)
	bsp.dbs[newversion] = db

	//newVersionFileLocation := filepath.Join(bsp.openDir, newversion.Name, strconv.Iota(int(noversion.Timestamp)))
	//if err = os.MkdirAll(newVersionFileLocation, 0700); err != nil {
	//	return
	//}
	//opts := badger.DefaultOptions
	//opts.Dir = newVersionFileLocation
	//opts.ValueDir = newVersionFileLocation
	//if bsp.dbs[version], err = badger.Open(opts); err != nil {
	//	return err
	//}
	// load backup

	//	newversion, err := bsp.vm.NewVersion(name)
	//	if err != nil {
	//		logrus.Error(err)
	//		return
	//	}
	//	// file for the current version
	//	dir := filepath.Join(bsp.basedir, current.Name, strconv.Itoa(int(current.Timestamp)))
	//	backupFile, err := os.Create(dir)
	//	if err != nil {
	//		return
	//	}
	//
	//	// backup since beginning of time
	//	_, err = db.Backup(backupFile, 0)
	//	if err != nil {
	//		logrus.Error(err)
	//		return
	//	}
	//
	//	newversion, err := bsp.vm.NewVersion(name)
	//	if err != nil {
	//		logrus.Error(err)
	//		return
	//	}
	//	logrus.Warning(newversion)

	tx = &BadgerGraph{
		version:        newversion,
		db:             db,
		tx:             db.NewTransaction(true),
		readonly:       false,
		backupLocation: filepath.Join(bsp.versionsDir, newversion.Name, strconv.Itoa(int(newversion.Timestamp))),
		inverse:        make(map[turtle.URI]turtle.URI),
	}

	return
}

// OpenVersion returns the given version of the graph with the given name; returns an error if the version doesn't exist
// The returned transaction should be read-only
func (bsp *BadgerStorageProvider) OpenVersion(ver Version) (tx Transaction, err error) {
	// get current version for given name
	var (
		db         *badger.DB
		found      bool
		backupFile *os.File
	)
	bsp.Lock()
	defer bsp.Unlock()
	db, found = bsp.dbs[ver]
	if !found {
		logrus.Warning("Did not find open version ", ver)
		backupFileLocation := filepath.Join(bsp.versionsDir, ver.Name, strconv.Itoa(int(ver.Timestamp)))
		backupFile, err = os.Open(backupFileLocation)
		defer backupFile.Close()
		if err != nil {
			err = errors.Wrap(err, "Could not open backup file")
			return
		}
		newVersionFileLocation := filepath.Join(bsp.openDir, ver.Name, strconv.Itoa(int(ver.Timestamp)))
		if err = os.MkdirAll(newVersionFileLocation, 0700); err != nil {
			return
		}
		opts := badger.DefaultOptions
		opts.Dir = newVersionFileLocation
		opts.ValueDir = newVersionFileLocation
		logrus.Warning("Opening ", ver, " at ", newVersionFileLocation)
		if db, err = badger.Open(opts); err != nil {
			return
		}
		err = db.Load(backupFile)
		if err != nil {
			return
		}
		bsp.dbs[ver] = db

		//bsp.Unlock()
	} else {
		logrus.Warning("Found open version ", ver, ": ", db == nil)
	}
	tx = &BadgerGraph{
		version:  ver,
		db:       db,
		tx:       db.NewTransaction(false), // read-only
		readonly: true,
		inverse:  make(map[turtle.URI]turtle.URI),
	}

	return
}

// ListVersions lists all stored versions for the given graph
func (bsp *BadgerStorageProvider) ListVersions(name string) (versions []Version, err error) {
	return bsp.vm.ListVersions(name)
}

// BadgerGraph is a badger transaction representing a Version of a Brick graph
type BadgerGraph struct {
	version        Version
	db             *badger.DB
	tx             *badger.Txn
	readonly       bool
	backupLocation string
	inverse        map[turtle.URI]turtle.URI
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

	if item, err := bg.tx.Get(hash[:]); err == badger.ErrKeyNotFound || item == nil {
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

func (bg *BadgerGraph) backup() error {
	a, b := os.Stat(bg.backupLocation)
	logrus.Warning("a ", a, "b ", b)
	backupFile, err := os.Create(bg.backupLocation)
	defer backupFile.Close()
	if err != nil {
		return errors.Wrap(err, "could not create backup file location")
	}
	_, err = bg.db.Backup(backupFile, 0)
	if err != nil {
		return errors.Wrap(err, "could not backup to file")
	}
	logrus.Info("Backup version ", bg.version, " to ", bg.backupLocation)
	return nil
}

// Commit the transaction
func (bg *BadgerGraph) Commit() error {
	logrus.Warning("committing ", bg.version)
	commitErr := bg.tx.Commit(nil)
	if commitErr != nil {
		return commitErr
	}
	if !bg.readonly {
		logrus.Warning("backing up in commit")
		err := bg.backup()
		if err != nil {
			return err
		}
	}
	return nil
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
