package multidb

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gtfierro/hod/config"
	hoddb "github.com/gtfierro/hod/db"
	turtle "github.com/gtfierro/hod/goraptor"
	query "github.com/gtfierro/hod/lang"
	sparql "github.com/gtfierro/hod/lang/ast"

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
}

func NewMultiDB(cfg *config.Config) (*MultiDB, error) {
	var mdb = new(MultiDB)
	// create path for dbs
	path := strings.TrimSuffix(cfg.DBPath, "/")
	if err := os.MkdirAll(path, os.ModeDir); err != nil {
		return nil, errors.Wrapf(err, "Could not create db directory %s", path)
	}

	p := turtle.GetParser()

	for buildingname, buildingttlfile := range cfg.Buildings {
		cfg.DBPath = filepath.Join(path, buildingname)
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

func (db *MultiDB) RunQueryString(querystring string) (hoddb.QueryResult, error) {
	var emptyres hoddb.QueryResult
	if q, err := query.Parse(querystring); err != nil {
		e := errors.Wrap(err, "Could not parse hod query")
		log.Error(e)
		return emptyres, e
	} else if result, err := db.RunQuery(q); err != nil {
		e := errors.Wrap(err, "Could not complete hod query")
		log.Error(e)
		return emptyres, e
	} else {
		return result, nil
	}
}

func (db *MultiDB) RunQuery(q *sparql.Query) (hoddb.QueryResult, error) {
	var databases = make(map[string]*hoddb.DB)

	// if no FROM clause, then query all dbs!
	if q.From.Empty() {
		q.From.AllDBs = true
	}

	if q.From.AllDBs {
		db.dbs.Range(func(_dbname, _db interface{}) bool {
			dbname := _dbname.(string)
			db := _db.(*hoddb.DB)
			databases[dbname] = db
			return true
		})
	} else {
		for _, dbname := range q.From.Databases {
			db, ok := db.dbs.Load(dbname)
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
		wg.Done()
	}

	wg.Wait()

	//else if result, err := db.RunQuery(q); err != nil {
	//	e := errors.Wrap(err, "Could not complete hod query")
	//	log.Error(e)
	//	return emptyres, e
	//} else {
	//	return result, nil
	//}
	return hoddb.QueryResult{}, nil
}

func (db *MultiDB) LoadDataset(name, ttlfile string) error {
	return nil
}
