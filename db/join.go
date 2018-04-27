package db

import (
	"sync"
)

type Row struct {
	content []byte
}

var ROWPOOL = sync.Pool{
	New: func() interface{} {
		return &Row{}
	},
}

func NewRow() *Row {
	return ROWPOOL.Get().(*Row)
}

func NewRowWithNum(withnum int) *Row {
	row := ROWPOOL.Get().(*Row)
	row.content = make([]byte, withnum*4+4)
	return row
}

func (row *Row) release() {
	row.content = row.content[:0]
	//for i := range row.content {
	//	row.content[i] = 0
	//}
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
	if len(row.content) < pos*4+4 {
		nrow := make([]byte, pos*4+4)
		copy(nrow, row.content)
		row.content = nrow
	}
	copy(row.content[pos*4:], value[:])
}

func (row Row) valueAt(pos int) Key {
	var k Key
	copy(k[:], row.content[pos*4:pos*4+4])
	return k
}
