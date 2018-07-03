package storage

import (
	"os"
	"strings"

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
	path       string
	entityDB   *leveldb.DB
	pkDB       *leveldb.DB
	predDB     *leveldb.DB
	graphDB    *leveldb.DB
	extendedDB *leveldb.DB
}

func (ldb *LevelDBStorageProvider) Initialize(name string, cfg *config.Config) (err error) {
	ldb.path = strings.TrimSuffix(cfg.DBPath, "/")
	options := &opt.Options{
		Filter: filter.NewBloomFilter(32),
	}

	// set up entity, pk databases
	entityDBPath := ldb.path + "/db-entities"
	ldb.entityDB, err = leveldb.OpenFile(entityDBPath, options)
	if err != nil {
		return errors.Wrapf(err, "Could not open entityDB file %s", entityDBPath)
	}

	pkDBPath := ldb.path + "/db-pk"
	ldb.pkDB, err = leveldb.OpenFile(pkDBPath, options)
	if err != nil {
		return errors.Wrapf(err, "Could not open pkDB file %s", pkDBPath)
	}

	// set up entity, pk databases
	extendedDBPath := ldb.path + "/db-extended"
	ldb.extendedDB, err = leveldb.OpenFile(extendedDBPath, options)
	if err != nil {
		return errors.Wrapf(err, "Could not open extendedDB file %s", extendedDBPath)
	}

	graphDBPath := ldb.path + "/db-graph"
	ldb.graphDB, err = leveldb.OpenFile(graphDBPath, options)
	if err != nil {
		return errors.Wrapf(err, "Could not open graphDB file %s", graphDBPath)
	}

	predDBPath := ldb.path + "/db-pred"
	ldb.predDB, err = leveldb.OpenFile(predDBPath, options)
	if err != nil {
		return errors.Wrapf(err, "Could not open predDB file %s", predDBPath)
	}

	return nil
}

func (ldb *LevelDBStorageProvider) Close() error {
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

func (ldb *LevelDBStorageProvider) OpenTransaction() (Transaction, error) {
	tx := &LevelDBTransaction{}
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
	var err error
	if tx.entity, err = getTransaction(ldb.entityDB); err != nil {
		return tx, err
	}
	if tx.pk, err = getTransaction(ldb.pkDB); err != nil {
		return tx, err
	}
	if tx.graph, err = getTransaction(ldb.graphDB); err != nil {
		return tx, err
	}
	if tx.ext, err = getTransaction(ldb.extendedDB); err != nil {
		return tx, err
	}
	if tx.pred, err = getTransaction(ldb.predDB); err != nil {
		return tx, err
	}
	return tx, nil
}

func (ldb *LevelDBStorageProvider) OpenSnapshot() (Traversable, error) {
	snap := &LevelDBSnapshot{}
	getSnapshot := func(db *leveldb.DB) (*leveldb.Snapshot, error) {
		if dbsnap, err := db.GetSnapshot(); err != nil {
			if snap.entitySnapshot != nil {
				snap.entitySnapshot.Release()
			}
			if snap.pkSnapshot != nil {
				snap.pkSnapshot.Release()
			}
			if snap.predSnapshot != nil {
				snap.predSnapshot.Release()
			}
			if snap.pkSnapshot != nil {
				snap.pkSnapshot.Release()
			}
			if snap.extendedSnapshot != nil {
				snap.extendedSnapshot.Release()
			}
			return nil, err
		} else {
			return dbsnap, nil
		}
	}
	var err error
	if snap.entitySnapshot, err = getSnapshot(ldb.entityDB); err != nil {
		return nil, err
	}
	if snap.pkSnapshot, err = getSnapshot(ldb.pkDB); err != nil {
		return nil, err
	}
	if snap.predSnapshot, err = getSnapshot(ldb.predDB); err != nil {
		return nil, err
	}
	if snap.graphSnapshot, err = getSnapshot(ldb.graphDB); err != nil {
		return nil, err
	}
	if snap.extendedSnapshot, err = getSnapshot(ldb.extendedDB); err != nil {
		return nil, err
	}
	return snap, nil
}

type LevelDBSnapshot struct {
	entitySnapshot   *leveldb.Snapshot
	pkSnapshot       *leveldb.Snapshot
	predSnapshot     *leveldb.Snapshot
	graphSnapshot    *leveldb.Snapshot
	extendedSnapshot *leveldb.Snapshot
}

func (snap *LevelDBSnapshot) Has(bucket HodBucket, key []byte) (exists bool, err error) {
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

func (snap *LevelDBSnapshot) Get(bucket HodBucket, key []byte) (value []byte, err error) {
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

func (snap *LevelDBSnapshot) Put(bucket HodBucket, key []byte, value []byte) (err error) {
	return errors.New("Cannot PUT on snapshot")
}

func (snap *LevelDBSnapshot) Iterate(bucket HodBucket) Iterator {
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

func (tx *LevelDBTransaction) Has(bucket HodBucket, key []byte) (exists bool, err error) {
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

func (tx *LevelDBTransaction) Get(bucket HodBucket, key []byte) (value []byte, err error) {
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

func (tx *LevelDBTransaction) Put(bucket HodBucket, key []byte, value []byte) (err error) {
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

func (tx *LevelDBTransaction) Iterate(bucket HodBucket) Iterator {
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
