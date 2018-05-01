package db

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	query "github.com/gtfierro/hod/lang"
	sparql "github.com/gtfierro/hod/lang/ast"
	"github.com/gtfierro/hod/turtle"
	logrus "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/blevesearch/bleve"
	"github.com/coocood/freecache"
	"github.com/kr/pretty"
	"github.com/mitghi/btree"
	"github.com/pkg/errors"
)

func prettyprint(v interface{}) {
	fmt.Printf("%# v", pretty.Formatter(v))
}

func (db *DB) runQueryString(q string) (QueryResult, error) {
	var emptyres QueryResult
	if q, err := query.Parse(q); err != nil {
		e := errors.Wrap(err, "Could not parse hod query")
		log.Error(e)
		return emptyres, e
	} else if result, err := db.runQuery(q); err != nil {
		e := errors.Wrap(err, "Could not complete hod query")
		log.Error(e)
		return emptyres, e
	} else {
		return result, nil
	}
}

func (db *DB) runQuery(q *sparql.Query) (QueryResult, error) {
	fullQueryStart := time.Now()

	// "clean" the query by expanding out the prefixes
	// make sure to first do the Filters, then the Or clauses
	q.IterTriples(func(triple sparql.Triple) sparql.Triple {
		if !strings.HasPrefix(triple.Subject.Value, "?") {
			if full, found := db.namespaces[triple.Subject.Namespace]; found {
				triple.Subject.Namespace = full
			}
		}
		if !strings.HasPrefix(triple.Object.Value, "?") {
			if full, found := db.namespaces[triple.Object.Namespace]; found {
				triple.Object.Namespace = full
			}
		}
		for idx2, pred := range triple.Predicates {
			if !strings.HasPrefix(pred.Predicate.Value, "?") {
				if full, found := db.namespaces[pred.Predicate.Namespace]; found {
					pred.Predicate.Namespace = full
				}
				triple.Predicates[idx2] = pred
			}
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

	// check query hash
	var queryhash []byte
	if db.queryCacheEnabled {
		queryhash = hashQuery(q)
		if ans, err := db.queryCache.Get(queryhash); err == nil {
			var res QueryResult
			if _, err := res.UnmarshalMsg(ans); err != nil {
				log.Error(errors.Wrap(err, "Could not fetch query from cache. Running..."))
			} else {
				// successful!
				res.Elapsed = time.Since(fullQueryStart)
				return res, nil
			}
		} else if err != nil && err == freecache.ErrNotFound {
			log.Notice("Could not fetch query from cache")
		} else if err != nil {
			log.Error(errors.Wrap(err, "Could not access query cache"))
		}
	}

	unionedRows := btree.New(BTREE_DEGREE, "")
	defer cleanResultRows(unionedRows)

	// if we have terms that are part of a set of OR statements, then we run
	// parallel queries for each fully-elaborated "branch" or the OR statement,
	// and then merge the results together at the end
	var stats *queryStats
	if len(ors) > 0 {
		var rowLock sync.Mutex
		var wg sync.WaitGroup
		var queryErr error
		wg.Add(len(ors))
		for _, group := range ors {
			tmpQuery := q.CopyWithNewTerms(group)
			tmpQuery.PopulateVars()
			if tmpQuery.Select.AllVars {
				tmpQuery.Select.Vars = tmpQuery.Variables
			}

			go func(q *sparql.Query) {
				results, _stats, err := db.getQueryResults(&tmpQuery)
				rowLock.Lock()

				if err != nil {
					queryErr = err
				} else {
					stats.merge(_stats)
					log.Debug("got", len(results))
					for _, row := range results {
						unionedRows.ReplaceOrInsert(row)
					}
				}
				rowLock.Unlock()
				wg.Done()
			}(&tmpQuery)
		}
		wg.Wait()
		if queryErr != nil {
			return QueryResult{}, queryErr
		}
	} else {
		q.PopulateVars()
		if q.Select.AllVars {
			q.Select.Vars = q.Variables
		}
		results, _stats, err := db.getQueryResults(q)
		if err != nil {
			return QueryResult{}, err
		}
		stats = &_stats
		for _, row := range results {
			unionedRows.ReplaceOrInsert(row)
		}
	}
	if stats != nil {
		logrus.WithFields(logrus.Fields{
			"Where":   q.Select.Vars,
			"Execute": stats.ExecutionTime,
			"Expand":  stats.ExpandTime,
			"Results": stats.NumResults,
			"Total":   time.Since(fullQueryStart),
		}).Info("Query")
	} else {
		logrus.WithFields(logrus.Fields{
			"Where": q.Select.Vars,
			"Total": time.Since(fullQueryStart),
		}).Info("Query")
	}

	var result = newQueryResult()
	result.selectVars = q.Select.Vars
	result.Elapsed = time.Since(fullQueryStart)

	// return the rows
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

	if db.queryCacheEnabled {
		// set this in the cache
		marshalled, err := result.MarshalMsg(nil)
		if err != nil {
			log.Error(errors.Wrap(err, "Could not marshal results"))
		}
		if err := db.queryCache.Set(queryhash, marshalled, -1); err != nil {
			log.Error(errors.Wrap(err, "Could not cache results"))
		}
	}

	return result, nil
}

func (db *DB) runQueryToSet(q *sparql.Query) ([]*ResultRow, error) {
	var result []*ResultRow

	fullQueryStart := time.Now()

	// "clean" the query by expanding out the prefixes
	// make sure to first do the Filters, then the Or clauses
	q.IterTriples(func(triple sparql.Triple) sparql.Triple {
		if !strings.HasPrefix(triple.Subject.Value, "?") {
			if full, found := db.namespaces[triple.Subject.Namespace]; found {
				triple.Subject.Namespace = full
			}
		}
		if !strings.HasPrefix(triple.Object.Value, "?") {
			if full, found := db.namespaces[triple.Object.Namespace]; found {
				triple.Object.Namespace = full
			}
		}
		for idx2, pred := range triple.Predicates {
			if !strings.HasPrefix(pred.Predicate.Value, "?") {
				if full, found := db.namespaces[pred.Predicate.Namespace]; found {
					pred.Predicate.Namespace = full
				}
				triple.Predicates[idx2] = pred
			}
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

	// if we have terms that are part of a set of OR statements, then we run
	// parallel queries for each fully-elaborated "branch" or the OR statement,
	// and then merge the results together at the end
	var stats *queryStats
	if len(ors) > 0 {
		var rowLock sync.Mutex
		var wg sync.WaitGroup
		var queryErr error
		wg.Add(len(ors))
		for _, group := range ors {
			tmpQuery := q.CopyWithNewTerms(group)
			tmpQuery.PopulateVars()
			if tmpQuery.Select.AllVars {
				tmpQuery.Select.Vars = tmpQuery.Variables
			}

			go func(q *sparql.Query) {
				results, _stats, err := db.getQueryResults(&tmpQuery)
				rowLock.Lock()
				if err != nil {
					queryErr = err
				} else {
					stats.merge(_stats)
					result = append(result, results...)
				}
				rowLock.Unlock()
				wg.Done()
			}(&tmpQuery)
		}
		wg.Wait()
		if queryErr != nil {
			return result, queryErr
		}
	} else {
		q.PopulateVars()
		if q.Select.AllVars {
			q.Select.Vars = q.Variables
		}
		results, _stats, err := db.getQueryResults(q)
		stats = &_stats
		if err != nil {
			return result, err
		}
		result = append(result, results...)
	}
	if stats != nil {
		logrus.WithFields(logrus.Fields{
			"Where":   q.Select.Vars,
			"Execute": stats.ExecutionTime,
			"Expand":  stats.ExpandTime,
			"Results": stats.NumResults,
			"Total":   time.Since(fullQueryStart),
		}).Info("Query")
	} else {
		logrus.WithFields(logrus.Fields{
			"Where": q.Select.Vars,
			"Total": time.Since(fullQueryStart),
		}).Info("Query")
	}
	return result, nil
}

// takes a query and returns a DOT representation to visualize
// the construction of the query
func (db *DB) queryToDOT(querystring string) (string, error) {
	q, err := query.Parse(querystring)
	if err != nil {
		return "", err
	}

	dot := ""
	dot += "digraph G {\n"
	dot += "ratio=\"auto\"\n"
	dot += "rankdir=\"LR\"\n"
	dot += "size=\"7.5,10\"\n"

	//if len(q.Where.Ors) > 0 {
	//	orTerms := query.FlattenOrClauseList(q.Where.Ors)
	//	oldFilters := q.Where.Filters
	//	for _, orTerm := range orTerms {
	//		filters := append(oldFilters, orTerm...)
	//		for _, filter := range filters {
	//			var parts []string
	//			for _, p := range filter.Path {
	//				parts = append(parts, fmt.Sprintf("%s%s", p.Predicate, p.Pattern))
	//			}
	//			line := fmt.Sprintf("\"%s\" -> \"%s\" [label=\"%s\"];\n", filter.Subject, filter.Object, strings.Join(parts, "/"))
	//			if !strings.Contains(dot, line) {
	//				dot += line
	//			}

	//		}
	//	}
	//} else {
	for _, filter := range q.Where.Terms {
		var parts []string
		for _, p := range filter.Predicates {
			parts = append(parts, fmt.Sprintf("%s%s", p.Predicate, p.Pattern))
		}
		line := fmt.Sprintf("\"%s\" -> \"%s\" [label=\"%s\"];\n", filter.Subject, filter.Object, strings.Join(parts, "/"))
		if !strings.Contains(dot, line) {
			dot += line
		}
	}
	//}
	for _, sv := range q.Select.Vars {
		dot += fmt.Sprintf("\"%s\" [fillcolor=#e57373]\n", sv)
	}
	dot += "}"
	return dot, nil
}

// executes a query and returns a DOT string of the classes involved
func (db *DB) queryToClassDOT(querystring string) (string, error) {
	q, err := query.Parse(querystring)
	if err != nil {
		return "", err
	}
	// create DOT template string
	dot := ""

	// get rdf:type predicate hash as a string
	typeURI := turtle.ParseURI("rdf:type")
	typeURI.Namespace = db.namespaces[typeURI.Namespace]
	typeKey, err := db.GetHash(typeURI)
	if err != nil {
		return "", err
	}
	typeKeyString := typeKey.String()

	getClass := func(ent *Entity) (classes []turtle.URI, err error) {
		_classes := ent.OutEdges[typeKeyString]
		for _, class := range _classes {
			classes = append(classes, db.MustGetURI(class))
		}
		return
	}

	getEdges := func(ent *Entity) (predicates, objects []turtle.URI, reterr error) {
		var predKey Key
		for predKeyString, objectList := range ent.OutEdges {
			predKey.FromSlice([]byte(predKeyString))
			predURI, err := db.GetURI(predKey)
			if err != nil {
				reterr = err
				return
			}
			for _, objectKey := range objectList {
				objectEnt, err := db.GetEntityFromHash(objectKey)
				if err != nil {
					reterr = err
					return
				}
				objectClasses, err := getClass(objectEnt)
				if err != nil {
					reterr = err
					return
				}
				for _, class := range objectClasses {
					if predURI.Value != "type" && class.Value != "Class" {
						predicates = append(predicates, predURI)
						objects = append(objects, class)
					}
				}

			}
		}
		return
	}

	result, err := db.runQuery(q)
	if err != nil {
		return "", err
	}
	for _, row := range result.Rows {
		for _, uri := range row {
			ent, err := db.GetEntity(uri)
			if err != nil {
				return "", err
			}
			classList, err := getClass(ent)
			if err != nil {
				return "", err
			}
			preds, objs, err := getEdges(ent)
			if err != nil {
				return "", err
			}
			// add class as node to graph
			for _, class := range classList {
				line := fmt.Sprintf("\"%s\" [fillcolor=\"#4caf50\"];\n", db.abbreviate(class))
				if !strings.Contains(dot, line) {
					dot += line
				}
				for i := 0; i < len(preds); i++ {
					line := fmt.Sprintf("\"%s\" -> \"%s\" [label=\"%s\"];\n", db.abbreviate(class), db.abbreviate(objs[i]), db.abbreviate(preds[i]))
					if !strings.Contains(dot, line) {
						dot += line
					}
				}
			}

		}
	}

	return dot, nil
}

func (db *DB) abbreviate(uri turtle.URI) string {
	for abbv, ns := range db.namespaces {
		if abbv != "" && ns == uri.Namespace {
			return abbv + ":" + uri.Value
		}
	}
	return uri.Value
}

// Searches all of the values in the database; basic wildcard search
func (db *DB) search(q string, n int) ([]string, error) {
	var res []string

	fmt.Println("Displaying", n, "results")
	query := bleve.NewMatchQuery(q)
	search := bleve.NewSearchRequestOptions(query, n, 0, false)
	searchResults, err := db.textidx.Search(search)
	if err != nil {
		fmt.Println(err)
		return res, err
	}
	for _, doc := range searchResults.Hits {
		res = append(res, db.abbreviate(turtle.ParseURI(doc.ID)))
	}
	return res, nil
}

// So, how do queries work?
// We have a list of filters, each of which has a subject, list of predicate things, and
// an object. Any of these might be variables, which we can distinguish by having a "?"
// in front of the value.
//
// First we "clean" these by making sure that they have their full
// namespaces rather than the prefix
func (db *DB) getQueryResults(q *sparql.Query) ([]*ResultRow, queryStats, error) {
	var stats queryStats

	if db.showQueryPlan {
		fmt.Println("-------------- start query plan -------------")
	}
	// start timer
	planStart := time.Now()

	// form dependency graph and build query plan out of it
	dg := db.sortQueryTerms(q)
	qp, err := db.formQueryPlan(dg, q)
	if err != nil {
		return nil, stats, err
	}

	if db.showDependencyGraph {
		dg.dump()
	}

	if db.showQueryPlan {
		for _, op := range qp.operations {
			log.Notice("op", op)
		}
	}
	if db.showQueryPlanLatencies {
		log.Infof("Formed execution plan in %s", time.Since(planStart))
	}
	if db.showQueryPlan {
		fmt.Println("-------------- end query plan -------------")
	}

	runStart := time.Now()
	ctx, err := db.executeQueryPlan(qp)
	defer ctx.Close()
	if err != nil {
		return nil, stats, err
	}
	since := time.Since(runStart)

	runStart = time.Now()
	results := ctx.getResults()
	stats.ExecutionTime = since
	stats.ExpandTime = time.Since(runStart)
	stats.NumResults = len(results)

	return results, stats, err
}

func (db *DB) executeQueryPlan(plan *queryPlan) (*queryContext, error) {
	ctx, err := newQueryContext(plan, db)
	if err != nil {
		return nil, errors.Wrap(err, "Could not get snapshot")
	}

	for _, op := range ctx.operations {
		now := time.Now()
		err := op.run(ctx)
		if db.showOperationLatencies {
			fmt.Println(op, time.Since(now))
		}
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

func (db *DB) sortQueryTerms(q *sparql.Query) *dependencyGraph {
	dg := makeDependencyGraph(q)
	terms := make([]*queryTerm, len(q.Where.Terms))
	for i, f := range q.Where.Terms {
		terms[i] = dg.makeQueryTerm(f)
	}

	// now we order the list such that each term tries to be adjacent
	// to those that it shares a variable with

	// do it twice. First time to put all of the definition terms up front
	// the second time to order by overlap
	sort.Sort(queryTermList(terms))
	sort.Sort(queryTermList(terms))

	dg.terms = terms
	return dg
}

// returns the uint64 hash of the given URI (this is adjusted for uniqueness)
func (db *DB) GetHash(entity turtle.URI) (Key, error) {
	var rethash Key
	if hash, err := db.entityHashCache.Get(entity.Bytes()); err != nil {
		if err == freecache.ErrNotFound {
			val, err := db.entityDB.Get(entity.Bytes(), nil)
			if err != nil {
				return emptyKey, errors.Wrapf(err, "Could not get Entity for %s", entity)
			}
			copy(rethash[:], val)
			if rethash == emptyKey {
				return emptyKey, errors.New("Got bad hash")
			}
			db.entityHashCache.Set(entity.Bytes(), rethash[:], -1) // no expiry
			return rethash, nil
		} else {
			return emptyKey, errors.Wrapf(err, "Could not get Entity for %s", entity)
		}
	} else {
		copy(rethash[:], hash)
	}
	return rethash, nil
}

func (db *DB) MustGetHash(entity turtle.URI) Key {
	val, err := db.GetHash(entity)
	if err != nil {
		log.Error(errors.Wrapf(err, "Could not get hash for %s", entity))
		return emptyKey
	}
	return val
}

func (db *DB) GetURI(hash Key) (turtle.URI, error) {
	db.uriLock.RLock()
	if uri, found := db.uriCache[hash]; found {
		db.uriLock.RUnlock()
		return uri, nil
	}
	db.uriLock.RUnlock()
	db.uriLock.Lock()
	defer db.uriLock.Unlock()
	val, err := db.pkDB.Get(hash[:], nil)
	if err != nil {
		return turtle.URI{}, err
	}
	uri := turtle.ParseURI(string(val))
	db.uriCache[hash] = uri
	return uri, nil
}

func (db *DB) MustGetURI(hash Key) turtle.URI {
	if hash == emptyKey {
		return turtle.URI{}
	}
	uri, err := db.GetURI(hash)
	if err != nil {
		log.Error(errors.Wrapf(err, "Could not get URI for %v", hash))
		return turtle.URI{}
	}
	return uri
}

func (db *DB) MustGetURIStringHash(hash string) turtle.URI {
	var c Key
	copy(c[:], []byte(hash))
	return db.MustGetURI(c)
}

func (db *DB) GetEntity(uri turtle.URI) (*Entity, error) {
	hash, err := db.GetHash(uri)
	if err != nil {
		return nil, err
	}
	return db.GetEntityFromHash(hash)
}

func (db *DB) GetEntityFromHash(hash Key) (*Entity, error) {
	db.eocLock.RLock()
	if ent, found := db.entityObjectCache[hash]; found {
		db.eocLock.RUnlock()
		return ent, nil
	}
	db.eocLock.RUnlock()
	db.eocLock.Lock()
	defer db.eocLock.Unlock()
	bytes, err := db.graphDB.Get(hash[:], nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not get Entity from graph for %s", db.MustGetURI(hash))
	}
	ent := NewEntity()
	_, err = ent.UnmarshalMsg(bytes)
	db.entityObjectCache[hash] = ent
	return ent, err
}

func (db *DB) MustGetEntityFromHash(hash Key) *Entity {
	e, err := db.GetEntityFromHash(hash)
	if err != nil {
		log.Error(errors.Wrap(err, "Could not get entity"))
		return nil
	}
	return e
}

func (db *DB) MustGetEntityIndexFromHash(hash Key) *EntityExtendedIndex {
	e, err := db.GetEntityIndexFromHash(hash)
	if err != nil {
		log.Error(errors.Wrap(err, "Could not get entity index"))
		return nil
	}
	return e
}

func (db *DB) DumpEntity(ent *Entity) {
	fmt.Println("DUMPING", db.MustGetURI(ent.PK))
	for edge, list := range ent.OutEdges {
		fmt.Printf(" OUT: %s \n", db.MustGetURIStringHash(edge).Value)
		for _, l := range list {
			fmt.Printf("     -> %s\n", db.MustGetURI(l).Value)
		}
	}
	for edge, list := range ent.InEdges {
		fmt.Printf(" In: %s \n", db.MustGetURIStringHash(edge).Value)
		for _, l := range list {
			fmt.Printf("     <- %s\n", db.MustGetURI(l).Value)
		}
	}
}

func (db *DB) GetEntityTx(graphtx *leveldb.Transaction, uri turtle.URI) (*Entity, error) {
	var entity = NewEntity()
	hash, err := db.GetHash(uri)
	if err != nil {
		return nil, err
	}
	bytes, err := graphtx.Get(hash[:], nil)
	if err != nil {
		return nil, err
	}
	_, err = entity.UnmarshalMsg(bytes)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (db *DB) GetEntityFromHashTx(graphtx *leveldb.Transaction, hash Key) (*Entity, error) {
	bytes, err := graphtx.Get(hash[:], nil)
	if err != nil {
		return nil, err
	}
	ent := NewEntity()
	_, err = ent.UnmarshalMsg(bytes)
	return ent, err
}

func (db *DB) GetEntityIndexFromHashTx(extendtx *leveldb.Transaction, hash Key) (*EntityExtendedIndex, error) {
	bytes, err := extendtx.Get(hash[:], nil)
	if err != nil {
		return nil, err
	}
	ent := NewEntityExtendedIndex()
	_, err = ent.UnmarshalMsg(bytes)
	return ent, err
}

func (db *DB) GetEntityIndexFromHash(hash Key) (*EntityExtendedIndex, error) {
	db.eicLock.RLock()
	if ent, found := db.entityIndexCache[hash]; found {
		db.eicLock.RUnlock()
		return ent, nil
	}
	db.eicLock.RUnlock()
	db.eicLock.Lock()
	defer db.eicLock.Unlock()
	bytes, err := db.extendedDB.Get(hash[:], nil)
	if err != nil && err != leveldb.ErrNotFound {
		return nil, errors.Wrapf(err, "Could not get EntityIndex from graph for %s", db.MustGetURI(hash))
	} else if err == leveldb.ErrNotFound {
		db.entityIndexCache[hash] = nil
		return nil, nil
	}
	ent := NewEntityExtendedIndex()
	_, err = ent.UnmarshalMsg(bytes)
	db.entityIndexCache[hash] = ent
	return ent, err
}
