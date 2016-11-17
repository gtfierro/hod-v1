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
