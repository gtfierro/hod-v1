package db

import (
	turtle "github.com/gtfierro/hod/goraptor"
)

type row struct {
	// entries in this row
	entries   []Key
	vars      []string
	lastvar   string
	lastidx   int
	numFilled int
}

func newrow(vars []string) *row {
	return &row{
		entries:   make([]Key, len(vars)),
		vars:      vars,
		numFilled: 0,
	}
}

func (r *row) isFull() bool {
	for _, entry := range r.entries {
		if entry == emptyHash {
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

func (r *row) expand(ctx *queryContext) []turtle.URI {
	newrow := make([]turtle.URI, len(ctx.selectVars))
	for i, v := range ctx.selectVars {
		newrow[i] = ctx.db.MustGetURI(r.entries[ctx.varpos[v]])
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
			return r.entries[idx] != emptyHash
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
	return emptyHash
}
