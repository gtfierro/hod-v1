package db

import (
	"fmt"
)

var trees = newBtreePool(BTREE_DEGREE)
var emptyHashTree = newKeyTree(BTREE_DEGREE)

type queryContext2 struct {
	// maps variable name to a position in a row
	variablePosition map[string]int
	selectVars       []string

	// variable definitions
	definitions map[string]*keyTree

	rows *RowTree

	rel *Relation

	// names of joined variables
	joined []string

	db *DB
}

func newQueryContext2(plan *queryPlan, db *DB) *queryContext2 {
	variablePosition := make(map[string]int)
	definitions := make(map[string]*keyTree)
	for idx, variable := range plan.query.Variables {
		variablePosition[variable] = idx
	}
	log.Debug("2>", variablePosition)
	return &queryContext2{
		variablePosition: variablePosition,
		definitions:      definitions,
		selectVars:       plan.selectVars,
		rows:             NewRowTree(),
		rel:              NewRelation(plan.query.Variables),
		db:               db,
	}
}

func (ctx *queryContext2) cardinalityUnique(varname string) int {
	if tree, found := ctx.definitions[varname]; found {
		return tree.Len()
	} else {
		return 0
	}
}

func (ctx *queryContext2) hasJoined(varname string) bool {
	for _, vname := range ctx.joined {
		if vname == varname {
			return true
		}
	}
	return false
}

func (ctx *queryContext2) validValue(varname string, value Key) bool {
	if tree, found := ctx.definitions[varname]; found {
		return tree.Has(value)
	}
	return true
}

func (ctx *queryContext2) markJoined(varname string) {
	for _, vname := range ctx.joined {
		if vname == varname {
			return
		}
	}
	ctx.joined = append(ctx.joined, varname)
}

func (ctx *queryContext2) addRowWithValue(varname string, value Key) {
	position := ctx.variablePosition[varname]
	row := NewRow()
	row.addValue(position, value)
	ctx.rows.Add(row)
}

func (ctx *queryContext2) getValuesForVariable(varname string) *keyTree {
	tree, found := ctx.definitions[varname]
	if found {
		return tree
	}
	return emptyHashTree
}

func (ctx *queryContext2) defineVariable(varname string, values *keyTree, intersect bool) {

	tree := ctx.definitions[varname]
	// TODO: intersect? merge?
	if tree == nil || tree.Len() == 0 {
		ctx.definitions[varname] = values
		//// for each value, we create the set of rows
		values.Iter(func(key Key) {
			ctx.addRowWithValue(varname, key)
		})
	}
}

func (ctx *queryContext2) addDefinition(varname string, value Key) {
	tree := ctx.definitions[varname]
	if tree == nil || tree.Len() == 0 {
		ctx.definitions[varname] = newKeyTree(BTREE_DEGREE)
		ctx.definitions[varname].Add(value)
	} else {
		tree.Add(value)
	}
	//ctx.addRowWithValue(varname, value)
}

// remove the values from 'values' that aren't in the values we already have
func (ctx *queryContext2) restrictToResolved(varname string, values *keyTree) {
	tree, found := ctx.definitions[varname]
	if !found {
		return // do not change the tree
	}
	// remove bad values
	cursor := values.Cursor()
	_item := cursor.Seek(values.Min())
	if _item == nil {
		return
	}
	item := _item.(Key)
	for {
		if !tree.Has(item) {
			values.Delete(item)
		}
		_next := cursor.Next()
		if _next == nil {
			break
		}
		next := _next.(Key)
		item = next
		cursor.Seek(item)
	}
}

// for rows where sourceVarname is sourceValue, add a new version of the row with each of the values in targetValues populated in the position for targetVarname

// want to be able to get all of the rows where sourceVar is populated with sourceValue.
// want to be able to copy those rows and add new values to them (and remove the old rows)
// want ot be able to remove rows that have a given value in a given position

func (ctx *queryContext2) populateValues(sourceVarname string, sourceValue Key, targetVarname string, addValues *keyTree) {
	addPos := ctx.variablePosition[targetVarname]
	sourceIdx := ctx.variablePosition[sourceVarname]

	if !ctx.hasJoined(sourceVarname) {
		row := NewRow()
		row.addValue(sourceIdx, sourceValue)
		ctx.rows.Add(row)
	}

	var toAdd []*Row
	var toRemove []*Row
	ctx.rows.iterRowsWithValue(sourceIdx, sourceValue, func(r *Row) {
		addValues.Iter(func(addValue Key) {
			newRow := r.copy()
			newRow.addValue(addPos, addValue)
			ctx.addDefinition(targetVarname, addValue)
			toAdd = append(toAdd, newRow)
		})
		toRemove = append(toRemove, r)
	})

	for _, row := range toRemove {
		ctx.rows.tree.Delete(row)
		row.release()
	}
	for _, row := range toAdd {
		ctx.rows.Add(row)
	}
}

func (ctx *queryContext2) joinValuePairs(targetVarname1, targetVarname2 string, targetValues [][]Key) {
	targetIdx1 := ctx.variablePosition[targetVarname1]
	targetIdx2 := ctx.variablePosition[targetVarname2]

	var joinKeyPos int
	var otherKeyPos int
	var joinPairIdx int
	var otherPairIdx int
	if ctx.hasJoined(targetVarname1) {
		joinKeyPos = targetIdx1
		otherKeyPos = targetIdx2
		joinPairIdx = 0
		otherPairIdx = 1
	} else if ctx.hasJoined(targetVarname2) {
		joinKeyPos = targetIdx2
		otherKeyPos = targetIdx1
		joinPairIdx = 1
		otherPairIdx = 0
	} else {
		ctx.addValuePairs(targetVarname1, targetVarname2, targetValues)
		return
	}

	var toAdd []*Row
	var toRemove []*Row
	for _, pair := range targetValues {
		ctx.rows.iterRowsWithValue(joinKeyPos, pair[joinPairIdx], func(r *Row) {
			newRow := r.copy()
			newRow.addValue(otherKeyPos, pair[otherPairIdx])
			toAdd = append(toAdd, newRow)
			toRemove = append(toRemove, r)
		})
	}
	for _, row := range toRemove {
		ctx.rows.tree.Delete(row)
		row.release()
	}
	for _, row := range toAdd {
		ctx.rows.Add(row)
	}
}

func (ctx *queryContext2) addValuePairs(sourceVarname, targetVarname string, pairs [][]Key) {
	targetIdx := ctx.variablePosition[targetVarname]
	sourceIdx := ctx.variablePosition[sourceVarname]
	for _, pair := range pairs {
		row := NewRow()
		//log.Debug("adding", sourceVarname, ctx.db.MustGetURI(pair[0]), targetVarname, ctx.db.MustGetURI(pair[1]))
		row.addValue(sourceIdx, pair[0])
		row.addValue(targetIdx, pair[1])
		ctx.addDefinition(sourceVarname, pair[0])
		ctx.addDefinition(targetVarname, pair[1])
		ctx.rows.Add(row)
	}
}

func (ctx *queryContext2) dumpRows() {
	ctx.rows.iterAll(func(row *Row) {
		ctx.dumpRow(row)
	})
}

func (ctx *queryContext2) dumpRow(row *Row) {
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

func (ctx *queryContext2) getResults() (results []*ResultRow) {

	//ctx.definitions[varname].Iter(func(key Key) {
	//	ctx.addRowWithValue(varname, key)
	//})
	//	ctx.rel.dumpRows(ctx.db)

	ctx.rel.rows.iterAll(func(row *Row) {
		resultrow := getResultRow(len(ctx.selectVars))
		for idx, varname := range ctx.selectVars {
			val := row.valueAt(ctx.variablePosition[varname])
			if val == emptyKey {
				return
			}
			resultrow.row[idx] = ctx.db.MustGetURI(row.valueAt(ctx.variablePosition[varname]))
		}
		results = append(results, resultrow)
	})
	return
}
