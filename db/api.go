package db

import (
	"sync"
	"time"

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
				if len(selectvar.Links) == 0 {
					continue
				}
				// TODO: check for select all
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
