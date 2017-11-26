package db

import (
	"bytes"
	"fmt"
	//	"github.com/zhangxinngang/murmur"
)

type Relation struct {
	rows *RowTree
	// map variable name to position in row
	vars map[string]int
	keys []string
}

func NewRelation(vars []string) *Relation {
	rel := &Relation{
		keys: vars,
		vars: make(map[string]int),
		rows: NewRowTree(),
	}
	for idx, varname := range vars {
		rel.vars[varname] = idx
	}
	return rel
}

func (rel *Relation) add1Value(key1 string, values *keyTree) {
	key1pos, found := rel.vars[key1]
	if !found {
		key1pos = len(rel.vars) + 1
		rel.vars[key1] = key1pos
	}

	values.Iter(func(value Key) {
		row := NewRow()
		row.addValue(key1pos, value)
		rel.rows.Add(row)
	})
}

func (rel *Relation) add2Values(key1, key2 string, values [][]Key) {
	key1pos, found := rel.vars[key1]
	if !found {
		key1pos = len(rel.vars) + 1
		rel.vars[key1] = key1pos
	}
	key2pos, found := rel.vars[key2]
	if !found {
		key2pos = len(rel.vars) + 1
		rel.vars[key2] = key2pos
	}

	for _, valuepair := range values {
		row := NewRow()
		row.addValue(key1pos, valuepair[0])
		row.addValue(key2pos, valuepair[1])
		rel.rows.Add(row)
	}
}

func (rel *Relation) add3Values(key1, key2, key3 string, values [][]Key) {
	key1pos, found := rel.vars[key1]
	if !found {
		key1pos = len(rel.vars) + 1
		rel.vars[key1] = key1pos
	}
	key2pos, found := rel.vars[key2]
	if !found {
		key2pos = len(rel.vars) + 1
		rel.vars[key2] = key2pos
	}
	key3pos, found := rel.vars[key3]
	if !found {
		key3pos = len(rel.vars) + 1
		rel.vars[key3] = key3pos
	}

	for _, valuetriple := range values {
		row := NewRow()
		row.addValue(key1pos, valuetriple[0])
		row.addValue(key2pos, valuetriple[1])
		row.addValue(key3pos, valuetriple[2])
		rel.rows.Add(row)
	}
}

// this is a left inner join onto 'rel' on the keys in 'on'
func (rel *Relation) join(other *Relation, on []string, ctx *queryContext) {

	// get the variable positions for the join variables for
	// each of the relations (these may be different)
	var relJoinKeyPos []int
	var otherJoinKeyPos []int
	for _, varname := range on {
		relJoinKeyPos = append(relJoinKeyPos, rel.vars[varname])
		otherJoinKeyPos = append(otherJoinKeyPos, other.vars[varname])
	}

	var toAdd []*Row
	var toRemove []*Row
	rel.rows.iterAll(func(relRow *Row) {
		merged := true
		other.rows.iterAll(func(otherRow *Row) {
			matches := true
			for joinIdx := range on {
				if !matches {
					merged = false
					return // skip this row
				}
				leftVal := relRow.valueAt(relJoinKeyPos[joinIdx])
				rightVal := otherRow.valueAt(otherJoinKeyPos[joinIdx])
				matches = matches && bytes.Compare(leftVal[:], rightVal[:]) == 0
			}

			// here, we know the two rows match. Merge in the otherRow values
			// into a *copy* of relRow
			newRow := relRow.copy()
			for otherVarname, otherIdx := range other.vars {
				newRow.addValue(rel.vars[otherVarname], otherRow.valueAt(otherIdx))
			}
			toAdd = append(toAdd, newRow)

		})
		// mark relRow for deletion if we had merged it above
		if merged {
			toRemove = append(toRemove, relRow)
		}
	})

	for _, row := range toRemove {
		rel.rows.tree.Delete(row)
		row.release()
	}
	for _, row := range toAdd {
		rel.rows.Add(row)
	}
	other.rows.releaseAll()
}

func (rel *Relation) dumpRow(db *DB, row *Row) {
	s := "["
	for varName, pos := range rel.vars {
		val := row.valueAt(pos)
		if val != emptyKey {
			s += varName + "=" + db.MustGetURI(val).String() + ", "
		}
	}
	s += "]"
	fmt.Println(s)
}

func (rel *Relation) dumpRows(db *DB) {
	rel.rows.iterAll(func(r *Row) {
		rel.dumpRow(db, r)
	})
}
