package db

import (
	"bytes"
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
			PK:   i.(Key),
			Next: make(map[string]*btree.BTree),
		}
		newTree.ReplaceOrInsert(ve)
		return i != max
	}
	src.Ascend(iter)
	return newTree
}

// merges all the keys from 'src' into 'dst'
func mergePointerTrees(dest, src *pointerTree) {
	max := src.Max()
	iter := func(e *Entity) bool {
		dest.Add(e)
		return e != max
	}
	src.Iter(iter)
}

// takes a btree of [4]byte hashes, and turns those into
// a tree of ResultEntity
func hashTreeToPointerTree(db *DB, src *btree.BTree) *pointerTree {
	newTree := newPointerTree(3)
	max := src.Max()
	iter := func(i btree.Item) bool {
		if i == nil {
			return i != max
		}
		ve := db.MustGetEntityFromHash(i.(Key))
		newTree.Add(ve)
		return i != max
	}
	src.Ascend(iter)
	return newTree
}

// takes the intersection of the two trees and returns it
func intersectTrees(a, b *btree.BTree) *btree.BTree {
	res := btree.New(3)
	// early skip
	if a.Len() == 0 || b.Len() == 0 || a.Max().Less(b.Min()) || b.Max().Less(a.Min()) {
		return res
	}
	if a.Len() < b.Len() {
		a, b = b, a
	}
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

// takes the intersection of the two pointertrees and returns it
func intersectPointerTrees(a, b *pointerTree) *pointerTree {
	res := newPointerTree(3)
	// early skip
	if a.Len() == 0 || b.Len() == 0 || a.Max().Less(b.Min()) || b.Max().Less(a.Min()) {
		return res
	}
	if a.Len() < b.Len() {
		a, b = b, a
	}
	max := a.Max()
	iter := func(e *Entity) bool {
		if b.Has(e) {
			res.Add(e)
		}
		return e != max
	}
	a.Iter(iter)
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
		fmt.Println(db.MustGetURI(i.(Key)))
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

func compareLinkUpdates(up1, up2 *LinkUpdates) bool {
	var found bool
	if len(up1.Adding) != len(up2.Adding) {
		return false
	}
	if len(up1.Removing) != len(up2.Removing) {
		return false
	}
	for _, val1 := range up1.Adding {
		found = false
		for _, val2 := range up2.Adding {
			if compareLink(val1, val2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	for _, val1 := range up1.Removing {
		found = false
		for _, val2 := range up2.Removing {
			if compareLink(val1, val2) {
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

func compareLink(l1, l2 *Link) bool {
	return l1.URI == l2.URI &&
		l1.entity == l2.entity &&
		bytes.Equal(l1.Key, l2.Key) &&
		bytes.Equal(l1.Value, l2.Value)
}
