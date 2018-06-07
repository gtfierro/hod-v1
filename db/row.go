package db

import (
	"bytes"
	"sync"
)

type Row struct {
	content []byte
}

var ROWPOOL = sync.Pool{
	New: func() interface{} {
		return &Row{
			content: make([]byte, 64),
		}
	},
}

func NewRow() *Row {
	return ROWPOOL.Get().(*Row)
}

func NewRowWithNum(withnum int) *Row {
	row := ROWPOOL.Get().(*Row)
	row.content = make([]byte, withnum*8+8)
	return row
}

func (row *Row) release() {
	row.content = row.content[:0]
	ROWPOOL.Put(row)
}

func (row *Row) copy() *Row {
	gr := ROWPOOL.Get().(*Row)
	if len(gr.content) < len(row.content) {
		gr.content = make([]byte, len(row.content))
	}
	copy(gr.content[:], row.content[:])
	return gr
}

func (row *Row) addValue(pos int, value Key) {
	if len(row.content) < pos*8+8 {
		nrow := make([]byte, pos*8+8)
		copy(nrow, row.content)
		row.content = nrow
	}
	copy(row.content[pos*8:], value[:])
}

func (row Row) valueAt(pos int) Key {
	var k Key
	if pos*8+8 > len(row.content) {
		return k
	}
	copy(k[:], row.content[pos*8:pos*8+8])
	return k
}

func (row Row) equals(other Row) bool {
	return bytes.Equal(row.content[:], other.content[:])
}
