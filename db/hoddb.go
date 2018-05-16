package db

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gtfierro/hod/config"
	query "github.com/gtfierro/hod/lang"
	sparql "github.com/gtfierro/hod/lang/ast"
	"github.com/gtfierro/hod/turtle"
	logrus "github.com/sirupsen/logrus"

	"github.com/mitghi/btree"
	"github.com/pkg/errors"
)

type HodDB struct {
	buildings []string
	// database name => *db.DB
	dbs sync.Map
	// filename => sha256 hash
	loadedfilehashes map[string][]byte
	// store the config so we can make more databases
	cfg   *config.Config
	dbdir string
}

// Creates or loads a new instance of HodDB from the provided config file. If any of the Turtle source files
// in the "buildings" section have changed, HodDB will load them anew.
func NewHodDB(cfg *config.Config) (*HodDB, error) {
	var hod = &HodDB{
		cfg:              cfg,
		loadedfilehashes: make(map[string][]byte),
	}

	// create path for dbs
	hod.dbdir = strings.TrimSuffix(cfg.DBPath, "/")
	if err := os.MkdirAll(hod.dbdir, 0700); err != nil {
		return nil, errors.Wrapf(err, "Could not create db directory %s", hod.dbdir)
	}

	fileHashPath := filepath.Join(hod.dbdir, "fileHashes")
	if _, err := os.Stat(fileHashPath); !os.IsNotExist(err) {
		f, err := os.Open(fileHashPath)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not open fileHash %s", fileHashPath)
		}
		dec := json.NewDecoder(f)
		if err := dec.Decode(&hod.loadedfilehashes); err != nil {
			return nil, errors.Wrapf(err, "Could not decode fileHash %s", fileHashPath)
		}
	}

	// load files.
	// For each file, we compute the sha256 hash. If we have already loaded the file and
	// it hasn't changed, the hash should be in hod.loadedfilehashes

	for buildingname, buildingttlfile := range cfg.Buildings {
		f, err := os.Open(buildingttlfile)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not read input file %s", buildingttlfile)
		}
		defer f.Close()
		filehasher := sha256.New()
		if _, err := io.Copy(filehasher, f); err != nil {
			return nil, errors.Wrapf(err, "Could not hash file %s", buildingttlfile)
		}
		filehash := filehasher.Sum(nil)
		if existinghash, found := hod.loadedfilehashes[buildingttlfile]; found && bytes.Equal(filehash, existinghash) {
			log.Infof("TTL file %s has not changed since we last loaded it! Skipping...", buildingttlfile)
			cfg.ReloadOntologies = false
			cfg.DBPath = filepath.Join(hod.dbdir, buildingname)
			db, err := newDB(buildingname, cfg)
			if err != nil {
				return nil, errors.Wrap(err, "Could not load existing database")
			}
			hod.dbs.Store(buildingname, db)
			hod.buildings = append(hod.buildings, buildingname)
			continue
		}
		hod.buildings = append(hod.buildings, buildingname)
		hod.loadedfilehashes[buildingttlfile] = filehash

		if err := hod.loadDataset(buildingname, buildingttlfile); err != nil {
			return nil, err
		}
	}

	if err := hod.saveIndexes(); err != nil {
		return nil, errors.Wrap(err, "Could not save file indexes")
	}

	return hod, nil
}

func (hod *HodDB) saveIndexes() error {
	f, err := os.Create(filepath.Join(hod.dbdir, "fileHashes"))
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	return enc.Encode(hod.loadedfilehashes)
}

// Execute the provided query against HodDB
func (hod *HodDB) RunQueryString(querystring string) (result QueryResult, err error) {
	var (
		q *sparql.Query
	)
	if q, err = query.Parse(querystring); err != nil {
		err = errors.Wrap(err, "Could not parse hod query")
		log.Error(err)
		return
	}

	if result, err = hod.RunQuery(q); err != nil {
		err = errors.Wrap(err, "Could not complete hod query")
		log.Error(err)
		return
	}
	return
}

// List the databases loaded into HodDB by name
func (hod *HodDB) Databases() []string {
	return hod.buildings
}

// Execute a parsed query against HodDB
func (hod *HodDB) RunQuery(q *sparql.Query) (QueryResult, error) {
	var databases = make(map[string]*DB)
	fullQueryStart := time.Now()

	if q.From.AllDBs {
		hod.dbs.Range(func(_dbname, _db interface{}) bool {
			dbname := _dbname.(string)
			db := _db.(*DB)
			databases[dbname] = db
			return true
		})
	} else {
		for _, dbname := range q.From.Databases {
			db, ok := hod.dbs.Load(dbname)
			if ok {
				databases[dbname] = db.(*DB)
			}
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(databases))
	//var rowlock sync.Mutex
	unionedRows := btree.New(4, "")
	var result QueryResult
	result.selectVars = q.Select.Vars
	var stats = new(queryStats)

	for dbname, db := range databases {
		//go func() {

		// handle SELECT query
		//if q.IsSelect() {
		singleresult, _stats, err := db.runQuery(q)
		stats.merge(_stats)
		if err != nil {
			log.Error(errors.Wrapf(err, "Error running query on %s", dbname))
		}
		//rowlock.Lock()
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
		//}

		// handle INSERT query
		if q.IsInsert() {
			insertstats, err := db.handleInsert(q.Insert, result)
			if err != nil {
				return result, err
			}
			stats.merge(insertstats)
		}
		if !q.IsSelect() {
			result.Rows = result.Rows[:0]
		}
		//rowlock.Unlock()
		//TODO: merge these or decide how to grouop them
		wg.Done()
		//}()
	}

	logrus.WithFields(logrus.Fields{
		"SelectVars": q.Select.Vars,
		"#Results":   stats.NumResults,
		"#Inserted":  stats.NumInserted,
		"#Deleted":   stats.NumDeleted,
		"Insert":     stats.InsertTime,
		"Where":      stats.WhereTime,
		"Expand":     stats.ExpandTime,
		"Total":      time.Since(fullQueryStart),
	}).Info("Query")

	wg.Wait()

	return result, nil
}

func (hod *HodDB) loadDataset(name, ttlfile string) error {
	hod.cfg.DBPath = filepath.Join(hod.dbdir, name)
	hod.cfg.ReloadOntologies = true
	db, err := newDB(name, hod.cfg)
	if err != nil {
		return errors.Wrapf(err, "Could not create database at %s", hod.cfg.DBPath)
	}
	p := turtle.GetParser()
	ds, duration := p.Parse(ttlfile)
	rate := float64((float64(ds.NumTriples()) / float64(duration.Nanoseconds())) * 1e9)
	log.Infof("Loaded %d triples, %d namespaces in %s (%.0f/sec)", ds.NumTriples(), ds.NumNamespaces(), duration, rate)
	tx, err := db.openTransaction()
	if err != nil {
		tx.discard()
		return err
	}
	if err := tx.addTriples(ds); err != nil {
		tx.discard()
		return err
	}

	if err := tx.commit(); err != nil {
		tx.discard()
		return err
	}
	for abbr, full := range ds.Namespaces {
		if abbr != "" {
			db.namespaces[abbr] = full
		}
	}
	if err = db.saveIndexes(); err != nil {
		return err
	}
	//err = db.loadDataset(ds)
	//if err != nil {
	//	return errors.Wrapf(err, "Could not load dataset %s", ttlfile)
	//}
	hod.dbs.Store(name, db)
	return nil
}

// Close HodDB
func (hod *HodDB) Close() {
	hod.dbs.Range(func(_dbname, _db interface{}) bool {
		db := _db.(*DB)
		db.Close()
		return true
	})
}

// Wildcard search using Bleve through all values in the database
func (hod *HodDB) Search(q string, n int) ([]string, error) {
	// just pick first db for now
	var (
		res []string
		err error
	)
	hod.dbs.Range(func(_dbname, _db interface{}) bool {
		db := _db.(*DB)
		res, err = db.search(q, n)
		if err != nil {
			return true
		}
		return false
	})
	return res, err
}

// Turn the results of the query into a GraphViz visualization of the classes and their relationships
func (hod *HodDB) QueryToClassDOT(q string) (string, error) {
	// just pick first db for now
	var (
		res string
		err error
	)

	if err != nil {
		return "", err
	}
	// create DOT template string
	dot := ""
	dot += "digraph G {\n"

	hod.dbs.Range(func(_dbname, _db interface{}) bool {
		db := _db.(*DB)
		res, err = db.queryToClassDOT(q)
		dot += res
		if err != nil {
			return true
		}
		return true
	})
	dot += "}"
	fmt.Println(dot)
	return res, err
}

// Turn the results of the query into a GraphViz visualization of the results
func (hod *HodDB) QueryToDOT(q string) (string, error) {
	// just pick first db for now
	var (
		res string
		err error
	)
	hod.dbs.Range(func(_dbname, _db interface{}) bool {
		db := _db.(*DB)
		res, err = db.queryToDOT(q)
		if err != nil {
			return true
		}
		return false
	})
	return res, err
}

func (hod *HodDB) abbreviate(uri turtle.URI) string {
	// just pick first db for now
	var (
		res string
		err error
	)
	hod.dbs.Range(func(_dbname, _db interface{}) bool {
		db := _db.(*DB)
		res = db.abbreviate(uri)
		if err != nil {
			return true
		}
		return false
	})
	return res
}
