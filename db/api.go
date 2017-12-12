package db

import (
	"fmt"
	"strings"
	"sync"
	"time"

	query "github.com/gtfierro/hod/lang"
	sparql "github.com/gtfierro/hod/lang/ast"
	"github.com/gtfierro/hod/turtle"

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
				results, err := db.getQueryResults(&tmpQuery)
				rowLock.Lock()
				if err != nil {
					queryErr = err
				} else {
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
		results, err := db.getQueryResults(q)
		if err != nil {
			return QueryResult{}, err
		}
		for _, row := range results {
			unionedRows.ReplaceOrInsert(row)
		}
	}
	if db.showQueryLatencies {
		log.Noticef("Full Query took %s", time.Since(fullQueryStart))
	}

	var result = newQueryResult()
	result.selectVars = q.Select.Vars
	result.Elapsed = time.Since(fullQueryStart)

	// TODO: count!
	// return the rows
	log.Debug(unionedRows.Len())

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
				results, err := db.getQueryResults(&tmpQuery)
				rowLock.Lock()
				if err != nil {
					queryErr = err
				} else {
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
		results, err := db.getQueryResults(q)
		if err != nil {
			return result, err
		}
		result = append(result, results...)
	}
	if db.showQueryLatencies {
		log.Noticef("Full Query took %s", time.Since(fullQueryStart))
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
