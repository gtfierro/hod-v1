package db

import (
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gtfierro/hod/config"
	query "github.com/gtfierro/hod/lang"
	sparql "github.com/gtfierro/hod/lang/ast"
	"github.com/gtfierro/hod/storage"
	"github.com/gtfierro/hod/turtle"

	"github.com/pkg/errors"
	logrus "github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, ForceColors: true})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
}

type graphLoadParams struct {
	name     string
	ttlfiles []string
	done     chan error
}

// HodDB provides versioned access to all building models specified in the provided configuration
type HodDB struct {
	storage storage.StorageProvider
	// store the config so we can make more databases
	cfg        *config.Config
	namespaces map[string]string

	// latest version
	loaded_versions map[storage.Version]*transaction
	sync.RWMutex
}

// NewHodDB creates a new instance of HodDB
func NewHodDB(cfg *config.Config) (*HodDB, error) {
	var hod = &HodDB{
		cfg:             cfg,
		namespaces:      make(map[string]string),
		loaded_versions: make(map[storage.Version]*transaction),
	}

	hod.storage = &storage.BadgerStorageProvider{}
	if err := hod.storage.Initialize(cfg); err != nil {
		return nil, err
	}

	var loadRequests []graphLoadParams

	loadedVersions, err := hod.storage.Graphs()
	if err != nil {
		return nil, err
	}
	logrus.Info("Loaded Versions: ", loadedVersions)

	// load in the ontology files and source file for each building in the config
	// but only if there is no preexisting version for this graph
	if len(loadedVersions) == 0 {
		for buildingname, buildingttlfile := range cfg.Buildings {
			var baseLoad graphLoadParams
			baseLoad.ttlfiles = append(baseLoad.ttlfiles, cfg.Ontologies...)
			baseLoad.ttlfiles = append(baseLoad.ttlfiles, "")
			baseLoad.done = make(chan error)
			latest, existed, err := hod.storage.AddGraph(buildingname)
			if err != nil {
				return nil, err
			}
			logrus.WithFields(logrus.Fields{
				"existed?": existed,
				"version":  latest,
				"filename": buildingttlfile,
				"building": buildingname,
			}).Info("Add graph")
			baseLoad.ttlfiles[len(baseLoad.ttlfiles)-1] = buildingttlfile
			baseLoad.name = buildingname
			loadRequests = append(loadRequests, baseLoad)
		}

		for _, loadreq := range loadRequests {
			if err := hod.loadFiles(loadreq); err != nil {
				logrus.WithFields(logrus.Fields{
					"building": loadreq.name,
					"files":    loadreq.ttlfiles,
					"error":    err,
				}).Error("Load graph")
				// TODO: report on channel
			}
		}
	} else {
		logrus.Info("Continuing from existing databases")
	}
	hod.namespaces, err = hod.storage.GetNamespaces()
	if err != nil {
		return nil, err
	}

	return hod, nil
}

func (hod *HodDB) loadFiles(loadreq graphLoadParams) error {

	tx, err := hod.openTransaction(loadreq.name)
	if err != nil {
		tx.discard()
		return err
	}

	p := turtle.GetParser()
	for _, ttlfile := range loadreq.ttlfiles {
		ds, _ := p.Parse(ttlfile)
		if err := tx.addTriples(ds); err != nil {
			tx.discard()
			return err
		}
		//if err := db.buildTextIndex(ds); err != nil {
		//	return nil, err
		//}
		hod.Lock()
		for abbr, full := range ds.Namespaces {
			if abbr != "" {
				hod.storage.SaveNamespace(abbr, full)
			}
		}
		hod.Unlock()
	}
	return tx.commit()
}

// Close safely closes all of the underlying storage used by HodDB
func (hod *HodDB) Close() error {
	for _, tx := range hod.loaded_versions {
		tx.discard()
	}
	return hod.storage.Close()
}

// RunQuery executes a parsed query against HodDB. This is helpful if you want to avoid
// the parsing overhead for whatever reason
func (hod *HodDB) RunQuery(q *sparql.Query) (result *QueryResult, rerr error) {

	// assemble versions for each of the databases
	totalQueryStart := time.Now()

	// TODO: factor this out
	findDatabasesStart := time.Now()
	var targetGraphs []storage.Version
	if q.From.AllDBs {
		targetGraphs, rerr = hod.storage.Graphs()
		if rerr != nil {
			return
		}
	} else {
		for _, dbname := range q.From.Databases {
			if q.IsInsert() {
				versions, err := hod.storage.ListVersions(dbname)
				if err != nil {
					rerr = err
					return
				}
				targetGraphs = append(targetGraphs, versions[len(versions)-1])
			} else {
				var version storage.Version
				var err error
				switch q.Time.Filter {
				case sparql.AT:
					version, err = hod.storage.VersionAt(dbname, q.Time.Timestamp)
				case sparql.BEFORE:
					version, err = hod.storage.VersionBefore(dbname, q.Time.Timestamp)
				case sparql.AFTER:
					version, err = hod.storage.VersionAfter(dbname, q.Time.Timestamp)
				}
				if err != nil {
					rerr = err
					return
				}
				targetGraphs = append(targetGraphs, version)
			}

		}
	}
	findDatabasesDur := time.Since(findDatabasesStart)

	makeQueriesStart := time.Now()
	// expand out the prefixes
	q.IterTriples(func(triple sparql.Triple) sparql.Triple {
		triple.Subject = hod.expand(triple.Subject)
		triple.Object = hod.expand(triple.Object)
		for idx2, pred := range triple.Predicates {
			triple.Predicates[idx2].Predicate = hod.expand(pred.Predicate)
		}
		return triple
	})

	// expand the graphgroup unions
	var ors [][]sparql.Triple
	if q.Where.GraphGroup != nil {
		for _, group := range q.Where.GraphGroup.Expand() {
			newterms := make([]sparql.Triple, len(q.Where.Terms))
			copy(newterms, q.Where.Terms)
			ors = append(ors, append(newterms, group...))
		}
	}

	var queries []*sparql.Query

	for _, group := range ors {
		tmpQ := q.CopyWithNewTerms(group)
		tmpQ.PopulateVars()
		if tmpQ.Select.AllVars {
			tmpQ.Select.Vars = tmpQ.Variables
		}
		queries = append(queries, &tmpQ)
	}
	if len(queries) == 0 {
		queries = append(queries, q)
	}
	makeQueriesDur := time.Since(makeQueriesStart)

	result = new(QueryResult)
	result.selectVars = q.Select.Vars
	var results []*resultRow

	for _, parsedQuery := range queries {
		// form dependency graph and query plan
		dg := makeDependencyGraph(parsedQuery)
		qp, err := formQueryPlan(dg, parsedQuery)
		if err != nil {
			rerr = err
			return
		}
		for _, op := range qp.operations {
			logrus.Info("op | ", op)
		}

		//parsedQuery.PopulateVars()
		for _, g := range targetGraphs {
			logrus.Info("Run query against> ", g)
			var tx *transaction
			if parsedQuery.IsInsert() {
				tx, err = hod.openTransaction(g.Name)
			} else {
				tx, err = hod.openVersion(g)
			}

			//defer tx.discard()
			if err != nil {
				rerr = err
				return
			}
			// TODO: turn the graph into a transaction
			ctx, err := newQueryContext(qp, tx)
			if err != nil {
				rerr = errors.Wrap(err, "Could not get snapshot")
			}

			for _, op := range ctx.operations {
				//now := time.Now()
				err := op.run(ctx)
				if err != nil {
					rerr = err
					return
				}
			}
			_results := ctx.getResults()
			results = append(results, _results...)

			var intermediateResult = new(QueryResult)
			intermediateResult.fromRows(_results, q.Select.Vars, true)
			if parsedQuery.IsInsert() {
				err := hod.insert(g.Name, parsedQuery.Insert, intermediateResult)
				if err != nil {
					return result, err
				}
			}
		}
	}
	result.fromRows(results, q.Select.Vars, true)
	logrus.WithFields(logrus.Fields{
		"total":       time.Since(totalQueryStart),
		"findversion": findDatabasesDur,
		"makequery":   makeQueriesDur,
	}).Info("Query")
	return
}

// RunQueryString executes a query against HodDB
func (hod *HodDB) RunQueryString(querystring string) (result *QueryResult, err error) {
	var (
		q *sparql.Query
	)
	if q, err = query.Parse(querystring); err != nil {
		err = errors.Wrap(err, "Could not parse hod query")
		logrus.Error(err)
		return
	}

	if result, err = hod.RunQuery(q); err != nil {
		err = errors.Wrap(err, "Could not complete hod query")
		logrus.Error(err)
		return
	}
	return
}

func (hod *HodDB) expand(uri turtle.URI) turtle.URI {
	if !strings.HasPrefix(uri.Value, "?") {
		if full, found := hod.namespaces[uri.Namespace]; found {
			uri.Namespace = full
		}
	}
	return uri
}

func (hod *HodDB) insert(graph string, insert sparql.InsertClause, result *QueryResult) (err error) {
	var additions turtle.DataSet
	//var stats queryStats
	for _, insertTerm := range insert.Terms {
		if result.Count == 0 {
			additions.AddTripleURIs(insertTerm.Subject, insertTerm.Predicates[0].Predicate, insertTerm.Object)
		} else {
			for _, row := range result.Rows {
				newterm := insertTerm.Copy()
				// replace all variables with content from query
				if newterm.Subject.IsVariable() {
					if value, found := row[newterm.Subject.Value]; found {
						newterm.Subject = value
					}
				}
				pred := newterm.Predicates[0].Predicate
				if pred.IsVariable() {
					if value, found := row[pred.Value]; found {
						newterm.Predicates[0].Predicate = value
					}
				}
				if newterm.Object.IsVariable() {
					if value, found := row[newterm.Object.Value]; found {
						newterm.Object = value
					}
				}
				additions.AddTripleURIs(newterm.Subject, newterm.Predicates[0].Predicate, newterm.Object)
				//stats.NumInserted += 1
			}
		}
	}

	tx, err := hod.openTransaction(graph)
	if err != nil {
		tx.discard()
		return err
	}
	if err := tx.addTriples(additions); err != nil {
		tx.discard()
		return err
	}
	if err := tx.commit(); err != nil {
		tx.discard()
		return err
	}
	//if err := hod.buildTextIndex(additions); err != nil {
	//	return err
	//}
	//err = hod.saveIndexes()
	//if err != nil {
	//	return err
	//}
	//stats.InsertTime = time.Since(insertStart)
	//return stats, nil
	return nil
}

func (hod *HodDB) Search(q string, n int) (results []string, err error) {
	return
}

func (hod *HodDB) QueryToClassDOT(q string) (dot string, err error) {
	return
}

func (hod *HodDB) QueryToDOT(q string) (dot string, err error) {
	return
}
