package db

import (
	"sync"
)

type Row [512]byte

var EMPTYROW = [512]byte{}

var ROWPOOL = sync.Pool{
	New: func() interface{} {
		return &Row{}
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
