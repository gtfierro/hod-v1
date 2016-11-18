package db

import (
	"github.com/google/btree"
)

// merges all the keys from 'src' into 'dst'
func mergeTrees(dest, src *btree.BTree) {
	iter := func(i btree.Item) bool {
		dest.ReplaceOrInsert(i)
		return i != src.Max()
	}
	src.Ascend(iter)
}

// takes a btree of [4]byte hashes, and turns those into
// a tree of VariableEntity
func hashTreeToEntityTree(src *btree.BTree) *btree.BTree {
	newTree := btree.New(3)
	iter := func(i btree.Item) bool {
		ve := &VariableEntity{
			PK:    i.(Item),
			Links: make(map[string]*btree.BTree),
		}
		newTree.ReplaceOrInsert(ve)
		return i != src.Max()
	}
	src.Ascend(iter)
	return newTree
}
