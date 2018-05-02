package db

import (
	"github.com/gtfierro/btree"
	"testing"
)

var intersectTreesBtree = func(a, b *btree.BTree) *btree.BTree {
	if a.Len() < b.Len() {
		a, b = b, a
	}
	res := btree.New(BTREE_DEGREE, "")
	if a.Max().Less(b.Min(), "") {
		return res
	}
	iter := func(i btree.Item) bool {
		if b.Has(i) {
			res.ReplaceOrInsert(i)
		}
		return i != a.Max()
	}
	a.Ascend(iter)
	return res
}

func BenchmarkInsertTree100(b *testing.B) {
	trees := make([]*btree.BTree, b.N)
	for i := 0; i < b.N; i++ {
		trees[i] = btree.New(BTREE_DEGREE, "")
	}
	for i := 0; i < b.N; i++ {
		bits := [4]byte{byte(i), byte(i + 1), byte(i + 2), byte(i + 3)}
		trees[i].ReplaceOrInsert(Key(bits))
	}
}

func BenchmarkIntersectTreesBtree50(b *testing.B) {
	A := btree.New(BTREE_DEGREE, "")
	B := btree.New(BTREE_DEGREE, "")
	for i := 0; i < 100; i++ {
		bitsa := [4]byte{byte(i), byte(i + 1), byte(i + 2), byte(i + 3)}
		bitsb := [4]byte{byte(i + 50), byte(i + 50 + 1), byte(i + 50 + 2), byte(i + 50 + 3)}
		A.ReplaceOrInsert(Key(bitsa))
		B.ReplaceOrInsert(Key(bitsb))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		intersectTreesBtree(A, B)
	}
}

func BenchmarkIntersectTreesBtree01(b *testing.B) {
	A := btree.New(BTREE_DEGREE, "")
	B := btree.New(BTREE_DEGREE, "")
	for i := 0; i < 100; i++ {
		bitsa := [4]byte{byte(i), byte(i + 1), byte(i + 2), byte(i + 3)}
		bitsb := [4]byte{byte(i + 99), byte(i + 99 + 1), byte(i + 99 + 2), byte(i + 99 + 3)}
		A.ReplaceOrInsert(Key(bitsa))
		B.ReplaceOrInsert(Key(bitsb))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		intersectTreesBtree(A, B)
	}
}

func BenchmarkIntersectTreesBtreeAll(b *testing.B) {
	A := btree.New(BTREE_DEGREE, "")
	B := btree.New(BTREE_DEGREE, "")
	for i := 0; i < 100; i++ {
		bitsa := [4]byte{byte(i), byte(i + 1), byte(i + 2), byte(i + 3)}
		bitsb := [4]byte{byte(i), byte(i + 1), byte(i + 2), byte(i + 3)}
		A.ReplaceOrInsert(Key(bitsa))
		B.ReplaceOrInsert(Key(bitsb))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		intersectTreesBtree(A, B)
	}
}

func BenchmarkIntersectTreesBtreeNone(b *testing.B) {
	A := btree.New(BTREE_DEGREE, "")
	B := btree.New(BTREE_DEGREE, "")
	for i := 0; i < 100; i++ {
		bitsa := [4]byte{byte(i), byte(i + 1), byte(i + 2), byte(i + 3)}
		bitsb := [4]byte{byte(i + 100), byte(i + 101), byte(i + 102), byte(i + 103)}
		A.ReplaceOrInsert(Key(bitsa))
		B.ReplaceOrInsert(Key(bitsb))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		intersectTreesBtree(A, B)
	}
}

func BenchmarkBTreeHas3(b *testing.B) {
	t := btree.New(BTREE_DEGREE, "")
	e := Key([4]byte{1, 2, 3, 4})
	for i := 0; i < b.N; i++ {
		t.Has(e)
	}
}

func BenchmarkBTreeInsertDuplicate3(b *testing.B) {
	t := btree.New(BTREE_DEGREE, "")
	e := Key([4]byte{1, 2, 3, 4})
	for i := 0; i < b.N; i++ {
		t.ReplaceOrInsert(e)
	}
}

func BenchmarkBTreeInsertDuplicateWithHas3(b *testing.B) {
	t := btree.New(BTREE_DEGREE, "")
	e := Key([4]byte{1, 2, 3, 4})
	for i := 0; i < b.N; i++ {
		if !t.Has(e) {
			t.ReplaceOrInsert(e)
		}
	}
}

func BenchmarkBTreeHas2(b *testing.B) {
	t := btree.New(BTREE_DEGREE, "")
	e := Key([4]byte{1, 2, 3, 4})
	for i := 0; i < b.N; i++ {
		t.Has(e)
	}
}

func BenchmarkBTreeInsertDuplicate2(b *testing.B) {
	t := btree.New(BTREE_DEGREE, "")
	e := Key([4]byte{1, 2, 3, 4})
	for i := 0; i < b.N; i++ {
		t.ReplaceOrInsert(e)
	}
}

func BenchmarkBTreeInsertDuplicateWithHas2(b *testing.B) {
	t := btree.New(BTREE_DEGREE, "")
	e := Key([4]byte{1, 2, 3, 4})
	for i := 0; i < b.N; i++ {
		if !t.Has(e) {
			t.ReplaceOrInsert(e)
		}
	}
}
