package db

import (
	"fmt"
	"github.com/google/btree"
)

// merges all the keys from 'src' into 'dst'
func mergeTrees(dest, src *btree.BTree) {
	max := src.Max()
	iter := func(i btree.Item) bool {
		dest.ReplaceOrInsert(i)
		return i != max
	}
	src.Ascend(iter)
}

// takes a btree of [4]byte hashes, and turns those into
// a tree of ResultEntity
func hashTreeToEntityTree(src *btree.BTree) *btree.BTree {
	newTree := btree.New(3)
	max := src.Max()
	iter := func(i btree.Item) bool {
		ve := &ResultEntity{
			PK:   i.(Item),
			Next: make(map[string]*btree.BTree),
		}
		newTree.ReplaceOrInsert(ve)
		return i != max
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
	max := a.Max()
	iter := func(i btree.Item) bool {
		if b.Has(i) {
			res.ReplaceOrInsert(i)
		}
		return i != max
	}
	a.Ascend(iter)
	return res
}

func dumpHashTree(tree *btree.BTree, db *DB, limit int) {
	max := tree.Max()
	iter := func(i btree.Item) bool {
		if limit == 0 {
			return false // stop iteration
		} else if limit > 0 {
			limit -= 1 //
		}
		fmt.Println(db.MustGetURI(i.(Item)))
		return i != max
	}
	tree.Ascend(iter)
}

func dumpEntityTree(tree *btree.BTree, db *DB, limit int) {
	max := tree.Max()
	iter := func(i btree.Item) bool {
		if limit == 0 {
			return false // stop iteration
		} else if limit > 0 {
			limit -= 1 //
		}
		fmt.Println(db.MustGetURI(i.(*ResultEntity).PK))
		return i != max
	}
	tree.Ascend(iter)
}

func compareResultMapList(rml1, rml2 []ResultMap) bool {
	var (
		found bool
	)

	if len(rml1) != len(rml2) {
		return false
	}

	for _, val1 := range rml1 {
		found = false
		for _, val2 := range rml2 {
			if compareResultMap(val1, val2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func compareResultMap(rm1, rm2 ResultMap) bool {
	if len(rm1) != len(rm2) {
		return false
	}
	for k, v := range rm1 {
		if v2, found := rm2[k]; !found {
			return false
		} else if v2 != v {
			return false
		}
	}
	return true
}
