package db

import (
	"github.com/google/btree"
)

type linkRecord struct {
	me    Key
	links []*linkRecord
}

type pointerTree struct {
	tree *btree.BTree
}

func newPointerTree(size int) *pointerTree {
	return &pointerTree{
		tree: btree.New(size),
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
	return pt.tree.Max().(*Entity)
}

func (pt *pointerTree) Min() *Entity {
	return pt.tree.Min().(*Entity)
}

func (pt *pointerTree) Delete(ent *Entity) *Entity {
	return pt.tree.Delete(ent).(*Entity)
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
