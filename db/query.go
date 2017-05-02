package db

import (
	"fmt"
	"sort"
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
	dg := db.sortQueryTerms(q)
	qp := db.formQueryPlan(dg, q)

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
	ctx := db.executeQueryPlan(qp)
	if db.showQueryLatencies {
		log.Infof("Ran query in %s", time.Since(runStart))
	}

	//ctx.dumpTraverseOrder()

	runStart = time.Now()
	results := ctx.expandTuples()
	if db.showQueryLatencies {
		log.Infof("Expanded tuples in %s", time.Since(runStart))
		log.Infof("Has %d results", len(results))
	}
	return results
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

func (db *DB) sortQueryTerms(q query.Query) *dependencyGraph {
	dg := makeDependencyGraph(q)
	terms := make([]*queryTerm, len(q.Where.Filters))
	for i, f := range q.Where.Filters {
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
