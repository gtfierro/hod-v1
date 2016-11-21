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
	PK          [4]byte
	Next        *btree.BTree
	NextVarname string
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

func (rm *resultMap) replaceEntity(entity *ResultEntity) {
	var replaceInTree func(*btree.BTree, *ResultEntity) bool
	replaceInTree = func(tree *btree.BTree, entity *ResultEntity) bool {
		if tree.Has(entity) {
			tree.ReplaceOrInsert(entity)
			return true
		}
		found := false
		iter := func(i btree.Item) bool {
			ent := i.(*ResultEntity)
			if replaceInTree(ent.Next, entity) {
				found = true
				return false // stop iteration
			}
			return i != tree.Max()
		}
		tree.Ascend(iter)
		return found
	}
}

// iterates through all the entries we have for variable
func (rm *resultMap) iterVariable(variable string) chan *ResultEntity {
	var _iterbtree func(btree *btree.BTree, itervars []string)
	results := make(chan *ResultEntity)
	iterorder := rm.getVariableChain(variable)
	if len(iterorder) == 0 {
		panic("no order for variable " + variable)
	}
	go func() {
		if rm.varOrder.vars[variable] == RESOLVED { // top level
			tree := rm.vars[variable]
			iter := func(i btree.Item) bool {
				results <- i.(*ResultEntity)
				return i != tree.Max()
			}
			tree.Ascend(iter)
			close(results)
			return
		}
		_iterbtree = func(tree *btree.BTree, itervars []string) {
			iter := func(i btree.Item) bool {
				entity := i.(*ResultEntity)
				if len(itervars) == 1 {
					if itervars[0] != variable {
						panic("this should not happen")
					}
					results <- entity
					return i != tree.Max()
				}
				_iterbtree(entity.Next, itervars[1:])
				return i != tree.Max()
			}
			tree.Ascend(iter)
		}
		tree := rm.vars[iterorder[0]]
		_iterbtree(tree, iterorder[1:])
		close(results)
	}()
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
	log.Debug("start with", startvar)
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
	iter := func(i btree.Item) bool {
		entity := i.(*ResultEntity)
		if entity.Next.Len() == 0 {
			vars := make(map[string]turtle.URI)
			vars[name] = uri
			vars[ve.NextVarname] = db.MustGetURI(entity.PK)
			ret = append(ret, vars)
			return i != ve.Next.Max()
		}
		for _, m := range db._getTuplesFromTree(ve.NextVarname, entity) {
			vars := make(map[string]turtle.URI)
			vars[name] = uri
			for k, v := range m {
				vars[k] = v
			}
			ret = append(ret, vars)
		}
		return i != ve.Next.Max()
	}
	ve.Next.Ascend(iter)
	return ret
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
