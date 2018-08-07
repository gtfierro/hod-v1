package db

import (
	"os"
	"strings"
	"sync"
	//"time"

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
	sync.RWMutex
}

// NewHodDB creates a new instance of HodDB
func NewHodDB(cfg *config.Config) (*HodDB, error) {
	var hod = &HodDB{
		cfg:        cfg,
		namespaces: make(map[string]string),
	}
	hod.storage = &storage.BadgerStorageProvider{}
	if err := hod.storage.Initialize(cfg); err != nil {
		return nil, err
	}

	var loadRequests []graphLoadParams

	// load in the ontology files and source file for each building in the config
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

	return hod, nil
}

func (hod *HodDB) loadFiles(loadreq graphLoadParams) error {
	hod.Lock()
	defer hod.Unlock()

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
		for abbr, full := range ds.Namespaces {
			if abbr != "" {
				hod.namespaces[abbr] = full
			}
		}
	}
	return tx.commit()
}

// Close safely closes all of the underlying storage used by HodDB
func (hod *HodDB) Close() error {
	return hod.storage.Close()
}

// RunQuery executes a parsed query against HodDB. This is helpful if you want to avoid
// the parsing overhead for whatever reason
func (hod *HodDB) RunQuery(q *sparql.Query) (result *QueryResult, rerr error) {

	// assemble versions for each of the databases
	var targetGraphs []storage.Version
	if q.From.AllDBs {
		targetGraphs, rerr = hod.storage.Graphs()
		if rerr != nil {
			return
		}
	} else {
		for _, dbname := range q.From.Databases {
			versions, err := hod.storage.ListVersions(dbname)
			if err != nil {
				rerr = err
				return
			}
			targetGraphs = append(targetGraphs, versions[len(versions)-1])

		}
	}

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

	// form dependency graph and query plan
	dg := makeDependencyGraph(q)
	qp, err := formQueryPlan(dg, q)
	if err != nil {
		rerr = err
		return
	}
	for _, op := range qp.operations {
		logrus.Info("op | ", op)
	}

	q.PopulateVars()
	result = new(QueryResult)
	result.selectVars = q.Select.Vars
	for _, g := range targetGraphs {
		logrus.Info("Run query against> ", g)
		tx, err := hod.openVersion(g)
		defer tx.discard()
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
		results := ctx.getResults()
		result.fromRows(results, q.Select.Vars, true)

	}
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
