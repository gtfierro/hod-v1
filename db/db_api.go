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
	"github.com/kr/pretty"
	"github.com/pkg/errors"
)

func prettyprint(v interface{}) {
	fmt.Printf("%# v", pretty.Formatter(v))
}

func (db *DB) runQueryString(q string) ([]*ResultRow, queryStats, error) {
	var (
		rows  []*ResultRow
		stats queryStats
	)
	if q, err := query.Parse(q); err != nil {
		e := errors.Wrap(err, "Could not parse hod query")
		log.Error(e)
		return rows, stats, e
	} else if rows, stats, err = db.runQuery(q); err != nil {
		e := errors.Wrap(err, "Could not complete hod query")
		log.Error(e)
		return rows, stats, e
	} else {
		return rows, stats, nil
	}
}

func (db *DB) runQuery(q *sparql.Query) ([]*ResultRow, queryStats, error) {
	var result []*ResultRow

	whereStart := time.Now()

	// expand out the prefixes
	q.IterTriples(func(triple sparql.Triple) sparql.Triple {
		triple.Subject = db.expand(triple.Subject)
		triple.Object = db.expand(triple.Object)
		for idx2, pred := range triple.Predicates {
			triple.Predicates[idx2].Predicate = db.expand(pred.Predicate)
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
	var stats queryStats
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
			return result, stats, queryErr
		}
	} else {
		results, _stats, err := db.getQueryResults(q)
		stats = _stats
		if err != nil {
			return result, stats, err
		}
		result = append(result, results...)
	}
	stats.WhereTime = time.Since(whereStart)
	//logrus.WithFields(logrus.Fields{
	//	"Name":    db.name,
	//	"Where":   q.Select.Vars,
	//	"Execute": stats.WhereTime,
	//	"Expand":  stats.ExpandTime,
	//	"Results": stats.NumResults,
	//	"Total":   time.Since(whereStart),
	//}).Info("Query")
	return result, stats, nil
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
	//// create DOT template string
	dot := ""

	// get rdf:type predicate hash as a string
	typeURI := turtle.ParseURI("rdf:type")
	typeURI.Namespace = db.namespaces[typeURI.Namespace]
	snap, err := db.snapshot()
	if err != nil {
		return "", err
	}
	typeKey, err := snap.getHash(typeURI)
	if err != nil {
		return "", err
	}
	log.Debug(typeURI, typeKey)
	typeKeyString := string(typeKey[:])

	getClass := func(ent *Entity) (classes []turtle.URI, err error) {
		_classes := ent.OutEdges[typeKeyString]
		//		for name := range ent.OutEdges {
		//			var k Key
		//			copy(k[:], []byte(name))
		//			uri, err := snap.getURI(k)
		//			if err != nil {
		//				panic(err)
		//			}
		//			log.Debug(uri, k, typeKey)
		//		}
		for _, class := range _classes {
			classes = append(classes, mustGetURI(snap, class))
		}
		return
	}

	getEdges := func(ent *Entity) (predicates, objects []turtle.URI, reterr error) {
		var predKey Key
		for predKeyString, objectList := range ent.OutEdges {
			predKey.FromSlice([]byte(predKeyString))
			predURI, err := snap.getURI(predKey)
			if err != nil {
				reterr = err
				return
			}
			for _, objectKey := range objectList {
				objectEnt, err := snap.getEntityByHash(objectKey)
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
				if predURI.Value == "uuid" {
					predicates = append(predicates, predURI)
					objects = append(objects, turtle.URI{Namespace: "bf", Value: "uuid"})
				}

			}
		}
		return
	}

	result, _, err := db.runQuery(q)
	if err != nil {
		return "", err
	}
	for _, row := range result {
		for _, uri := range row.row {
			ent, err := snap.getEntityByURI(uri)
			if err != nil {
				log.Debug(err)
				return "", err
			}
			classList, err := getClass(ent)
			if err != nil {
				log.Debug(err)
				return "", err
			}
			preds, objs, err := getEdges(ent)
			if err != nil {
				log.Debug(err)
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
	return ""
}

func (db *DB) expand(uri turtle.URI) turtle.URI {
	if !strings.HasPrefix(uri.Value, "?") {
		if full, found := db.namespaces[uri.Namespace]; found {
			uri.Namespace = full
		}
	}
	return uri
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
		abb := db.abbreviate(turtle.ParseURI(doc.ID))
		if abb != "" {
			res = append(res, abb)
		}
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
	//dg := db.sortQueryTerms(q)
	dg := makeDependencyGraph(q)
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
	defer ctx.t.under.done()
	if err != nil {
		return nil, stats, err
	}
	since := time.Since(runStart)

	runStart = time.Now()
	results := ctx.getResults()
	stats.WhereTime = since
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
