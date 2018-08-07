package storage

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gtfierro/hod/config"

	"github.com/pkg/errors"
	logrus "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, ForceColors: true})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)
}

type LevelDBStorageProvider struct {
	path            string
	name            string
	cfg             *config.Config
	loaded_versions sync.Map
	sync.Mutex
}

type levelDBInstance struct {
	name       string
	version    uint64
	entityDB   *leveldb.DB
	pkDB       *leveldb.DB
	predDB     *leveldb.DB
	graphDB    *leveldb.DB
	extendedDB *leveldb.DB
}

func (ldb *levelDBInstance) Close() error {
	checkError := func(err error) {
		if err != nil {
			logrus.Fatal(err)
		}
	}
	checkError(ldb.entityDB.Close())
	checkError(ldb.pkDB.Close())
	checkError(ldb.predDB.Close())
	checkError(ldb.graphDB.Close())
	checkError(ldb.extendedDB.Close())
	return nil
}

func (ldb *LevelDBStorageProvider) Initialize(name string, cfg *config.Config) (err error) {
	ldb.path = strings.TrimSuffix(cfg.DBPath, "/")
	path_base := filepath.Join(ldb.path, strconv.FormatUint(ldb.latest_version(), 10))
	err = os.MkdirAll(path_base, 0755)
	if err != nil {
		logrus.Warning(err)
		return err
	}
	ldb.cfg = cfg
	return nil
}

func (ldb *LevelDBStorageProvider) Close() (err error) {
	ldb.loaded_versions.Range(func(_version, _instance interface{}) bool {
		instance := _instance.(*levelDBInstance)
		if _err := instance.Close(); _err != nil {
			err = _err
		}
		return true
	})
	return
}

func (ldb *LevelDBStorageProvider) OpenVersion(version uint64) (trav Traversable, err error) {
	if version == 0 {
		version = ldb.latest_version()
	}
	logrus.Warning("OPENING version ", version)
	if _instance, ok := ldb.loaded_versions.Load(version); !ok {
		return ldb.newWithVersion(version)
	} else {
		return _instance.(*levelDBInstance), nil
	}
}

func (ldb *LevelDBStorageProvider) newWithVersion(version uint64) (instance *levelDBInstance, err error) {
	ldb.Lock()
	defer ldb.Unlock()
	logrus.Warning("new with version")

	instance = &levelDBInstance{name: ldb.name, version: version}

	options := &opt.Options{
		Filter: filter.NewBloomFilter(32),
	}

	path_base := filepath.Join(ldb.path, ldb.name, strconv.FormatUint(version, 10))
	logrus.Warning("here", path_base)
	_, err = os.Stat(path_base)
	if err != nil && !os.IsNotExist(err) {
		return
	} else if os.IsNotExist(err) {
		logrus.Warning(err)
		err = os.MkdirAll(path_base, os.ModeDir)
		if err != nil {
			return
		}
	}

	// set up entity, pk databases
	entityDBPath := filepath.Join(path_base, "db-entities")
	instance.entityDB, err = leveldb.OpenFile(entityDBPath, options)
	if err != nil {
		err = errors.Wrapf(err, "Could not open entityDB file %s", entityDBPath)
		return
	}

	pkDBPath := filepath.Join(path_base, "db-pk")
	instance.pkDB, err = leveldb.OpenFile(pkDBPath, options)
	if err != nil {
		err = errors.Wrapf(err, "Could not open pkDB file %s", pkDBPath)
		return
	}

	// set up entity, pk databases
	extendedDBPath := filepath.Join(path_base, "db-extended")
	instance.extendedDB, err = leveldb.OpenFile(extendedDBPath, options)
	if err != nil {
		err = errors.Wrapf(err, "Could not open extendedDB file %s", extendedDBPath)
		return
	}

	graphDBPath := filepath.Join(path_base, "db-graph")
	instance.graphDB, err = leveldb.OpenFile(graphDBPath, options)
	if err != nil {
		err = errors.Wrapf(err, "Could not open graphDB file %s", graphDBPath)
		return
	}

	predDBPath := filepath.Join(path_base, "db-pred")
	instance.predDB, err = leveldb.OpenFile(predDBPath, options)
	if err != nil {
		err = errors.Wrapf(err, "Could not open predDB file %s", predDBPath)
		return
	}

	ldb.loaded_versions.Store(version, instance)
	logrus.Warningf("%v", instance)
	return instance, nil
}

func (ldb *LevelDBStorageProvider) latest_version() uint64 {
	latest_version := uint64(0)
	ldb.loaded_versions.Range(func(_version, _instance interface{}) bool {
		version := _version.(uint64)
		if version > latest_version {
			latest_version = version
		}
		return true
	})

	path_base := filepath.Join(ldb.path, ldb.name, "*")
	matches, err := filepath.Glob(path_base)
	if err != nil {
		panic(err)
	}
	for _, match := range matches {
		version, err := strconv.ParseUint(filepath.Base(match), 10, 64)
		if err != nil {
			continue
		}
		if version > latest_version {
			latest_version = version
		}
	}

	if latest_version == 0 {
		latest_version = uint64(time.Now().UnixNano())
	}
	return latest_version
}

func (ldb *LevelDBStorageProvider) OpenTransaction() (Transaction, error) {
	ldb.Lock()
	latest_version := ldb.latest_version()
	newest_version := uint64(time.Now().UnixNano())

	old_path_base := filepath.Join(ldb.path, ldb.name, strconv.FormatUint(latest_version, 10))
	new_path_base := filepath.Join(ldb.path, ldb.name, strconv.FormatUint(newest_version, 10))
	logrus.Warning("copy>", old_path_base, " > ", new_path_base)
	if err := CopyDir(old_path_base, new_path_base); err != nil {
		ldb.Unlock()
		return nil, err
	}

	logrus.Warning("done copying")
	ldb.Unlock()
	return ldb.newWithVersion(newest_version)
}

type LevelDBSnapshot struct {
	entitySnapshot   *leveldb.Snapshot
	pkSnapshot       *leveldb.Snapshot
	predSnapshot     *leveldb.Snapshot
	graphSnapshot    *leveldb.Snapshot
	extendedSnapshot *leveldb.Snapshot
}

func (snap *LevelDBSnapshot) Has(bucket HodNamespace, key []byte) (exists bool, err error) {
	switch bucket {
	case EntityBucket:
		return snap.entitySnapshot.Has(key, nil)
	case PKBucket:
		return snap.pkSnapshot.Has(key, nil)
	case PredBucket:
		return snap.predSnapshot.Has(key, nil)
	case GraphBucket:
		return snap.graphSnapshot.Has(key, nil)
	case ExtendedBucket:
		return snap.extendedSnapshot.Has(key, nil)
	}
	return false, errors.New("Invalid bucket")
}

func (snap *LevelDBSnapshot) Get(bucket HodNamespace, key []byte) (value []byte, err error) {
	switch bucket {
	case EntityBucket:
		value, err = snap.entitySnapshot.Get(key, nil)
	case PKBucket:
		value, err = snap.pkSnapshot.Get(key, nil)
	case PredBucket:
		value, err = snap.predSnapshot.Get(key, nil)
	case GraphBucket:
		value, err = snap.graphSnapshot.Get(key, nil)
	case ExtendedBucket:
		value, err = snap.extendedSnapshot.Get(key, nil)
	}
	if err == leveldb.ErrNotFound {
		err = ErrNotFound
	}
	return value, err
}

func (snap *LevelDBSnapshot) Put(bucket HodNamespace, key []byte, value []byte) (err error) {
	return errors.New("Cannot PUT on snapshot")
}

func (snap *LevelDBSnapshot) Iterate(bucket HodNamespace) Iterator {
	switch bucket {
	case EntityBucket:
		return snap.entitySnapshot.NewIterator(nil, nil)
	case PKBucket:
		return snap.pkSnapshot.NewIterator(nil, nil)
	case PredBucket:
		return snap.predSnapshot.NewIterator(nil, nil)
	case GraphBucket:
		return snap.graphSnapshot.NewIterator(nil, nil)
	case ExtendedBucket:
		return snap.extendedSnapshot.NewIterator(nil, nil)
	}
	return nil
}

func (snap *LevelDBSnapshot) Release() {
	snap.entitySnapshot.Release()
	snap.pkSnapshot.Release()
	snap.predSnapshot.Release()
	snap.graphSnapshot.Release()
	snap.extendedSnapshot.Release()
}

type LevelDBTransaction struct {
	entity *leveldb.Transaction
	pk     *leveldb.Transaction
	graph  *leveldb.Transaction
	ext    *leveldb.Transaction
	pred   *leveldb.Transaction
}

func (tx *LevelDBTransaction) Commit() error {
	if err := tx.entity.Commit(); err != nil {
		tx.Release()
		return err
	}
	if err := tx.pk.Commit(); err != nil {
		tx.Release()
		return err
	}
	if err := tx.graph.Commit(); err != nil {
		tx.Release()
		return err
	}
	if err := tx.ext.Commit(); err != nil {
		tx.Release()
		return err
	}
	if err := tx.pred.Commit(); err != nil {
		tx.Release()
		return err
	}
	return nil
}

func (tx *LevelDBTransaction) Has(bucket HodNamespace, key []byte) (exists bool, err error) {
	switch bucket {
	case EntityBucket:
		return tx.entity.Has(key, nil)
	case PKBucket:
		return tx.pk.Has(key, nil)
	case PredBucket:
		return tx.pred.Has(key, nil)
	case GraphBucket:
		return tx.graph.Has(key, nil)
	case ExtendedBucket:
		return tx.ext.Has(key, nil)
	}
	return false, errors.New("Invalid bucket")
}

func (tx *LevelDBTransaction) Get(bucket HodNamespace, key []byte) (value []byte, err error) {
	switch bucket {
	case EntityBucket:
		value, err = tx.entity.Get(key, nil)
	case PKBucket:
		value, err = tx.pk.Get(key, nil)
	case PredBucket:
		value, err = tx.pred.Get(key, nil)
	case GraphBucket:
		value, err = tx.graph.Get(key, nil)
	case ExtendedBucket:
		value, err = tx.ext.Get(key, nil)
	}
	if err == leveldb.ErrNotFound {
		err = ErrNotFound
	}
	return value, err
}

func (tx *LevelDBTransaction) Put(bucket HodNamespace, key []byte, value []byte) (err error) {
	switch bucket {
	case EntityBucket:
		return tx.entity.Put(key, value, nil)
	case PKBucket:
		return tx.pk.Put(key, value, nil)
	case PredBucket:
		return tx.pred.Put(key, value, nil)
	case GraphBucket:
		return tx.graph.Put(key, value, nil)
	case ExtendedBucket:
		return tx.ext.Put(key, value, nil)
	}
	return errors.New("Invalid bucket")
}

func (tx *LevelDBTransaction) Iterate(bucket HodNamespace) Iterator {
	switch bucket {
	case EntityBucket:
		return tx.entity.NewIterator(nil, nil)
	case PKBucket:
		return tx.pk.NewIterator(nil, nil)
	case PredBucket:
		return tx.pred.NewIterator(nil, nil)
	case GraphBucket:
		return tx.graph.NewIterator(nil, nil)
	case ExtendedBucket:
		return tx.ext.NewIterator(nil, nil)
	}
	return nil
}

func (tx *LevelDBTransaction) Release() {
	tx.entity.Discard()
	tx.pk.Discard()
	tx.graph.Discard()
	tx.ext.Discard()
	tx.pred.Discard()
}

func (inst *levelDBInstance) Has(bucket HodNamespace, key []byte) (exists bool, err error) {
	switch bucket {
	case EntityBucket:
		return inst.entityDB.Has(key, nil)
	case PKBucket:
		return inst.pkDB.Has(key, nil)
	case PredBucket:
		return inst.predDB.Has(key, nil)
	case GraphBucket:
		return inst.graphDB.Has(key, nil)
	case ExtendedBucket:
		return inst.extendedDB.Has(key, nil)
	}
	return false, errors.New("Invalid bucket")
}

func (inst *levelDBInstance) Get(bucket HodNamespace, key []byte) (value []byte, err error) {
	switch bucket {
	case EntityBucket:
		value, err = inst.entityDB.Get(key, nil)
	case PKBucket:
		value, err = inst.pkDB.Get(key, nil)
	case PredBucket:
		value, err = inst.predDB.Get(key, nil)
	case GraphBucket:
		value, err = inst.graphDB.Get(key, nil)
	case ExtendedBucket:
		value, err = inst.extendedDB.Get(key, nil)
	}
	if err == leveldb.ErrNotFound {
		err = ErrNotFound
	}
	return value, err
}

func (inst *levelDBInstance) Put(bucket HodNamespace, key []byte, value []byte) (err error) {
	switch bucket {
	case EntityBucket:
		return inst.entityDB.Put(key, value, nil)
	case PKBucket:
		return inst.pkDB.Put(key, value, nil)
	case PredBucket:
		return inst.predDB.Put(key, value, nil)
	case GraphBucket:
		return inst.graphDB.Put(key, value, nil)
	case ExtendedBucket:
		return inst.extendedDB.Put(key, value, nil)
	}
	return errors.New("Invalid bucket")
}

func (inst *levelDBInstance) Iterate(bucket HodNamespace) Iterator {
	switch bucket {
	case EntityBucket:
		return inst.entityDB.NewIterator(nil, nil)
	case PKBucket:
		return inst.pkDB.NewIterator(nil, nil)
	case PredBucket:
		return inst.predDB.NewIterator(nil, nil)
	case GraphBucket:
		return inst.graphDB.NewIterator(nil, nil)
	case ExtendedBucket:
		return inst.extendedDB.NewIterator(nil, nil)
	}
	return nil
}

func (inst *levelDBInstance) Commit() error {
	return nil
}

func (inst *levelDBInstance) Release() {
}
