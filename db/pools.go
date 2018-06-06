package db

import (
	"sync"

	"github.com/gtfierro/btree"
)

const BTREE_DEGREE = 8

var (
	traversedBTreePool *btreePool
)

func init() {
	traversedBTreePool = newBtreePool(BTREE_DEGREE)
}

type btreePool struct {
	pool sync.Pool
	size int
}

func newBtreePool(size int) *btreePool {
	return &btreePool{
		size: size,
		pool: sync.Pool{
			New: func() interface{} {
				return btree.New(size, "")
			},
		},
	}
}

func (btp *btreePool) Get() *btree.BTree {
	return btp.pool.Get().(*btree.BTree)
}

func (btp *btreePool) Put(b *btree.BTree) {
	// clear the tree
	for b.Max() != nil {
		b.DeleteMax()
	}
	btp.pool.Put(b)
}
