package main

import (
	"fmt"

	"github.com/mitghi/btree"
)

const BTREE_DEGREE = 4

// need a structure that allows us to resort rows based on a given value

type RowTree struct {
	tree *btree.BTree
}

func NewRowTree() *RowTree {
	return &RowTree{
		tree: btree.New(BTREE_DEGREE, struct{}{}),
	}
}

func (tree *RowTree) Add(row Row) {
	tree.tree.ReplaceOrInsert(row)
}

func (tree *RowTree) copyAndSortByValue(pos int) *RowTree {
	sorted := NewRowTree()
	tree.iterAll(func(row Row) {
		sorted.Add(row.swap(0, pos))
	})
	return sorted
}

func (tree *RowTree) iterRowsWithValue(pos int, value Key, f func(r Row)) {
	sorted := tree.copyAndSortByValue(pos)
	var rangeStart, rangeEnd Row
	copy(rangeStart[:], iterLower[:])
	copy(rangeEnd[:], iterUpper[:])
	copy(rangeStart[:], value[:])
	copy(rangeEnd[:], value[:])

	sorted.tree.AscendRange(rangeStart, rangeEnd, func(_row btree.Item) bool {
		row := _row.(Row)
		f(row.swap(0, pos))
		return _row.Less(rangeEnd, struct{}{})
	})
}

func (tree *RowTree) deleteRowsWithValue(pos int, value Key) {
	var toDelete []Row
	tree.iterRowsWithValue(pos, value, func(r Row) {
		toDelete = append(toDelete, r)
	})
	for _, row := range toDelete {
		tree.tree.Delete(row)
	}
	tree.iterRowsWithValue(pos, value, func(r Row) {
		fmt.Println("        ", r)
		for _, row := range toDelete {
			fmt.Println("deleting", row)
		}
	})
}

func (tree *RowTree) iterAll(f func(row Row)) {
	max := tree.tree.Max()
	tree.tree.Ascend(func(_row btree.Item) bool {
		f(_row.(Row))
		return _row != max
	})
}

var iterLower = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
var iterUpper = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
