package db

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	turtle "github.com/gtfierro/hod/goraptor"
	"github.com/gtfierro/hod/query"

	"github.com/google/btree"
)

func (db *DB) RunQuery(q query.Query) QueryResult {
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

	unionedRows := btree.New(3)
	fullQueryStart := time.Now()

	// if we have terms that are part of a set of OR statements, then we run
	// parallel queries for each fully-elaborated "branch" or the OR statement,
	// and then merge the results together at the end
	if len(orTerms) > 0 {
		var rowLock sync.Mutex
		var wg sync.WaitGroup
		wg.Add(len(orTerms))
		for _, orTerm := range orTerms {
			orTerm := orTerm
			q := q
			go func(orTerm []query.Filter) {
				// augment with the filters
				q.Where.Filters = append(oldFilters, orTerm...)
				results := db.getQueryResults(q)
				rowLock.Lock()
				for _, row := range results {
					unionedRows.ReplaceOrInsert(ResultRow(row))
				}
				rowLock.Unlock()
				wg.Done()
			}(orTerm)
		}
		wg.Wait()
	} else {
		results := db.getQueryResults(q)
		for _, row := range results {
			unionedRows.ReplaceOrInsert(ResultRow(row))
		}
	}
	if db.showQueryLatencies {
		log.Noticef("Full Query took %s", time.Since(fullQueryStart))
	}

	var result QueryResult

	if q.Select.Count {
		// return the count of results
		result.Count = unionedRows.Len()
	} else if q.Select.HasLinks {
		// resolve the links
		max := unionedRows.Max()
		iter := func(i btree.Item) bool {
			row := i.(ResultRow)
			var links = make(LinkResultMap)
			var hasContent = false
			for idx, selectvar := range q.Select.Variables {
				if len(selectvar.Links) == 0 && !selectvar.AllLinks {
					continue
				}
				// check for select all
				if selectvar.AllLinks {
					hash, err := db.GetHash(row[idx])
					if err != nil {
						log.Fatal(err)
					}
					keys, values, err := db.linkDB.getAll(hash)
					if err != nil {
						log.Fatal(err)
					}
					if len(keys) > 0 {
						hasContent = true
						links[row[idx]] = make(map[string]string)
					}
					for i := 0; i < len(keys); i++ {
						links[row[idx]][string(keys[i])] = string(values[i])
					}
				} else {
					for _, _link := range selectvar.Links {
						link := &Link{URI: row[idx], Key: []byte(_link.Name)}
						value, err := db.linkDB.get(link)
						if err != nil {
							log.Fatal(err)
						}
						if len(links[row[idx]]) == 0 {
							links[row[idx]] = make(map[string]string)
						}
						if len(value) > 0 {
							links[row[idx]][string(link.Key)] = string(value)
							hasContent = true
						}
					}
				}
			}
			if hasContent {
				result.Links = append(result.Links, links)
			}
			return row.Less(max)
		}
		unionedRows.Ascend(iter)
	} else {
		// return the rows
		max := unionedRows.Max()
		iter := func(i btree.Item) bool {
			row := i.(ResultRow)
			m := make(ResultMap)
			for idx, vname := range q.Select.Variables {
				m[vname.Var.String()] = row[idx]
			}
			result.Rows = append(result.Rows, m)
			return row.Less(max)
		}
		unionedRows.Ascend(iter)
	}
	return result
}

// TODO: add api call for adding/removing links for entities
func (db *DB) UpdateLinks(updates *LinkUpdates) error {
	tx, err := db.linkDB.startTx()
	if err != nil {
		return err
	}
	for _, linkAdd := range updates.Adding {
		db.linkDB.set(tx, linkAdd)
	}
	for _, linkRm := range updates.Removing {
		db.linkDB.delete(tx, linkRm)
	}
	return tx.Commit()
}

// TODO: add api call for getting links for entities
// for getting links from entities, we probably want to adopt a more generator-based approach
// to actually getting the rows from the database; as we get each row, we get the associated links,
// pipe that out to our accumulator (probably just appending to a list).

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
					predicates = append(predicates, predURI)
					objects = append(objects, class)
				}

			}
		}
		return
	}

	result := db.RunQuery(q)
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

	dot += "}"

	return dot, nil
}
