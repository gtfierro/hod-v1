package db

import (
	"fmt"
	"github.com/gtfierro/hod/query"
	"github.com/mitghi/btree"
)

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
		fmt.Println(db.MustGetURI(i.(*Entity).PK))
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

func rowIsFull(row []Key) bool {
	for _, entry := range row {
		if entry == emptyKey {
			return false
		}
	}
	return true
}

func reversePath(path []query.PathPattern) {
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
}
