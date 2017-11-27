package db

import (
	"bytes"
	"sync"

	"github.com/mitghi/btree"
)

type Row [32]byte

var EMPTYROW = [32]byte{}

var ROWPOOL = sync.Pool{
	New: func() interface{} {
		return &Row{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	},
}

func NewRow() *Row {
	return ROWPOOL.Get().(*Row)
}

func (row *Row) release() {
	copy(row[:], EMPTYROW[:])
	ROWPOOL.Put(row)
}

func (row *Row) copy() *Row {
	gr := ROWPOOL.Get().(*Row)
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

func (row *Row) swap(pos1, pos2 int) *Row {
	if pos1 == pos2 {
		return row
	}
	newRow := row.copy()
	copy(newRow[pos1*4:], row[pos2*4:pos2*4+4])
	copy(newRow[pos2*4:], row[pos1*4:pos1*4+4])
	return newRow
}

func (row *Row) Less(_than btree.Item, ctx interface{}) bool {

	than := _than.(*Row)
	return bytes.Compare(row[:], than[:]) < 0
}
