package db

import (
	"fmt"

	"github.com/gtfierro/hod/proto"
	"github.com/gtfierro/hod/storage"
	//logrus "github.com/sirupsen/logrus"
)

var emptyHashTree = newKeymap()

type queryContext struct {
	// maps variable name to a position in a row
	variablePosition map[string]int
	selectVars       []string
	rows             []*resultRow

	// variable definitions
	definitions map[string]*keymap

	rel *relation

	// names of joined variables
	joined []string

	tx *transaction
	// embedded query plan
	*queryPlan
}

func newQueryContext(plan *queryPlan, tx *transaction) (*queryContext, error) {
	variablePosition := make(map[string]int)
	definitions := make(map[string]*keymap)
	for idx, variable := range plan.query.Variables {
		variablePosition[variable] = idx
	}

	return &queryContext{
		variablePosition: variablePosition,
		definitions:      definitions,
		selectVars:       plan.selectVars,
		rel:              newRelation(plan.query.Variables),
		tx:               tx,
		queryPlan:        plan,
	}, nil
}

func (ctx *queryContext) cardinalityUnique(varname string) int {
	if tree, found := ctx.definitions[varname]; found {
		return tree.Len()
	}
	return 0
}

func (ctx *queryContext) hasJoined(varname string) bool {
	for _, vname := range ctx.joined {
		if vname == varname {
			return true
		}
	}
	return false
}

func (ctx *queryContext) validValue(varname string, value storage.HashKey) bool {
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

func (ctx *queryContext) getValuesForVariable(varname string) *keymap {
	tree, found := ctx.definitions[varname]
	if found {
		return tree
	}
	return emptyHashTree
}

func (ctx *queryContext) defineVariable(varname string, values *keymap) {
	tree := ctx.definitions[varname]
	if tree == nil || tree.Len() == 0 {
		ctx.definitions[varname] = values
	}
}

func (ctx *queryContext) defined(varname string) bool {
	_, found := ctx.definitions[varname]
	return found
}

func (ctx *queryContext) unionDefinitions(varname string, values *keymap) {
	ctx.restrictToResolved(varname, values)
	ctx.definitions[varname] = values
}

// remove the values from 'values' that aren't in the values we already have
func (ctx *queryContext) restrictToResolved(varname string, values *keymap) {
	tree, found := ctx.definitions[varname]
	if !found {
		return // do not change the tree
	}
	// remove bad values

	var toDelete []storage.HashKey

	values.Iter(func(k storage.HashKey) {
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
		ctx.dumpRow("", row)
	}
}

func (ctx *queryContext) dumpRow(prefix string, row *relationRow) {
	s := prefix + " ["
	for varName, pos := range ctx.variablePosition {
		val := row.valueAt(pos)
		if val != storage.EmptyKey {
			uri, err := ctx.tx.getURI(val)
			if err != nil {
				panic(err)
			}
			s += varName + "=" + uri.String() + ", "
		}
	}
	s += "]"
	fmt.Println(s)
}

func (ctx *queryContext) release() {
	ctx.rel.done()
	for _, row := range ctx.rows {
		finishResultRow(row)
	}
}

func (ctx *queryContext) getResults() (results []*proto.Row) {
	results = make([]*proto.Row, len(ctx.rel.rows))
	var jtest = make(map[uint32]struct{})
	numRows := 0
	var positions = make([]int, len(ctx.selectVars))
	for idx, varname := range ctx.selectVars {
		positions[idx] = ctx.variablePosition[varname]
	}
rowIter:
	for _, row := range ctx.rel.rows {
		hash := hashRowWithPos(row, positions)
		if _, found := jtest[hash]; found {
			continue
		}
		jtest[hash] = struct{}{}

		var resultrow = new(proto.Row)

		for _, varname := range ctx.selectVars {
			val := row.valueAt(ctx.variablePosition[varname])
			if val == storage.EmptyKey {
				continue rowIter
			}
			uri, err := ctx.tx.getURI(row.valueAt(ctx.variablePosition[varname]))
			if err != nil {
				panic(err)
			}
			resultrow.Uris = append(resultrow.Uris, &proto.URI{Namespace: uri.Namespace, Value: uri.Value})
		}
		results[numRows] = resultrow
		numRows++
	}
	return results[:numRows]
}
