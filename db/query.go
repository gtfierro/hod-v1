package db

import (
	"fmt"
	"time"

	turtle "github.com/gtfierro/hod/goraptor"
	"github.com/gtfierro/hod/query"
)

// make a set of structs that capture what these queries want to do

// So, how do queries work?
// We have a list of filters, each of which has a subject, list of predicate things, and
// an object. Any of these might be variables, which we can distinguish by having a "?"
// in front of the value.
//
// First we "clean" these by making sure that they have their full
// namespaces rather than the prefix

func (db *DB) getQueryResults(q query.Query) [][]turtle.URI {
	if db.showQueryPlan {
		fmt.Println("-------------- start query plan -------------")
	}
	// start timer
	planStart := time.Now()

	// form dependency graph and build query plan out of it
	dg := db.formDependencyGraph(q)
	qp := db.formQueryPlan(dg, q)

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
	ctx := db.executeQueryPlan(qp)
	if db.showQueryLatencies {
		log.Infof("Ran query in %s", time.Since(runStart))
	}

	runStart = time.Now()
	ctx.dumpTraverseOrder()
	results := ctx.expandTuples()
	if db.showQueryLatencies {
		log.Infof("Expanded tuples in %s", time.Since(runStart))
	}
	return results
}

// We need an execution plan for the list of filters contained in a query. How do we do this?
func (db *DB) formDependencyGraph(q query.Query) *dependencyGraph {
	dg := makeDependencyGraph(q)
	terms := make([]*queryTerm, len(q.Where.Filters))
	for i, f := range q.Where.Filters {
		terms[i] = dg.makeQueryTerm(f)
	}

	numUnresolved := func(qt *queryTerm) int {
		num := 0
		for _, v := range qt.variables {
			if !dg.variables[v] {
				num++
			}
		}
		return num
	}

	originalLength := len(terms)
	for len(terms) > 0 {
		// first find all the terms with 0 or 1 unresolved variable terms
		var added = []*queryTerm{}
		for _, term := range terms {
			if numUnresolved(term) < 2 {
				if len(dg.roots) == 0 {
					dg.addRootTerm(term)
				} else {
					dg.addChild(term)
				}
				added = append(added, term)
			}
		}
		// remove the terms that we added to the root set
		terms = filterTermList(terms, added)
		added = []*queryTerm{}
		for _, term := range terms {
			if dg.addChild(term) {
				added = append(added, term)
			}
		}
		terms = filterTermList(terms, added)
		if len(terms) == originalLength {
			// we don't have any root elements. Need to consider 2-variable terms
			added = []*queryTerm{}
			for _, term := range terms {
				if numUnresolved(term) == 2 {
					dg.addRootTerm(term)
					added = append(added, term)
					break
				}
			}
			terms = filterTermList(terms, added)
		}
	}
	if db.showDependencyGraph {
		dg.dump()
	}
	return dg
}

func (db *DB) executeQueryPlan(plan *queryPlan) *queryContext {
	ctx := newQueryContext(plan, db)

	for _, op := range ctx.operations {
		now := time.Now()
		err := op.run(ctx)
		if db.showQueryPlanLatencies {
			fmt.Println(op, time.Since(now))
		}
		if err != nil {
			log.Fatal(err)
		}
	}
	return ctx
}
