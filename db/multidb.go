package db

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gtfierro/hod/config"
	query "github.com/gtfierro/hod/lang"
	sparql "github.com/gtfierro/hod/lang/ast"
	"github.com/gtfierro/hod/turtle"

	"github.com/mitghi/btree"
	"github.com/pkg/errors"
)

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
	if err := os.MkdirAll(mdb.dbdir, 0700); err != nil {
		return nil, errors.Wrapf(err, "Could not create db directory %s", mdb.dbdir)
	}

	p := turtle.GetParser()

	for buildingname, buildingttlfile := range cfg.Buildings {
		cfg.DBPath = filepath.Join(mdb.dbdir, buildingname)
		db, err := NewDB(cfg)
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
		db, err := NewDB(mdb.cfg)
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

func (mdb *MultiDB) RunQueryString(querystring string) (QueryResult, error) {
	var emptyres QueryResult
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

func (mdb *MultiDB) RunQuery(q *sparql.Query) (QueryResult, error) {
	var databases = make(map[string]*DB)

	// if no FROM clause, then query all dbs!
	if q.From.Empty() {
		q.From.AllDBs = true
	}

	if q.From.AllDBs {
		mdb.dbs.Range(func(_dbname, _db interface{}) bool {
			dbname := _dbname.(string)
			db := _db.(*DB)
			databases[dbname] = db
			return true
		})
	} else {
		for _, dbname := range q.From.Databases {
			db, ok := mdb.dbs.Load(dbname)
			if ok {
				databases[dbname] = db.(*DB)
			}
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(databases))
	unionedRows := btree.New(4, "")
	var result QueryResult

	for dbname, db := range databases {
		singleresult, err := db.runQueryToSet(q)
		if err != nil {
			log.Error(errors.Wrapf(err, "Error running query on %s", dbname))
		}
		for _, row := range singleresult {
			unionedRows.ReplaceOrInsert(row)
		}

		result.Count = unionedRows.Len()
		if !q.Count {
			i := unionedRows.DeleteMax()
			for i != nil {
				row := i.(*ResultRow)
				m := make(ResultMap)
				for idx, vname := range q.Select.Vars {
					m[vname] = row.row[idx]
				}
				result.Rows = append(result.Rows, m)
				finishResultRow(row)
				i = unionedRows.DeleteMax()
			}
		}

		//TODO: merge these or decide how to grouop them
		wg.Done()
	}

	wg.Wait()

	return result, nil
}

func (db *MultiDB) LoadDataset(name, ttlfile string) error {
	return nil
}

func (mdb *MultiDB) Close() {
	mdb.dbs.Range(func(_dbname, _db interface{}) bool {
		db := _db.(*DB)
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
		db := _db.(*DB)
		res, err = db.search(q, n)
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
		db := _db.(*DB)
		res, err = db.queryToClassDOT(q)
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
		db := _db.(*DB)
		res, err = db.queryToDOT(q)
		if err != nil {
			return true
		}
		return false
	})
	return res, err
}

func (mdb *MultiDB) abbreviate(uri turtle.URI) string {
	// just pick first db for now
	var (
		res string
		err error
	)
	mdb.dbs.Range(func(_dbname, _db interface{}) bool {
		db := _db.(*DB)
		res = db.abbreviate(uri)
		if err != nil {
			return true
		}
		return false
	})
	return res
}
