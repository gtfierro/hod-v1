package main

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/mitghi/btree"
)

type Key [4]byte

type Row [32]byte

var EMPTYROW = [32]byte{}

var ROWPOOL = sync.Pool{
	New: func() interface{} {
		return Row{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	},
}

func NewRow() Row {
	return ROWPOOL.Get().(Row)
}

func (row *Row) release() {
	copy(row[:], EMPTYROW[:])
	ROWPOOL.Put(row)
}

func (row Row) copy() Row {
	gr := ROWPOOL.Get().(Row)
	copy(gr[:], row[:])
	return gr
}

func (row *Row) addValue(pos int, value Key) {
	copy(row[pos*4:], value[:])
}

func (row Row) valueAt(pos int) Key {
	var k Key
	copy(k[:], row[pos*4:pos*4+4])
	return k
}

func (row Row) swap(pos1, pos2 int) Row {
	if pos1 == pos2 {
		return row
	}
	newRow := row.copy()
	copy(newRow[pos1*4:], row[pos2*4:pos2*4+4])
	copy(newRow[pos2*4:], row[pos1*4:pos1*4+4])
	return newRow
}

func (row Row) Less(_than btree.Item, ctx interface{}) bool {

	than := _than.(Row)
	return bytes.Compare(row[:], than[:]) <= 0
}

// now to fix the join structures!

// want to be able to get all of the rows where [position] is populated with [value].
// want to be able to copy those rows and add new values to them (and remove the old rows)
// want to be able to remove rows that have a given value in a given position

type Joiner struct {
}

func main() {
	var rows = []Row{
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{1, 1, 1, 1, 0, 0, 0, 0, 4, 4, 4, 4, 0, 0, 0, 0, 4, 4, 4, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{1, 1, 1, 1, 2, 2, 2, 2, 4, 4, 4, 4, 0, 0, 0, 0, 4, 4, 4, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{2, 2, 2, 2, 2, 2, 2, 2, 4, 4, 4, 4, 0, 0, 0, 0, 4, 4, 4, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{2, 2, 2, 2, 2, 2, 2, 2, 5, 5, 5, 5, 6, 6, 6, 6, 4, 4, 4, 4, 0, 0, 0, 0, 0, 0, 0, 0, 8, 8, 8, 8},
		{2, 2, 2, 2, 2, 2, 2, 2, 5, 5, 5, 5, 0, 0, 0, 0, 4, 4, 4, 4, 0, 0, 0, 0, 0, 0, 0, 0, 9, 9, 9, 9},
	}

	rowtree := NewRowTree()
	for _, row := range rows {
		rowtree.Add(row)
	}

	fmt.Println("Test iter")
	rowtree.iterRowsWithValue(0, Key{2, 2, 2, 2}, func(r Row) {
		fmt.Printf("  %+v\n", r)
	})
	fmt.Println("--")
	rowtree.iterRowsWithValue(3, Key{6, 6, 6, 6}, func(r Row) {
		fmt.Printf("  %+v\n", r)
	})

	fmt.Println("Test Remove")
	rowtree.deleteRowsWithValue(3, Key{6, 6, 6, 6})
	rowtree.iterRowsWithValue(0, Key{2, 2, 2, 2}, func(r Row) {
		fmt.Printf("  %+v\n", r)
	})
}
