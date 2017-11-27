package db

import (
	"fmt"
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

	db *DB

	// embedded query plan
	*queryPlan
}

func newQueryContext(plan *queryPlan, db *DB) *queryContext {
	variablePosition := make(map[string]int)
	definitions := make(map[string]*keyTree)
	for idx, variable := range plan.query.Variables {
		variablePosition[variable] = idx
	}
	log.Debug("2>", variablePosition)
	return &queryContext{
		variablePosition: variablePosition,
		definitions:      definitions,
		selectVars:       plan.selectVars,
		rel:              NewRelation(plan.query.Variables),
		db:               db,
		queryPlan:        plan,
	}
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

func (ctx *queryContext) addDefinition(varname string, value Key) {
	tree := ctx.definitions[varname]
	if tree == nil || tree.Len() == 0 {
		ctx.definitions[varname] = newKeyTree()
		ctx.definitions[varname].Add(value)
	} else {
		tree.Add(value)
	}
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
			s += varName + "=" + ctx.db.MustGetURI(val).String() + ", "
		}
	}
	s += "]"
	fmt.Println(s)
}

func (ctx *queryContext) getResults() (results []*ResultRow) {

	for _, row := range ctx.rel.rows {
		resultrow := getResultRow(len(ctx.selectVars))
		for idx, varname := range ctx.selectVars {
			val := row.valueAt(ctx.variablePosition[varname])
			if val == emptyKey {
				return
			}
			resultrow.row[idx] = ctx.db.MustGetURI(row.valueAt(ctx.variablePosition[varname]))
		}
		results = append(results, resultrow)
		row.release()
	}
	return
}
