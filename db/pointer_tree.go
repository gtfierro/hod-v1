package db

import (
	"github.com/mitghi/btree"
)

const BTREE_DEGREE = 4

var fl = btree.NewFreeList(16384)

type linkRecord struct {
	me    Key
	links []*linkRecord
}

type pointerTree struct {
	tree *btree.BTree
}

func newPointerTree(size int) *pointerTree {
	return &pointerTree{
		tree: btree.NewWithFreeList(size, fl, ""),
	}
}

func (pt *pointerTree) Add(ent *Entity) {
	pt.tree.ReplaceOrInsert(ent)
}

func (pt *pointerTree) Has(ent *Entity) bool {
	return pt.tree.Has(ent)
}

func (pt *pointerTree) Len() int {
	return pt.tree.Len()
}

func (pt *pointerTree) Max() *Entity {
	max := pt.tree.Max()
	if max == nil {
		return nil
	}
	return max.(*Entity)
}

func (pt *pointerTree) Min() *Entity {
	min := pt.tree.Min()
	if min == nil {
		return nil
	}
	return min.(*Entity)
}

func (pt *pointerTree) DeleteMax() *Entity {
	return pt.tree.DeleteMax().(*Entity)
}

func (pt *pointerTree) Iter(iter func(ent *Entity) bool) {
	pt.tree.Ascend(func(i btree.Item) bool {
		e := i.(*Entity)
		return iter(e)
	})
}

// TODO: pretty sure this should intersect, but it might be union
// intersects the contents of the pointertree onto the link record
func (pt *pointerTree) mergeOntoLinkRecord(rec *linkRecord) {
	max := pt.Max()
	iter := func(ent *Entity) bool {
		newlink := &linkRecord{me: ent.PK}
		rec.links = append(rec.links, newlink)
		return ent != max
	}
	pt.Iter(iter)
}

func (pt *pointerTree) mergeFromTree(t *pointerTree) {
	max := t.Max()
	iter := func(e *Entity) bool {
		pt.Add(e)
		return e != max
	}
	t.Iter(iter)
}
