package db

import (
	"fmt"
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
// a tree of ResultEntity
func hashTreeToEntityTree(src *btree.BTree) *btree.BTree {
	newTree := btree.New(3)
	iter := func(i btree.Item) bool {
		ve := &ResultEntity{
			PK:   i.(Item),
			Next: make(map[string]*btree.BTree),
		}
		newTree.ReplaceOrInsert(ve)
		return i != src.Max()
	}
	src.Ascend(iter)
	return newTree
}

// takes the intersection of the two trees and returns it
func intersectTrees(a, b *btree.BTree) *btree.BTree {
	if a.Len() < b.Len() {
		a, b = b, a
	}
	res := btree.New(3)
	iter := func(i btree.Item) bool {
		if b.Has(i) {
			res.ReplaceOrInsert(i)
		}
		return i != a.Max()
	}
	a.Ascend(iter)
	return res
}

func dumpHashTree(tree *btree.BTree, db *DB, limit int) {
	iter := func(i btree.Item) bool {
		if limit == 0 {
			return false // stop iteration
		} else if limit > 0 {
			limit -= 1 //
		}
		fmt.Println(db.MustGetURI(i.(Item)))
		return i != tree.Max()
	}
	tree.Ascend(iter)
}

func dumpEntityTree(tree *btree.BTree, db *DB, limit int) {
	iter := func(i btree.Item) bool {
		if limit == 0 {
			return false // stop iteration
		} else if limit > 0 {
			limit -= 1 //
		}
		fmt.Println(db.MustGetURI(i.(*ResultEntity).PK))
		return i != tree.Max()
	}
	tree.Ascend(iter)
}
