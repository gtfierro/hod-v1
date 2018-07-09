package db

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/gtfierro/hod/config"
	"github.com/gtfierro/hod/storage"
	"github.com/gtfierro/hod/turtle"

	"github.com/blevesearch/bleve"
	"github.com/coocood/freecache"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
)

var emptyKey = Key{}

type DB struct {
	path      string
	name      string
	backing   storage.StorageProvider
	predIndex map[turtle.URI]*PredicateEntity
	// store relationships and their inverses
	relationships map[turtle.URI]turtle.URI
	relLock       sync.RWMutex
	// stores which edges can be 'rolled forward' in the index
	transitiveEdges map[turtle.URI]struct{}
	// store the namespace prefixes as strings
	namespaces map[string]string
	// config options for output
	showDependencyGraph    bool
	showQueryPlan          bool
	showQueryPlanLatencies bool
	showOperationLatencies bool
	showQueryLatencies     bool
	// cache for query results
	queryCache        *freecache.Cache
	queryCacheEnabled bool
	loading           bool

	cache *dbcache

	// text index
	textidx bleve.Index
}

func newDB(name string, cfg *config.Config) (*DB, error) {
	store := &storage.LevelDBStorageProvider{}
	//store := &storage.RedisStorageProvider{}
	if err := store.Initialize(name, cfg); err != nil {
		return nil, err
	}
	mapping := bleve.NewIndexMapping()
	path := strings.TrimSuffix(cfg.DBPath, "/")
	index, err := bleve.New(path+"/myExampleIndex.bleve", mapping)
	if err != nil && err == bleve.ErrorIndexPathExists {
		index, err = bleve.Open(path + "/myExampleIndex.bleve")
	}
	if err != nil {
		return nil, errors.Wrapf(err, "Could not open bleve index %s", path+"/myExampleIndex.bleve")
	}

	db := &DB{
		path:                   path,
		name:                   name,
		backing:                store,
		predIndex:              make(map[turtle.URI]*PredicateEntity),
		relationships:          make(map[turtle.URI]turtle.URI),
		transitiveEdges:        make(map[turtle.URI]struct{}),
		namespaces:             make(map[string]string),
		showDependencyGraph:    cfg.ShowDependencyGraph,
		showQueryPlan:          cfg.ShowQueryPlan,
		showQueryPlanLatencies: cfg.ShowQueryPlanLatencies,
		showOperationLatencies: cfg.ShowOperationLatencies,
		showQueryLatencies:     cfg.ShowQueryLatencies,
		queryCacheEnabled:      !cfg.DisableQueryCache,
		loading:                false,
		textidx:                index,
		cache:                  newCache(16),
	}

	if db.queryCacheEnabled {
		db.queryCache = freecache.NewCache(64 * 1024 * 1024) // 64 MB
	}

	// load predIndex and relationships from database
	predIndexPath := path + "/predIndex"
	relshipIndexPath := path + "/relshipIndex"
	namespaceIndexPath := path + "/namespaceIndex"
	if _, err := os.Stat(predIndexPath); !os.IsNotExist(err) {
		f, err := os.Open(predIndexPath)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not open predIndex file %s", predIndexPath)
		}
		var pi = new(PredIndex)
		if err := msgp.Decode(f, pi); err != nil {
			return nil, err
		}
		for uri, pe := range *pi {
			db.predIndex[turtle.ParseURI(uri)] = pe
		}
	}
	if _, err := os.Stat(relshipIndexPath); !os.IsNotExist(err) {
		f, err := os.Open(relshipIndexPath)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not open relshipIndexPath file %s", relshipIndexPath)
		}
		var ri = new(RelshipIndex)
		if err := msgp.Decode(f, ri); err != nil {
			return nil, err
		}
		for uri, uri2 := range *ri {
			db.relationships[turtle.ParseURI(uri)] = turtle.ParseURI(uri2)
		}
	}
	if _, err := os.Stat(namespaceIndexPath); !os.IsNotExist(err) {
		f, err := os.Open(namespaceIndexPath)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not open namespaceIndexPath file %s", namespaceIndexPath)
		}
		var ni = new(NamespaceIndex)
		if err := msgp.Decode(f, ni); err != nil {
			return nil, err
		}
		for ns, full := range *ni {
			db.namespaces[ns] = full
		}
	}

	// load in Brick
	if cfg.ReloadOntologies {
		p := turtle.GetParser()
		for _, ontologyFile := range cfg.Ontologies {
			ds, _ := p.Parse(ontologyFile)

			tx, err := db.openTransaction()
			if err != nil {
				log.Error(err)
				tx.discard()
				panic(err)
			}
			if err := tx.addTriples(ds); err != nil {
				tx.discard()
				panic(err)
			}

			if err := tx.done(); err != nil {
				tx.discard()
				return nil, err
			}
			if err := db.buildTextIndex(ds); err != nil {
				return nil, err
			}

			for abbr, full := range ds.Namespaces {
				if abbr != "" {
					db.namespaces[abbr] = full
				}
			}

		}
		err = db.saveIndexes()
		if err != nil {
			return nil, err
		}
	}

	//if cfg.ReloadOntologies {
	//	p := turtle.GetParser()
	//	for _, ontologyFile := range cfg.Ontologies {
	//		ds, _ := p.Parse(ontologyFile)
	//		err = db.loadDataset(ds)

	//		if err != nil {
	//			return nil, err
	//		}
	//	}
	//	err = db.saveIndexes()
	//	if err != nil {
	//		return nil, err
	//	}
	//}

	if cfg.ShowNamespaces {
		var dmp strings.Builder
		lenK := len("Prefix")
		lenV := len("Namespace")

		for k, v := range db.namespaces {
			if len(k) > lenK {
				lenK = len(k)
			}
			if len(v) > lenV {
				lenV = len(v)
			}
		}
		fmt.Fprintf(&dmp, "+ %s +\n", strings.Repeat("-", lenK+lenV+3))
		fmt.Fprintf(&dmp, "| Prefix%s | Namespace%s |\n", strings.Repeat(" ", lenK-len("Prefix")), strings.Repeat(" ", lenV-len("Namespace")))
		fmt.Fprintf(&dmp, "+ %s +\n", strings.Repeat("-", lenK+lenV+3))
		for k, v := range db.namespaces {
			kpad := strings.Repeat(" ", lenK-len(k))
			vpad := strings.Repeat(" ", lenV-len(v))
			fmt.Fprintf(&dmp, "| %s%s | %s%s |\n", k, kpad, v, vpad)
		}
		fmt.Fprintf(&dmp, "+ %s +\n", strings.Repeat("-", lenK+lenV+3))
		fmt.Println(dmp.String())

	}

	return db, nil
}

func (db *DB) Close() {
	if err := db.backing.Close(); err != nil {
		log.Fatal(err)
	}
	if err := db.textidx.Close(); err != nil {
		log.Fatal(err)
	}
}

func (db *DB) buildTextIndex(dataset turtle.DataSet) error {
	b := db.textidx.NewBatch()
	for _, triple := range dataset.Triples {
		// add classes to the text idx
		if triple.Predicate.String() == "http://www.w3.org/1999/02/22-rdf-syntax-ns#type" && triple.Object.String() == "http://www.w3.org/2002/07/owl#Class" && triple.Subject.Namespace != "" {
			sub := strings.Replace(triple.Subject.Value, "_", " ", -1)
			if err := b.Index(triple.Subject.String(), sub); err != nil && len(triple.Subject.String()) > 0 {
				return errors.Wrapf(err, "Could not add subject %s to text index (%s)", triple.Subject, triple)
			}
		}
	}
	if err := db.textidx.Batch(b); err != nil {
		return errors.Wrap(err, "Could not save batch text index")
	}
	return nil
}

func (db *DB) saveIndexes() error {
	f, err := os.Create(db.path + "/predIndex")
	if err != nil {
		return err
	}

	pi := make(PredIndex)
	for uri, pe := range db.predIndex {
		pi[uri.String()] = pe
	}

	if err := msgp.Encode(f, pi); err != nil {
		return err
	}

	f, err = os.Create(db.path + "/relshipIndex")
	if err != nil {
		return err
	}

	ri := make(RelshipIndex)
	for uri, uri2 := range db.relationships {
		ri[uri.String()] = uri2.String()
	}

	if err := msgp.Encode(f, ri); err != nil {
		return err
	}

	f, err = os.Create(db.path + "/namespaceIndex")
	if err != nil {
		return err
	}
	if err := msgp.Encode(f, NamespaceIndex(db.namespaces)); err != nil {
		return err
	}

	return nil
}
