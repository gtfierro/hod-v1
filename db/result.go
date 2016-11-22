package db

import (
	"encoding/binary"
	"github.com/google/btree"
	turtle "github.com/gtfierro/hod/goraptor"
)

/*
Need to think of the data structure and API for the item that will be storing/collating
the results of the query.
- resolve{Sub,Obj}: variable gets resolved an added to top level scope:
    addVariable(varname, tree), where 'tree' is a btree of ResultEntity
*/

type ResultEntity struct {
	PK   [4]byte
	Next map[string]*btree.BTree
}

func (re ResultEntity) Less(than btree.Item) bool {
	t := than.(*ResultEntity)
	return binary.LittleEndian.Uint32(re.PK[:]) < binary.LittleEndian.Uint32(t.PK[:])
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
	rm.vars[variable] = hashTreeToEntityTree(tree)
}

func (rm *resultMap) getVariableChain(variable string) []string {
	var _getchain func(rm *resultMap, variable string)
	chain := []string{variable}
	_getchain = func(rm *resultMap, variable string) {
		next := rm.varOrder.vars[variable]
		if next != RESOLVED {
			chain = append([]string{next}, chain...)
			_getchain(rm, next)
		}
	}
	_getchain(rm, variable)
	return chain
}

func (rm *resultMap) replaceEntity(varname string, entity *ResultEntity) bool {
	var replaceInTree func(*btree.BTree, []string, *ResultEntity) bool
	chain := rm.getVariableChain(varname)
	replaceInTree = func(tree *btree.BTree, varorder []string, entity *ResultEntity) bool {
		if tree.Has(entity) {
			tree.ReplaceOrInsert(entity)
			return true
		}
		finishedReplace := false
		iter := func(i btree.Item) bool {
			ent := i.(*ResultEntity)
			if ntree, found := ent.Next[varname]; found {
				if replaceInTree(ntree, varorder[1:], entity) {
					finishedReplace = true
					return false // stop iteration
				}
			}
			if len(varorder) == 0 {
				return i != tree.Max()
			}
			if ntree, found := ent.Next[varorder[0]]; found {
				if replaceInTree(ntree, varorder[1:], entity) {
					finishedReplace = true
					return false // stop iteration
				}
			}
			return i != tree.Max()
		}
		tree.Ascend(iter)
		return finishedReplace
	}
	return replaceInTree(rm.vars[chain[0]], chain[1:], entity)
}

// iterates through all the entries we have for variable
func (rm *resultMap) iterVariable(variable string) []*ResultEntity {
	var _iterbtree func(btree *btree.BTree, itervars []string)
	var results []*ResultEntity
	iterorder := rm.getVariableChain(variable)
	if len(iterorder) == 0 {
		panic("no order for variable " + variable)
	}
	if rm.varOrder.vars[variable] == RESOLVED { // top level
		tree := rm.vars[variable]
		iter := func(i btree.Item) bool {
			results = append(results, i.(*ResultEntity))
			return i != tree.Max()
		}
		tree.Ascend(iter)
		return results
	}
	_iterbtree = func(tree *btree.BTree, itervars []string) {
		iter := func(i btree.Item) bool {
			entity := i.(*ResultEntity)
			if len(itervars) == 0 {
				results = append(results, entity)
				return i != tree.Max()
			}
			if subtree, found := entity.Next[variable]; found {
				_iterbtree(subtree, itervars[1:])
			} else {
				_iterbtree(entity.Next[itervars[0]], itervars[1:])
			}
			return i != tree.Max()
		}
		if tree == nil {
			return
		}
		tree.Ascend(iter)
	}
	tree := rm.vars[iterorder[0]]
	_iterbtree(tree, iterorder[1:])
	return results
}

// check the variable status.
// because variables can be nested, we need to figure out how to filter those out
func (rm *resultMap) getVar(variable string) *btree.BTree {
	if tree, found := rm.vars[variable]; found {
		return tree
	}
	return nil
}

func newResultMap() *resultMap {
	return &resultMap{
		vars:   make(map[string]*btree.BTree),
		tuples: btree.New(3),
	}
}

func (db *DB) expandTuples(rm *resultMap, selectVars []string) [][]turtle.URI {
	var tuples []map[string]turtle.URI
	var startvar string
	for v, state := range rm.varOrder.vars {
		if state == RESOLVED {
			startvar = v
			break
		}
	}
	tree := rm.vars[startvar]
	iter := func(i btree.Item) bool {
		entity := i.(*ResultEntity)
		newtups := db._getTuplesFromTree(startvar, entity)
		tuples = append(tuples, newtups...)
		return i != tree.Max()
	}
	tree.Ascend(iter)

	var results [][]turtle.URI
tupleLoop:
	for _, tup := range tuples {
		var row []turtle.URI
		for _, varname := range selectVars {
			if _, found := tup[varname]; !found {
				continue tupleLoop
			}
			row = append(row, tup[varname])
		}
		results = append(results, row)
	}
	return results
}

func (db *DB) _getTuplesFromTree(name string, ve *ResultEntity) []map[string]turtle.URI {
	uri := db.MustGetURI(ve.PK)
	var ret []map[string]turtle.URI
	if len(ve.Next) == 0 {
		ret = append(ret, map[string]turtle.URI{name: uri})
	} else {
		for lname, etree := range ve.Next {
			vars := make(map[string]turtle.URI)
			vars[name] = uri
			iter := func(i btree.Item) bool {
				entity := i.(*ResultEntity)
				for _, m := range db._getTuplesFromTree(lname, entity) {
					for k, v := range m {
						vars[k] = v
					}
				}
				return i != etree.Max()
			}
			etree.Ascend(iter)
			ret = append(ret, vars)
		}
	}
	return ret
}
