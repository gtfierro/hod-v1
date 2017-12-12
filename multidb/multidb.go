package multidb

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gtfierro/hod/config"
	hoddb "github.com/gtfierro/hod/db"
	query "github.com/gtfierro/hod/lang"
	sparql "github.com/gtfierro/hod/lang/ast"
	"github.com/gtfierro/hod/turtle"

	"github.com/op/go-logging"
	"github.com/pkg/errors"
)

// logger
var log *logging.Logger

func init() {
	log = logging.MustGetLogger("hod-multidb")
	var format = "%{color}%{level} %{shortfile} %{time:Jan 02 15:04:05} %{color:reset} ▶ %{message}"
	var logBackend = logging.NewLogBackend(os.Stderr, "", 0)
	logBackendLeveled := logging.AddModuleLevel(logBackend)
	logging.SetBackend(logBackendLeveled)
	logging.SetFormatter(logging.MustStringFormatter(format))
}

type MultiDB struct {
	// database name => *db.DB
	dbs sync.Map
	// store the config so we can make more databases
	cfg   *config.Config
	dbdir string
}

func NewMultiDB(cfg *config.Config) (*MultiDB, error) {
	var mdb = &MultiDB{cfg: cfg}
	// create path for dbs
	mdb.dbdir = strings.TrimSuffix(cfg.DBPath, "/")
	if err := os.MkdirAll(mdb.dbdir, os.ModeDir); err != nil {
		return nil, errors.Wrapf(err, "Could not create db directory %s", mdb.dbdir)
	}

	p := turtle.GetParser()

	for buildingname, buildingttlfile := range cfg.Buildings {
		cfg.DBPath = filepath.Join(mdb.dbdir, buildingname)
		db, err := hoddb.NewDB(cfg)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not create database at %s", cfg.DBPath)
		}
		ds, duration := p.Parse(buildingttlfile)
		rate := float64((float64(ds.NumTriples()) / float64(duration.Nanoseconds())) * 1e9)
		log.Infof("Loaded %d triples, %d namespaces in %s (%.0f/sec)", ds.NumTriples(), ds.NumNamespaces(), duration, rate)
		err = db.LoadDataset(ds)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not load dataset %s", buildingttlfile)
		}
		mdb.dbs.Store(buildingname, db)
	}

	return mdb, nil
}

func (mdb *MultiDB) LoadMulti(dbs map[string]string) error {
	p := turtle.GetParser()
	for buildingname, buildingttlfile := range dbs {
		mdb.cfg.DBPath = filepath.Join(mdb.dbdir, buildingname)
		db, err := hoddb.NewDB(mdb.cfg)
		if err != nil {
			return errors.Wrapf(err, "Could not create database at %s", mdb.cfg.DBPath)
		}
		ds, duration := p.Parse(buildingttlfile)
		rate := float64((float64(ds.NumTriples()) / float64(duration.Nanoseconds())) * 1e9)
		log.Infof("Loaded %d triples, %d namespaces in %s (%.0f/sec)", ds.NumTriples(), ds.NumNamespaces(), duration, rate)
		err = db.LoadDataset(ds)
		if err != nil {
			return errors.Wrapf(err, "Could not load dataset %s", buildingttlfile)
		}
		mdb.dbs.Store(buildingname, db)
	}
	return nil
}

func (mdb *MultiDB) RunQueryString(querystring string) (hoddb.QueryResult, error) {
	var emptyres hoddb.QueryResult
	if q, err := query.Parse(querystring); err != nil {
		e := errors.Wrap(err, "Could not parse hod query")
		log.Error(e)
		return emptyres, e
	} else if result, err := mdb.RunQuery(q); err != nil {
		e := errors.Wrap(err, "Could not complete hod query")
		log.Error(e)
		return emptyres, e
	} else {
		return result, nil
	}
}

func (mdb *MultiDB) RunQuery(q *sparql.Query) (hoddb.QueryResult, error) {
	var databases = make(map[string]*hoddb.DB)

	// if no FROM clause, then query all dbs!
	if q.From.Empty() {
		q.From.AllDBs = true
	}

	if q.From.AllDBs {
		mdb.dbs.Range(func(_dbname, _db interface{}) bool {
			dbname := _dbname.(string)
			db := _db.(*hoddb.DB)
			databases[dbname] = db
			return true
		})
	} else {
		for _, dbname := range q.From.Databases {
			db, ok := mdb.dbs.Load(dbname)
			if ok {
				databases[dbname] = db.(*hoddb.DB)
			}
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(databases))

	for dbname, db := range databases {
		result, err := db.RunQuery(q)
		if err != nil {
			log.Error(errors.Wrapf(err, "Error running query on %s", dbname))
		}
		log.Info(dbname, result.Count)
		//TODO: merge these or decide how to grouop them
		wg.Done()
	}

	wg.Wait()

	return hoddb.QueryResult{}, nil
}

func (db *MultiDB) LoadDataset(name, ttlfile string) error {
	return nil
}

func (mdb *MultiDB) Close() {
	mdb.dbs.Range(func(_dbname, _db interface{}) bool {
		db := _db.(*hoddb.DB)
		db.Close()
		return true
	})
}

func (mdb *MultiDB) Search(q string, n int) ([]string, error) {
	// just pick first db for now
	var (
		res []string
		err error
	)
	mdb.dbs.Range(func(_dbname, _db interface{}) bool {
		db := _db.(*hoddb.DB)
		res, err = db.Search(q, n)
		if err != nil {
			return true
		}
		return false
	})
	return res, err
}

func (mdb *MultiDB) QueryToClassDOT(q string) (string, error) {
	// just pick first db for now
	var (
		res string
		err error
	)
	mdb.dbs.Range(func(_dbname, _db interface{}) bool {
		db := _db.(*hoddb.DB)
		res, err = db.QueryToClassDOT(q)
		if err != nil {
			return true
		}
		return false
	})
	return res, err
}

func (mdb *MultiDB) QueryToDOT(q string) (string, error) {
	// just pick first db for now
	var (
		res string
		err error
	)
	mdb.dbs.Range(func(_dbname, _db interface{}) bool {
		db := _db.(*hoddb.DB)
		res, err = db.QueryToDOT(q)
		if err != nil {
			return true
		}
		return false
	})
	return res, err
}
