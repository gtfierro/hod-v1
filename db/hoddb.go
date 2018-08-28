package db

import (
	"context"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gtfierro/hod/config"
	query "github.com/gtfierro/hod/lang"
	sparql "github.com/gtfierro/hod/lang/ast"
	"github.com/gtfierro/hod/proto"
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

	versionCaches map[storage.Version]*dbcache
	sync.RWMutex
}

// NewHodDB creates a new instance of HodDB
func NewHodDB(cfg *config.Config) (*HodDB, error) {
	var hod = &HodDB{
		cfg:           cfg,
		namespaces:    make(map[string]string),
		versionCaches: make(map[storage.Version]*dbcache),
	}

	switch cfg.StorageEngine {
	case "memory":
		hod.storage = &storage.MemoryStorageProvider{}
	case "badger":
		fallthrough
	default:
		hod.storage = &storage.BadgerStorageProvider{}
	}

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
func (hod *HodDB) Close() {
	hod.storage.Close()
}

// RunQuery executes a parsed query against HodDB. This is helpful if you want to avoid
// the parsing overhead for whatever reason
func (hod *HodDB) RunQuery(q *sparql.Query) (result *QueryResult, rerr error) {

	return
}

//
// 	if q.IsVersions() {
// 		return hod.resolveVersionQuery(q.Version)
// 	}
//
// 	// assemble versions for each of the databases
// 	totalQueryStart := time.Now()
//
// 	// TODO: factor this out
// 	findDatabasesStart := time.Now()
// 	var targetGraphs []storage.Version
// 	var names []string
// 	if q.From.AllDBs {
// 		_allgraphs, err := hod.storage.Graphs()
// 		if err != nil {
// 			rerr = err
// 			return
// 		}
// 		_n := make(map[string]struct{})
// 		for _, ag := range _allgraphs {
// 			_n[ag.Name] = struct{}{}
// 		}
// 		for n := range _n {
// 			names = append(names, n)
// 		}
// 	} else {
// 		names = q.From.Databases
// 	}
//
// 	for _, dbname := range names {
// 		if q.IsInsert() {
// 			versions, err := hod.storage.ListVersions(dbname)
// 			if err != nil {
// 				rerr = err
// 				return
// 			}
// 			targetGraphs = append(targetGraphs, versions[len(versions)-1])
// 		} else {
// 			switch q.Time.Filter {
// 			case sparql.AT:
// 				_version, err := hod.storage.VersionAt(dbname, q.Time.Timestamp)
// 				rerr = err
// 				targetGraphs = append(targetGraphs, _version)
// 			case sparql.BEFORE:
// 				_versions, err := hod.storage.VersionsBefore(dbname, q.Time.Timestamp, 1)
// 				rerr = err
// 				targetGraphs = append(targetGraphs, _versions...)
// 			case sparql.AFTER:
// 				_versions, err := hod.storage.VersionsAfter(dbname, q.Time.Timestamp, 1)
// 				rerr = err
// 				targetGraphs = append(targetGraphs, _versions...)
// 			}
// 			if rerr != nil {
// 				return
// 			}
// 		}
//
// 	}
// 	//}
// 	findDatabasesDur := time.Since(findDatabasesStart)
//
// 	makeQueriesStart := time.Now()
// 	// expand out the prefixes
// 	q.IterTriples(func(triple sparql.Triple) sparql.Triple {
// 		triple.Subject = hod.expand(triple.Subject)
// 		triple.Object = hod.expand(triple.Object)
// 		for idx2, pred := range triple.Predicates {
// 			triple.Predicates[idx2].Predicate = hod.expand(pred.Predicate)
// 		}
// 		return triple
// 	})
//
// 	// expand the graphgroup unions
// 	var ors [][]sparql.Triple
// 	if q.Where.GraphGroup != nil {
// 		for _, group := range q.Where.GraphGroup.Expand() {
// 			newterms := make([]sparql.Triple, len(q.Where.Terms))
// 			copy(newterms, q.Where.Terms)
// 			ors = append(ors, append(newterms, group...))
// 		}
// 	}
//
// 	var queries []*sparql.Query
//
// 	for _, group := range ors {
// 		tmpQ := q.CopyWithNewTerms(group)
// 		tmpQ.PopulateVars()
// 		if tmpQ.Select.AllVars {
// 			tmpQ.Select.Vars = tmpQ.Variables
// 		}
// 		queries = append(queries, &tmpQ)
// 	}
// 	if len(queries) == 0 {
// 		queries = append(queries, q)
// 	}
// 	makeQueriesDur := time.Since(makeQueriesStart)
//
// 	result = new(QueryResult)
// 	result.selectVars = q.Select.Vars
// 	var results []*resultRow
//
// 	for _, parsedQuery := range queries {
// 		// form dependency graph and query plan
// 		dg := makeDependencyGraph(parsedQuery)
// 		qp, err := formQueryPlan(dg, parsedQuery)
// 		if err != nil {
// 			rerr = err
// 			return
// 		}
// 		for _, op := range qp.operations {
// 			logrus.Info("op | ", op)
// 		}
//
// 		//parsedQuery.PopulateVars()
// 		for _, g := range targetGraphs {
// 			logrus.Info("Run query against> ", g)
// 			var tx *transaction
// 			tx, err = hod.openVersion(g)
//
// 			//defer tx.discard()
// 			if err != nil {
// 				rerr = err
// 				return
// 			}
// 			// TODO: turn the graph into a transaction
// 			ctx, err := newQueryContext(qp, tx)
// 			defer ctx.release()
// 			if err != nil {
// 				rerr = errors.Wrap(err, "Could not get snapshot")
// 			}
//
// 			for _, op := range ctx.operations {
// 				//now := time.Now()
// 				err := op.run(ctx)
// 				if err != nil {
// 					rerr = err
// 					return
// 				}
// 			}
// 			_results := ctx.getResults()
// 			results = append(results, _results...)
//
// 			if parsedQuery.IsInsert() {
// 				var intermediateResult = new(QueryResult)
// 				intermediateResult.fromRows(_results, q.Select.Vars, true)
// 				err := hod.insert(g.Name, parsedQuery.Insert, intermediateResult)
// 				if err != nil {
// 					return result, err
// 				}
// 			}
// 		}
// 	}
// 	result.fromRows(results, q.Select.Vars, true)
// 	logrus.WithFields(logrus.Fields{
// 		"total":       time.Since(totalQueryStart),
// 		"findversion": findDatabasesDur,
// 		"makequery":   makeQueriesDur,
// 	}).Info("Query")
// 	return
// }

// RunQueryString executes a query against HodDB
func (hod *HodDB) RunQueryString(querystring string) (result *proto.QueryResponse, err error) {
	return hod.ExecuteQuery(context.Background(), &proto.QueryRequest{Query: querystring})
}

func (hod *HodDB) expand(uri turtle.URI) turtle.URI {
	if !strings.HasPrefix(uri.Value, "?") {
		if full, found := hod.namespaces[uri.Namespace]; found {
			uri.Namespace = full
		}
	}
	return uri
}

func (hod *HodDB) insert(graph string, insert sparql.InsertClause, result *proto.QueryResponse) (resp *proto.QueryResponse, err error) {
	builder := newResultBuilder([]string{"subject", "predicate", "object"})
	var additions turtle.DataSet
	//var stats queryStats
	var positions = make(map[string]int)
	for pos, varname := range result.Variable {
		positions[varname] = pos
	}
	for _, insertTerm := range insert.Terms {
		if result.Count == 0 {
			additions.AddTripleURIs(insertTerm.Subject, insertTerm.Predicates[0].Predicate, insertTerm.Object)
		} else {
			for _, row := range result.Rows {
				newterm := insertTerm.Copy()
				// replace all variables with content from query
				if newterm.Subject.IsVariable() {
					if pos, found := positions[newterm.Subject.Value]; found {
						newterm.Subject = turtle.URI{row.Uris[pos].Namespace, row.Uris[pos].Value}
					}
				}
				pred := newterm.Predicates[0].Predicate
				if pred.IsVariable() {
					if pos, found := positions[pred.Value]; found {
						newterm.Predicates[0].Predicate = turtle.URI{row.Uris[pos].Namespace, row.Uris[pos].Value}
					}
				}
				if newterm.Object.IsVariable() {
					if pos, found := positions[newterm.Object.Value]; found {
						newterm.Object = turtle.URI{row.Uris[pos].Namespace, row.Uris[pos].Value}
					}
				}
				additions.AddTripleURIs(newterm.Subject, newterm.Predicates[0].Predicate, newterm.Object)
				builder.addRowString([]string{newterm.Subject.String(), newterm.Predicates[0].Predicate.String(), newterm.Object.String()})
				//stats.NumInserted += 1
			}
		}
	}

	tx, err := hod.openTransaction(graph)
	if err != nil {
		tx.discard()
		return builder.finish(), err
	}
	if err := tx.addTriples(additions); err != nil {
		tx.discard()
		return builder.finish(), err
	}
	if err := tx.commit(); err != nil {
		tx.discard()
		return builder.finish(), err
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
	return builder.finish(), nil
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

func (hod *HodDB) Names() ([]string, error) {
	return hod.storage.Names()
}

func (hod *HodDB) AllVersions() ([]storage.Version, error) {
	return hod.storage.Graphs()
}

func (hod *HodDB) VersionAt(name string, t time.Time) (storage.Version, error) {
	return hod.storage.VersionAt(name, t)
}

func (hod *HodDB) VersionAfter(name string, t time.Time, limit int) ([]storage.Version, error) {
	return hod.storage.VersionsAfter(name, t, limit)
}

func (hod *HodDB) VersionBefore(name string, t time.Time, limit int) ([]storage.Version, error) {
	return hod.storage.VersionsBefore(name, t, limit)
}

func (hod *HodDB) ExecuteQuery(ctx context.Context, request *proto.QueryRequest) (response *proto.QueryResponse, rerr error) {
	totalQueryStart := time.Now()
	defer ctx.Done()

	// TODO: factor this out
	findDatabasesStart := time.Now()
	var targetGraphs []storage.Version
	var names []string

	if request == nil {
		rerr = errors.New("No query provided")
		return
	}

	q, err := query.Parse(request.Query)
	if err != nil {
		rerr = errors.Wrap(err, "could not parse query")
		return
	}

	if q.IsVersions() {
		return hod.resolveVersionQuery(q.Version)
	}

	if q.From.AllDBs {
		_allgraphs, err := hod.storage.Graphs()
		if err != nil {
			rerr = err
			return
		}
		_n := make(map[string]struct{})
		for _, ag := range _allgraphs {
			_n[ag.Name] = struct{}{}
		}
		for n := range _n {
			names = append(names, n)
		}
	} else {
		names = q.From.Databases
	}

	for _, dbname := range names {
		if q.IsInsert() {
			versions, err := hod.storage.ListVersions(dbname)
			if err != nil {
				rerr = err
				return
			}
			targetGraphs = append(targetGraphs, versions[len(versions)-1])
		} else {
			switch q.Time.Filter {
			case sparql.AT:
				_version, err := hod.storage.VersionAt(dbname, q.Time.Timestamp)
				rerr = err
				targetGraphs = append(targetGraphs, _version)
			case sparql.BEFORE:
				_versions, err := hod.storage.VersionsBefore(dbname, q.Time.Timestamp, 1)
				rerr = err
				targetGraphs = append(targetGraphs, _versions...)
			case sparql.AFTER:
				_versions, err := hod.storage.VersionsAfter(dbname, q.Time.Timestamp, 1)
				rerr = err
				targetGraphs = append(targetGraphs, _versions...)
			}
			if rerr != nil {
				return
			}
		}

	}
	//}
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

	builder := newResultBuilder(q.Select.Vars)

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
			tx, err = hod.openVersion(g)

			//defer tx.discard()
			if err != nil {
				rerr = err
				return
			}
			// TODO: turn the graph into a transaction
			ctx, err := newQueryContext(qp, tx)
			defer ctx.release()
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

			builder.addRowsFrom(ctx)

			if parsedQuery.IsInsert() {
				//var intermediateResult = new(QueryResult)
				//intermediateResult.fromRows(_results, q.Select.Vars, true)
				return hod.insert(g.Name, parsedQuery.Insert, builder.finish())
			}
		}
	}

	response = builder.finish()
	logrus.WithFields(logrus.Fields{
		"total":       time.Since(totalQueryStart),
		"findversion": findDatabasesDur,
		"makequery":   makeQueriesDur,
	}).Info("Query")
	return response, nil
}

func (hod *HodDB) resolveVersionQuery(q sparql.VersionsQuery) (response *proto.QueryResponse, rerr error) {
	if len(q.Names.Databases) == 0 && !q.Names.AllDBs {
		builder := newResultBuilder([]string{"name"})
		var names []string
		names, rerr = hod.Names()
		if rerr != nil {
			return
		}
		builder.addRowString(names)
		response = builder.finish()
		return
	}
	builder := newResultBuilder([]string{"name", "version"})

	if q.Names.AllDBs {
		q.Names.Databases, rerr = hod.Names()
		if rerr != nil {
			return
		}
	}

	for _, dbname := range q.Names.Databases {
		switch q.Filter.Filter {
		case sparql.AT:
			_version, err := hod.VersionAt(dbname, q.Filter.Timestamp)
			rerr = err
			builder.addRowString([]string{_version.Name, time.Unix(0, int64(_version.Timestamp)).Format(time.RFC3339)})
		case sparql.BEFORE:
			_versions, err := hod.VersionBefore(dbname, q.Filter.Timestamp, q.Limit)
			rerr = err
			for _, _version := range _versions {
				builder.addRowString([]string{_version.Name, time.Unix(0, int64(_version.Timestamp)).Format(time.RFC3339)})
			}
		case sparql.AFTER:
			_versions, err := hod.VersionAfter(dbname, q.Filter.Timestamp, q.Limit)
			rerr = err
			for _, _version := range _versions {
				builder.addRowString([]string{_version.Name, time.Unix(0, int64(_version.Timestamp)).Format(time.RFC3339)})
			}
		}
		if rerr != nil {
			return
		}
	}
	response = builder.finish()
	return
}
