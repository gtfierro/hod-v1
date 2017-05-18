package db

import (
	"github.com/mitghi/btree"
)

type hashTree struct {
	tree *btree.BTree
}

func newHashTree(size int) *hashTree {
	return &hashTree{
		tree: btree.New(size, ""),
	}
}

func (pt *hashTree) Add(key Key) {
	pt.tree.ReplaceOrInsert(key)
}

func (pt *hashTree) Has(key Key) bool {
	return pt.tree.Has(key)
}

func (pt *hashTree) Len() int {
	return pt.tree.Len()
}

func (pt *hashTree) Max() Key {
	max := pt.tree.Max()
	if max == nil {
		return emptyKey
	}
	return max.(Key)
}

func (pt *hashTree) Min() Key {
	min := pt.tree.Min()
	if min == nil {
		return emptyKey
	}
	return min.(Key)
}

func (pt *hashTree) DeleteMax() Key {
	return pt.tree.DeleteMax().(Key)
}

func (pt *hashTree) Iter(iter func(key Key) bool) {
	pt.tree.Ascend(func(i btree.Item) bool {
		e := i.(Key)
		return iter(e)
	})
}

// TODO: pretty sure this should intersect, but it might be union
// intersects the contents of the pointertree onto the link record
func (pt *hashTree) mergeOntoLinkRecord(rec *linkRecord) {
	max := pt.Max()
	iter := func(k Key) bool {
		newlink := &linkRecord{me: k}
		rec.links = append(rec.links, newlink)
		return k != max
	}
	pt.Iter(iter)
}

func (pt *hashTree) mergeFromTree(t *hashTree) {
	max := t.Max()
	iter := func(e Key) bool {
		pt.Add(e)
		return e != max
	}
	t.Iter(iter)
}
