package db

import (
	"bytes"
	"sync"

	"github.com/mitghi/btree"
)

type Row [32]byte

var EMPTYROW = [32]byte{}

var ROWPOOL = sync.Pool{
	New: func() interface{} {
		return &Row{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	},
}

func NewRow() *Row {
	return ROWPOOL.Get().(*Row)
}

func (row *Row) release() {
	copy(row[:], EMPTYROW[:])
	ROWPOOL.Put(row)
}

func (row *Row) copy() *Row {
	gr := ROWPOOL.Get().(*Row)
	copy(gr[:], row[:])
	return gr
}

func (row *Row) addValue(pos int, value Key) {
	copy(row[pos*4:], value[:])
}

func (row Row) valueAt(pos int) Key {
	var k Key
	copy(k[:], row[pos*4:pos*4+4])
	return k
}

func (row *Row) swap(pos1, pos2 int) *Row {
	if pos1 == pos2 {
		return row
	}
	newRow := row.copy()
	copy(newRow[pos1*4:], row[pos2*4:pos2*4+4])
	copy(newRow[pos2*4:], row[pos1*4:pos1*4+4])
	return newRow
}

func (row *Row) Less(_than btree.Item, ctx interface{}) bool {

	than := _than.(*Row)
	return bytes.Compare(row[:], than[:]) < 0
}

// now to fix the join structures!

// want to be able to get all of the rows where [position] is populated with [value].
// want to be able to copy those rows and add new values to them (and remove the old rows)
// want to be able to remove rows that have a given value in a given position

type Joiner struct {
}

/*
Need to remember when performing these operations that we are joining onto the structure.

Operations:
- defineVariable (variable name, values):
    - keep track of these values as the valid values for that variable
- restrictValues (variable name, values):
    - return the intersection of these incoming values and the values we already have
    - if we don't have any defined values (i.e. variable has'nt been seen before), then
      return all values
    - if we *have* seen the variable before and there are no values, then we return nothing
- crossProduct (key varname, key value, target varname, target value(s)):
    - for all rows where varname==value, we add a new row for each target value (for target varname)
    - do we also need to go the other way? What about rows where targetvar==targetvalue ?

    example:
    ?floor rdf:type brick:Floor .    <--- these first 3 rows get 'defined'
    ?room rdf:type brick:Room .
    ?zone rdf:type brick:HVAC_Zone .

    ?room bf:isPartOf+ ?floor .      <--- adds rows with [?room ?floor]
    ?room bf:isPartOf+ ?zone .

    The last row here is the tricky one. We can't just naively evaluate it like we did
    the one before, because then we end up with rows that are [?room ?floor] and others that are
    [?zone ?room]. We need to keep track of "joined variables" so that we can use them

- addPairs (key varname, key values, target varname, target values)
*/

//func main() {
//	var rows = []*Row{
//		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
//		{1, 1, 1, 1, 0, 0, 0, 0, 4, 4, 4, 4, 0, 0, 0, 0, 4, 4, 4, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
//		{1, 1, 1, 1, 2, 2, 2, 2, 4, 4, 4, 4, 0, 0, 0, 0, 4, 4, 4, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
//		{2, 2, 2, 2, 2, 2, 2, 2, 4, 4, 4, 4, 0, 0, 0, 0, 4, 4, 4, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
//		{2, 2, 2, 2, 2, 2, 2, 2, 5, 5, 5, 5, 6, 6, 6, 6, 4, 4, 4, 4, 0, 0, 0, 0, 0, 0, 0, 0, 8, 8, 8, 8},
//		{2, 2, 2, 2, 2, 2, 2, 2, 5, 5, 5, 5, 0, 0, 0, 0, 4, 4, 4, 4, 0, 0, 0, 0, 0, 0, 0, 0, 9, 9, 9, 9},
//	}
//
//	rowtree := NewRowTree()
//	for _, row := range rows {
//		rowtree.Add(row)
//	}
//
//	fmt.Println("Test iter")
//	rowtree.iterRowsWithValue(0, Key{2, 2, 2, 2}, func(r *Row) {
//		fmt.Printf("  %+v\n", r)
//	})
//	fmt.Println("--")
//	rowtree.iterRowsWithValue(3, Key{6, 6, 6, 6}, func(r *Row) {
//		fmt.Printf("  %+v\n", r)
//	})
//
//	fmt.Println("Test Remove")
//	rowtree.deleteRowsWithValue(3, Key{6, 6, 6, 6})
//	rowtree.iterRowsWithValue(0, Key{2, 2, 2, 2}, func(r *Row) {
//		fmt.Printf("  %+v\n", r)
//	})
//
//	fmt.Println("Test Augment")
//	rowtree.augmentByValue(0, Key{0, 0, 0, 0}, 1, Key{1, 1, 1, 1})
//	rowtree.iterRowsWithValue(0, Key{0, 0, 0, 0}, func(r *Row) {
//		fmt.Printf("  %+v\n", r)
//	})
//
//	fmt.Println("Test Augment2")
//	rowtree.augmentByValues(1, Key{1, 1, 1, 1}, 6, []Key{{4, 5, 6, 7}, {1, 2, 3, 4}, {2, 4, 6, 8}})
//	rowtree.iterRowsWithValue(1, Key{1, 1, 1, 1}, func(r *Row) {
//		fmt.Printf("  %+v\n", r)
//	})
//	fmt.Println("all")
//	rowtree.iterAll(func(r *Row) {
//		fmt.Printf(" %+v\n", r)
//	})
//}

var _freelist = btree.NewFreeList(10000)

// need a structure that allows us to resort rows based on a given value

type RowTree struct {
	tree *btree.BTree
}

func NewRowTree() *RowTree {
	return &RowTree{
		tree: btree.NewWithFreeList(BTREE_DEGREE, _freelist, struct{}{}),
	}
}

func (tree *RowTree) Add(row *Row) {
	tree.tree.ReplaceOrInsert(row)
}

func (tree *RowTree) copyAndSortByValue(pos int) *RowTree {
	sorted := NewRowTree()
	tree.iterAll(func(row *Row) {
		sorted.Add(row.swap(0, pos))
	})
	return sorted
}

func (tree *RowTree) iterRowsWithValue(pos int, value Key, f func(r *Row)) {
	sorted := tree.copyAndSortByValue(pos)
	var rangeStart, rangeEnd Row
	copy(rangeStart[:], iterLower[:])
	copy(rangeEnd[:], iterUpper[:])
	copy(rangeStart[:], value[:])
	copy(rangeEnd[:], value[:])

	sorted.tree.AscendRange(&rangeStart, &rangeEnd, func(_row btree.Item) bool {
		row := _row.(*Row)
		f(row.swap(0, pos))
		return _row.Less(&rangeEnd, struct{}{})
	})
}

func (tree *RowTree) deleteRowsWithValue(pos int, value Key) {
	var toDelete []*Row
	tree.iterRowsWithValue(pos, value, func(r *Row) {
		toDelete = append(toDelete, r)
	})
	for _, row := range toDelete {
		tree.tree.Delete(row)
		row.release()
	}
}

func (tree *RowTree) augmentByValue(pos int, value Key, addPos int, addValue Key) {
	var toAdd []*Row
	var toRemove []*Row
	tree.iterRowsWithValue(pos, value, func(r *Row) {
		newRow := r.copy()
		newRow.addValue(addPos, addValue)
		toAdd = append(toAdd, newRow)
		toRemove = append(toRemove, r)
	})
	for _, row := range toRemove {
		tree.tree.Delete(row)
		row.release()
	}
	for _, row := range toAdd {
		tree.Add(row)
	}
}

func (tree *RowTree) augmentByValues(pos int, value Key, addPos int, addValues *keyTree) {
	var toAdd []*Row
	var toRemove []*Row
	tree.iterRowsWithValue(pos, value, func(r *Row) {
		addValues.Iter(func(addValue Key) {
			newRow := r.copy()
			newRow.addValue(addPos, addValue)
			toAdd = append(toAdd, newRow)
		})
		toRemove = append(toRemove, r)
	})

	for _, row := range toRemove {
		tree.tree.Delete(row)
		row.release()
	}
	for _, row := range toAdd {
		tree.Add(row)
	}
}

func (tree *RowTree) augmentByValuePairs(pos int, value Key, addPos1, addPos2 int, addValues [][]Key) {
	var toAdd []*Row
	var toRemove []*Row
	tree.iterRowsWithValue(pos, value, func(r *Row) {
		for _, pair := range addValues {
			newRow := r.copy()
			newRow.addValue(addPos1, pair[0])
			newRow.addValue(addPos2, pair[1])
			toAdd = append(toAdd, newRow)
		}
		toRemove = append(toRemove, r)
	})

	for _, row := range toRemove {
		tree.tree.Delete(row)
		row.release()
	}
	for _, row := range toAdd {
		tree.Add(row)
	}
}
func (tree *RowTree) iterAll(f func(row *Row)) {
	max := tree.tree.Max()
	tree.tree.Ascend(func(_row btree.Item) bool {
		f(_row.(*Row))
		return _row != max
	})
}

func (tree *RowTree) releaseAll() {
	max := tree.tree.DeleteMax()
	for max != nil {
		max.(*Row).release()
		max = tree.tree.DeleteMax()
	}
}

var iterLower = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
var iterUpper = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
