package main

import (
	"github.com/mitghi/btree"
)

const BTREE_DEGREE = 4

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

func (tree *RowTree) augmentByValues(pos int, value Key, addPos int, addValues []Key) {
	var toAdd []*Row
	var toRemove []*Row
	tree.augmentByValue(pos, value, addPos, Key{2, 2, 2, 2})
	tree.iterRowsWithValue(pos, value, func(r *Row) {
		for _, addValue := range addValues {
			newRow := r.copy()
			newRow.addValue(addPos, addValue)
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

var iterLower = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
var iterUpper = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
