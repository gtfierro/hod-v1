package db

import (
	"sync"

	turtle "github.com/gtfierro/hod/goraptor"
)

const emptyVar = ""

var emptyEntries = make([]Key, 16)

var _ROWPOOL = sync.Pool{
	New: func() interface{} {
		return &row{
			entries:   make([]Key, 16),
			numFilled: 0,
		}
	},
}

type row struct {
	// entries in this row
	entries   []Key
	vars      []string
	lastvar   string
	lastidx   int
	numFilled int
}

func newrow(vars []string) *row {
	r := _ROWPOOL.Get().(*row)
	if len(vars) > len(r.entries) {
		r.entries = make([]Key, len(vars))
	}
	r.vars = vars
	return r
}

func finishrow(r *row) {
	r.numFilled = 0
	r.lastvar = emptyVar
	r.lastidx = 0
	copy(r.entries, emptyEntries)
	_ROWPOOL.Put(r)
}

func (r *row) isFull() bool {
	for _, entry := range r.entries[:len(r.vars)] {
		if entry == emptyKey {
			return false
		}
	}
	return true
}

func (r *row) addVar(name string, index int, value Key) {
	r.lastvar = name
	r.lastidx = index
	r.entries[index] = value
	r.numFilled += 1
}

func (r *row) expand(ctx *queryContext) *ResultRow {
	newrow := getResultRow(len(ctx.selectVars))
	for i, v := range ctx.selectVars {
		newrow.row[i] = ctx.db.MustGetURI(r.entries[ctx.varpos[v]])
	}
	return newrow
}

func (r *row) expandFull(ctx *queryContext) []turtle.URI {
	newrow := make([]turtle.URI, len(r.vars))
	for i := range r.vars {
		newrow[i] = ctx.db.MustGetURI(r.entries[i])
	}
	return newrow
}

func (r *row) lastKey() Key {
	return r.entries[r.lastidx]
}

func (r *row) lastVar() string {
	return r.lastvar
}

func (r *row) isSet(varname string) bool {
	if varname == r.lastvar {
		return true
	}
	for idx, name := range r.vars {
		if name == varname {
			return r.entries[idx] != emptyKey
		}
	}
	return false
}

func (r *row) getValue(varname string) Key {
	if varname == r.lastvar {
		return r.entries[r.lastidx]
	}
	for idx, name := range r.vars {
		if name == varname {
			return r.entries[idx]
		}
	}
	return emptyKey
}
