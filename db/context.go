package db

import (
	"fmt"

	"github.com/pkg/errors"
)

var trees = newBtreePool(BTREE_DEGREE)
var emptyHashTree = newKeyTree()

type queryContext struct {
	// maps variable name to a position in a row
	variablePosition map[string]int
	selectVars       []string

	// variable definitions
	definitions map[string]*keyTree

	rel *Relation

	// names of joined variables
	joined []string

	graph traversable
	db    *DB
	// embedded query plan
	*queryPlan
}

func newQueryContext(plan *queryPlan, db *DB) (*queryContext, error) {
	variablePosition := make(map[string]int)
	definitions := make(map[string]*keyTree)
	for idx, variable := range plan.query.Variables {
		variablePosition[variable] = idx
	}

	snap, err := db.snapshot()
	if err != nil {
		return nil, errors.Wrap(err, "Could not get snapshot")
	}

	return &queryContext{
		variablePosition: variablePosition,
		definitions:      definitions,
		selectVars:       plan.selectVars,
		rel:              NewRelation(plan.query.Variables),
		db:               db,
		queryPlan:        plan,
		graph:            snap,
	}, nil
}

func (ctx *queryContext) cardinalityUnique(varname string) int {
	if tree, found := ctx.definitions[varname]; found {
		return tree.Len()
	} else {
		return 0
	}
}

func (ctx *queryContext) hasJoined(varname string) bool {
	for _, vname := range ctx.joined {
		if vname == varname {
			return true
		}
	}
	return false
}

func (ctx *queryContext) validValue(varname string, value Key) bool {
	if tree, found := ctx.definitions[varname]; found {
		return tree.Has(value)
	}
	return true
}

func (ctx *queryContext) markJoined(varname string) {
	for _, vname := range ctx.joined {
		if vname == varname {
			return
		}
	}
	ctx.joined = append(ctx.joined, varname)
}

func (ctx *queryContext) getValuesForVariable(varname string) *keyTree {
	tree, found := ctx.definitions[varname]
	if found {
		return tree
	}
	return emptyHashTree
}

func (ctx *queryContext) defineVariable(varname string, values *keyTree) {
	tree := ctx.definitions[varname]
	if tree == nil || tree.Len() == 0 {
		ctx.definitions[varname] = values
	}
}

func (ctx *queryContext) defined(varname string) bool {
	_, found := ctx.definitions[varname]
	return found
}

func (ctx *queryContext) unionDefinitions(varname string, values *keyTree) {
	ctx.restrictToResolved(varname, values)
	ctx.definitions[varname] = values
}

// remove the values from 'values' that aren't in the values we already have
func (ctx *queryContext) restrictToResolved(varname string, values *keyTree) {
	tree, found := ctx.definitions[varname]
	if !found {
		return // do not change the tree
	}
	// remove bad values

	var toDelete []Key

	values.Iter(func(k Key) {
		if !tree.Has(k) {
			toDelete = append(toDelete, k)
		}
	})
	for _, k := range toDelete {
		values.Delete(k)
	}
}

func (ctx *queryContext) dumpRows() {
	for _, row := range ctx.rel.rows {
		ctx.dumpRow(row)
	}
}

func (ctx *queryContext) dumpRow(row *Row) {
	s := "["
	for varName, pos := range ctx.variablePosition {
		val := row.valueAt(pos)
		if val != emptyKey {
			uri, err := ctx.graph.getURI(val)
			if err != nil {
				panic(err)
			}
			s += varName + "=" + uri.String() + ", "
		}
	}
	s += "]"
	fmt.Println(s)
}

func (ctx *queryContext) getResults() (results []*ResultRow) {

	results = make([]*ResultRow, len(ctx.rel.rows))
	idx := 0
rowIter:
	for _, row := range ctx.rel.rows {
		resultrow := getResultRow(len(ctx.selectVars))
		for idx, varname := range ctx.selectVars {
			val := row.valueAt(ctx.variablePosition[varname])
			if val == emptyKey {
				continue rowIter
			}
			var err error
			resultrow.row[idx], err = ctx.graph.getURI(row.valueAt(ctx.variablePosition[varname]))
			if err != nil {
				panic(err)
			}
		}
		results[idx] = resultrow
		idx++
		row.release()
	}
	results = results[:idx]
	return
}
