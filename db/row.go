package db

import (
	"bytes"
	"sync"

	"github.com/gtfierro/hod/storage"
)

type relationRow struct {
	content []byte
}

var rowPool = sync.Pool{
	New: func() interface{} {
		return &relationRow{
			content: make([]byte, 64),
		}
	},
}

func newRelationRow() *relationRow {
	row := rowPool.Get().(*relationRow)
	row.content = row.content[:0]
	return row
}

func (row *relationRow) release() {
	rowPool.Put(row)
}

func (row *relationRow) copy() *relationRow {
	gr := rowPool.Get().(*relationRow)
	if len(gr.content) < len(row.content) {
		gr.content = make([]byte, len(row.content))
	}
	copy(gr.content[:], row.content[:])
	return gr
}

func (row *relationRow) addValue(pos int, value storage.HashKey) {
	if len(row.content) < pos*8+8 {
		nrow := make([]byte, pos*8+8)
		copy(nrow, row.content)
		row.content = nrow
	}
	copy(row.content[pos*8:], value[:])
}

func (row relationRow) valueAt(pos int) storage.HashKey {
	var k storage.HashKey
	if pos*8+8 > len(row.content) {
		return k
	}
	copy(k[:], row.content[pos*8:pos*8+8])
	return k
}

func (row relationRow) equals(other relationRow) bool {
	return bytes.Equal(row.content[:], other.content[:])
}
