package db

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	turtle "github.com/gtfierro/hod/goraptor"
	"github.com/gtfierro/hod/query"

	"github.com/blevesearch/bleve"
	"github.com/coocood/freecache"
	"github.com/mitghi/btree"
	"github.com/pkg/errors"
)

func (db *DB) RunQuery(q query.Query) (QueryResult, error) {
	fullQueryStart := time.Now()

	// "clean" the query by expanding out the prefixes
	// make sure to first do the Filters, then the Or clauses
	for idx, filter := range q.Where.Filters {
		q.Where.Filters[idx] = db.expandFilter(filter)
	}
	for idx, orclause := range q.Where.Ors {
		q.Where.Ors[idx] = db.expandOrClauseFilters(orclause)
	}

	// we flatten the OR clauses to get the array of queries we are going
	// to run and then merge
	orTerms := query.FlattenOrClauseList(q.Where.Ors)
	oldFilters := q.Where.Filters

	// check query hash
	var queryhash []byte
	if db.queryCacheEnabled {
		queryhash = q.Hash(orTerms)
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
	if len(orTerms) > 0 {
		var rowLock sync.Mutex
		var wg sync.WaitGroup
		var queryErr error
		wg.Add(len(orTerms))
		for _, orTerm := range orTerms {
			tmpQuery := q.Copy()
			// augment with the filters
			tmpQuery.Where.Filters = make([]query.Filter, len(oldFilters)+len(orTerm))
			copy(tmpQuery.Where.Filters, oldFilters)
			copy(tmpQuery.Where.Filters[len(oldFilters):], orTerm)

			//	go func(q query.Query) {
			results, err := db.getQueryResults(tmpQuery)
			rowLock.Lock()
			if err != nil {
				queryErr = err
			} else {
				for _, row := range results {
					unionedRows.ReplaceOrInsert(row)
				}
			}
			rowLock.Unlock()
			wg.Done()
			//}(tmpQuery)
		}
		wg.Wait()
		if queryErr != nil {
			return QueryResult{}, queryErr
		}
	} else {
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
	result.selectVars = q.Select.Variables
	result.Elapsed = time.Since(fullQueryStart)

	if q.Select.Count {
		// return the count of results
		result.Count = unionedRows.Len()
	} else {
		// return the rows
		i := unionedRows.DeleteMax()
		for i != nil {
			row := i.(*ResultRow)
			m := make(ResultMap)
			for idx, vname := range q.Select.Variables {
				m[vname.Var.String()] = row.row[idx]
			}
			result.Rows = append(result.Rows, m)
			finishResultRow(row)
			i = unionedRows.DeleteMax()
		}
		result.Count = len(result.Rows)
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

// takes a query and returns a DOT representation to visualize
// the construction of the query
func (db *DB) QueryToDOT(querystring io.Reader) (string, error) {
	q, err := query.Parse(querystring)
	if err != nil {
		return "", err
	}

	dot := ""
	dot += "digraph G {\n"
	dot += "ratio=\"auto\"\n"
	dot += "rankdir=\"LR\"\n"
	dot += "size=\"7.5,10\"\n"

	if len(q.Where.Ors) > 0 {
		orTerms := query.FlattenOrClauseList(q.Where.Ors)
		oldFilters := q.Where.Filters
		for _, orTerm := range orTerms {
			filters := append(oldFilters, orTerm...)
			for _, filter := range filters {
				var parts []string
				for _, p := range filter.Path {
					parts = append(parts, fmt.Sprintf("%s%s", p.Predicate, p.Pattern))
				}
				line := fmt.Sprintf("\"%s\" -> \"%s\" [label=\"%s\"];\n", filter.Subject, filter.Object, strings.Join(parts, "/"))
				if !strings.Contains(dot, line) {
					dot += line
				}

			}
		}
	} else {
		for _, filter := range q.Where.Filters {
			var parts []string
			for _, p := range filter.Path {
				parts = append(parts, fmt.Sprintf("%s%s", p.Predicate, p.Pattern))
			}
			line := fmt.Sprintf("\"%s\" -> \"%s\" [label=\"%s\"];\n", filter.Subject, filter.Object, strings.Join(parts, "/"))
			if !strings.Contains(dot, line) {
				dot += line
			}
		}
	}
	for _, sv := range q.Select.Variables {
		dot += fmt.Sprintf("\"%s\" [fillcolor=#e57373]\n", sv.Var)
	}
	dot += "}"
	return dot, nil
}

// executes a query and returns a DOT string of the classes involved
func (db *DB) QueryToClassDOT(querystring io.Reader) (string, error) {
	q, err := query.Parse(querystring)
	if err != nil {
		return "", err
	}
	// create DOT template string
	dot := ""
	dot += "digraph G {\n"

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

	result, err := db.RunQuery(q)
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
				line := fmt.Sprintf("\"%s\" [fillcolor=\"#4caf50\"];\n", db.Abbreviate(class))
				if !strings.Contains(dot, line) {
					dot += line
				}
				for i := 0; i < len(preds); i++ {
					line := fmt.Sprintf("\"%s\" -> \"%s\" [label=\"%s\"];\n", db.Abbreviate(class), db.Abbreviate(objs[i]), db.Abbreviate(preds[i]))
					if !strings.Contains(dot, line) {
						dot += line
					}
				}
			}

		}
	}

	dot += "}"

	return dot, nil
}

func (db *DB) Abbreviate(uri turtle.URI) string {
	for abbv, ns := range db.namespaces {
		if abbv != "" && ns == uri.Namespace {
			return abbv + ":" + uri.Value
		}
	}
	return uri.Value
}

// Searches all of the values in the database; basic wildcard search
func (db *DB) Search(q string, n int) ([]string, error) {
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
		res = append(res, db.Abbreviate(turtle.ParseURI(doc.ID)))
	}
	return res, nil
}
