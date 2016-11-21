package db

import (
	"encoding/binary"
	"github.com/google/btree"
)

/*
Need to think of the data structure and API for the item that will be storing/collating
the results of the query.
- resolve{Sub,Obj}: variable gets resolved an added to top level scope:
    addVariable(varname, tree), where 'tree' is a btree of ResultEntity
*/

type ResultEntity struct {
	PK      [4]byte
	Next    *btree.BTree
	Varname string
}

// map wrapper for storing intermediate results
type resultMap struct {
	vars     map[string]*btree.BTree
	varOrder *variableStateMap
	tuples   *btree.BTree
}

func (rm *resultMap) has(variable string) bool {
	_, found := rm.vars[variable]
	return found
}

func (rm *resultMap) filterByTree(variable string, tree *btree.BTree) {
	log.Warning("variable", variable, rm.varOrder.vars[variable])
	if rm.varOrder.varIsTop(variable) {
		if curTree, found := rm.vars[variable]; found {
			rm.vars[variable] = intersectTrees(curTree, tree)
		} else {
			rm.vars[variable] = tree
		}
	} else {
		// if variable is not at the top scope, then
	}
}

func (rm *resultMap) addVariable(variable string, tree *btree.BTree) {
	rm.vars[variable] = tree
}

// check the variable status.
// because variables can be nested, we need to figure out how to filter those out
func (rm *resultMap) getVar(variable string) *btree.BTree {
	if tree, found := rm.vars[variable]; found {
		return tree
	}
	return btree.New(3)
}

func newResultMap() *resultMap {
	return &resultMap{
		vars:   make(map[string]*btree.BTree),
		tuples: btree.New(3),
	}
}

type VariableEntity struct {
	PK [4]byte
	// a link has key: variable name, values: set of VariableEntities
	Links map[string]*btree.BTree
}

func (ve VariableEntity) Less(than btree.Item) bool {
	t := than.(*VariableEntity)
	return binary.LittleEndian.Uint32(ve.PK[:]) < binary.LittleEndian.Uint32(t.PK[:])
}
