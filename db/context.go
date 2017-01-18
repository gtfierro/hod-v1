package db

import (
	"container/list"
	"fmt"

	turtle "github.com/gtfierro/hod/goraptor"
	"github.com/gtfierro/hod/query"
)

var emptyTree = newPointerTree(3)

// queryContext
type queryContext struct {
	candidates       map[string]*pointerTree
	chains           map[Key]map[string]*linkRecord
	db               *DB
	traverseOrder    *list.List
	traverseVars     map[string]*list.Element
	linkedValueCache map[Key]map[string]*pointerTree
	tupleCache       map[string][]map[string]turtle.URI
	// embedded query plan
	*queryPlan
}

func newQueryContext(plan *queryPlan, db *DB) *queryContext {
	ctx := &queryContext{
		candidates:       make(map[string]*pointerTree),
		chains:           make(map[Key]map[string]*linkRecord),
		queryPlan:        plan,
		traverseOrder:    list.New(),
		traverseVars:     make(map[string]*list.Element),
		linkedValueCache: make(map[Key]map[string]*pointerTree),
		tupleCache:       make(map[string][]map[string]turtle.URI),
		db:               db,
	}
	return ctx
}

func (ctx *queryContext) dumpVarCounts() {
	for varname, tree := range ctx.candidates {
		fmt.Println("var count", varname, tree.Len())
	}
}

func (ctx *queryContext) dumpTraverseOrder() {
	elem := ctx.traverseOrder.Front()
	for elem.Next() != nil {
		varname := elem.Value.(string)
		fmt.Println(varname, "next =>", elem.Next().Value.(string))
		elem = elem.Next()
	}
	fmt.Println(elem.Value.(string))
}

// now we need to plan out the set of actions for adding/filtering vars on the query context

// returns the set of current guesses for the given variable
// returns TRUE if the tree is known, FALSE otherwise
func (ctx *queryContext) getValues(varname string) (*pointerTree, bool) {
	if tree, found := ctx.candidates[varname]; found && tree != nil {
		return tree, true
	}
	return emptyTree, false
}

func (ctx *queryContext) candidateHasValue(varname string, ent *Entity) bool {
	tree, found := ctx.candidates[varname]
	if found {
		return tree.Has(ent)
	}
	return true // no values to filter, so we just say "yes"
}

// returns the set of reachable values from the given entity
func (ctx *queryContext) getLinkedValues(varname string, ent *Entity) *pointerTree {
	if tree, found := ctx.linkedValueCache[ent.PK][varname]; found {
		return tree
	}
	var res = newPointerTree(3)
	chain := ctx.chains[ent.PK][varname]
	if chain != nil {
		for _, link := range chain.links {
			res.Add(ctx.db.MustGetEntityFromHash(link.me))
		}
	}
	if _, found := ctx.linkedValueCache[ent.PK]; !found {
		ctx.linkedValueCache[ent.PK] = make(map[string]*pointerTree)
	}
	ctx.linkedValueCache[ent.PK][varname] = res
	return res
}

// if we already have values for the given variable name, we filter the values given by those
// (take the intersection). Else we just keep the provided values the same
func (ctx *queryContext) filterIfDefined(varname string, values *pointerTree) *pointerTree {
	if tree, found := ctx.getValues(varname); found {
		values = intersectPointerTrees(tree, values)
	}
	return values
}

func (ctx *queryContext) define(varname string, values *pointerTree) {
	ctx.candidates[varname] = values
	_, found := ctx.traverseVars[varname]
	if !found {
		elem := ctx.traverseOrder.PushBack(varname)
		ctx.traverseVars[varname] = elem
	}
}

// if values don't exist for the variable w/n this context, then we just add these values
// if values DO already exist, then we take the intersection
func (ctx *queryContext) addOrFilterVariable(varname string, values *pointerTree) {
	if oldValues, exists := ctx.candidates[varname]; exists {
		ctx.candidates[varname] = intersectPointerTrees(oldValues, values)
		values = ctx.candidates[varname]
	} else {
		ctx.candidates[varname] = values
	}

	_, found := ctx.traverseVars[varname]
	if !found {
		elem := ctx.traverseOrder.PushBack(varname)
		ctx.traverseVars[varname] = elem
	}

}

// unions, not intersects
func (ctx *queryContext) addOrMergeVariable(varname string, values *pointerTree) {
	if oldValues, exists := ctx.candidates[varname]; exists {
		mergePointerTrees(oldValues, values)
		ctx.candidates[varname] = oldValues
	} else {
		ctx.candidates[varname] = values
	}

	_, found := ctx.traverseVars[varname]
	if !found {
		elem := ctx.traverseOrder.PushBack(varname)
		ctx.traverseVars[varname] = elem
	}
}
func (ctx *queryContext) addReachable(parent *Entity, parentVar string, reachable *pointerTree, reachableVar string) {
	chain, found := ctx.chains[parent.PK][reachableVar]
	if !found {
		chain = &linkRecord{me: parent.PK}
	}
	reachable.mergeOntoLinkRecord(chain)
	if _, found := ctx.chains[parent.PK]; !found {
		ctx.chains[parent.PK] = make(map[string]*linkRecord)
	}
	ctx.chains[parent.PK][reachableVar] = chain

	parentElem, found := ctx.traverseVars[parentVar]
	if !found {
		parentElem = ctx.traverseOrder.PushBack(parentVar)
		ctx.traverseVars[parentVar] = parentElem
	}

	childElem, found := ctx.traverseVars[reachableVar]
	if found {
		ctx.traverseOrder.MoveAfter(childElem, parentElem)
	} else {
		elem := ctx.traverseOrder.InsertAfter(reachableVar, parentElem)
		ctx.traverseVars[reachableVar] = elem
	}
}

func (ctx *queryContext) addReachableSingle(parent *Entity, parentVar string, reachable *Entity, reachableVar string) {
	chain, found := ctx.chains[parent.PK][reachableVar]
	if !found {
		chain = &linkRecord{me: parent.PK}
	}
	// add on this one record
	chain.links = append(chain.links, &linkRecord{me: reachable.PK})
	if _, found := ctx.chains[parent.PK]; !found {
		ctx.chains[parent.PK] = make(map[string]*linkRecord)
	}
	ctx.chains[parent.PK][reachableVar] = chain

	parentElem, found := ctx.traverseVars[parentVar]
	if !found {
		parentElem = ctx.traverseOrder.PushBack(parentVar)
		ctx.traverseVars[parentVar] = parentElem
	}

	childElem, found := ctx.traverseVars[reachableVar]
	if found {
		ctx.traverseOrder.MoveAfter(childElem, parentElem)
	} else {
		elem := ctx.traverseOrder.InsertAfter(reachableVar, parentElem)
		ctx.traverseVars[reachableVar] = elem
	}
}

// returns true if any vars are reachable from this entity
func (ctx *queryContext) entityHasFollowers(ent *Entity) bool {
	if ent == nil || ent.PK == emptyHash {
		return false
	}
	if m, found := ctx.chains[ent.PK]; found && len(m) > 0 {
		for _, links := range m {
			if len(links.links) > 0 {
				return true
			}
		}
	}
	return false
}

func (ctx *queryContext) expandTuples() [][]turtle.URI {
	var (
		startvar string
		results  [][]turtle.URI
		tuples   []map[string]turtle.URI
	)
	// choose first variable
	for v, state := range ctx.vars {
		if state == RESOLVED {
			startvar = v
			break
		}
	}
	if len(startvar) == 0 {
		// need to choose the "parent" if there is no RESOLVED variable
		for _, parent := range ctx.vars {
			if _, exists := ctx.vars[parent]; !exists {
				startvar = parent
				break
			}
		}
	}

	topVarTree := ctx.candidates[startvar]
	if topVarTree == nil {
		return results // fail early
	}
	max := topVarTree.Max()
	iter := func(ent *Entity) bool {
		newtups := ctx._getTuplesFromTree(startvar, ent)
		tuples = append(tuples, newtups...)
		return ent != max
	}
	topVarTree.Iter(iter)
tupleLoop:
	for _, tup := range tuples {
		var row []turtle.URI
		for _, varname := range ctx.selectVars {
			if _, found := tup[varname]; !found {
				if ctx.query.Select.Partial {
					continue
				} else {
					continue tupleLoop
				}
			}
			row = append(row, tup[varname])
		}
		results = append(results, row)
		if ctx.query.Select.Limit > 0 && len(results) == ctx.query.Select.Limit {
			return results
		}
	}
	return results
}

func (ctx *queryContext) _getTuplesFromTree(name string, ent *Entity) []map[string]turtle.URI {
	var ret []map[string]turtle.URI
	if ent == nil || ent.PK == emptyHash {
		return ret
	}
	uri := ctx.db.MustGetURI(ent.PK)
	if ret, found := ctx.tupleCache[name+ent.PK.String()]; found {
		return ret
	}

	vars := make(map[string]turtle.URI)
	vars[name] = uri
	if !ctx.entityHasFollowers(ent) {
		ret = append(ret, map[string]turtle.URI{name: uri})
	} else {
		// loop through the values of the child var
		for childName, _links := range ctx.chains[ent.PK] {
			if len(_links.links) == 0 {
				continue
			}
			if childName == name {
				log.Debug("continuing from", uri, childName)
				log.Debug(ret)
				ret = append(ret, map[string]turtle.URI{name: uri})
				continue
			}
			childValues := ctx.getLinkedValues(childName, ent)
			max := childValues.Max()
			iter := func(child *Entity) bool {
				for _, m := range ctx._getTuplesFromTree(childName, child) {
					for k, v := range m {
						vars[k] = v
					}
					// when we want to append, make sure to allocate a new map
					newvar := make(map[string]turtle.URI)
					for k, v := range vars {
						newvar[k] = v
					}
					ret = append(ret, newvar)
				}
				return child != max
			}
			childValues.Iter(iter)
		}
	}
	ctx.tupleCache[name+ent.PK.String()] = ret
	return ret
}

const (
	RESOLVED   = "RESOLVED"
	UNRESOLVED = ""
)

// contains all useful state information for executing a query
type queryPlan struct {
	operations []operation
	selectVars []string
	dg         *dependencyGraph
	query      query.Query
	vars       map[string]string
}

func newQueryPlan(dg *dependencyGraph, q query.Query) *queryPlan {
	plan := &queryPlan{
		selectVars: dg.selectVars,
		dg:         dg,
		query:      q,
		vars:       make(map[string]string),
	}
	return plan
}

func (qp *queryPlan) dumpVarchain() {
	for k, v := range qp.vars {
		fmt.Println(k, "=>", v)
	}
}

func (plan *queryPlan) hasVar(variable string) bool {
	return plan.vars[variable] != UNRESOLVED
}

func (plan *queryPlan) varIsChild(variable string) bool {
	return plan.hasVar(variable) && plan.vars[variable] != RESOLVED
}

func (plan *queryPlan) varIsTop(variable string) bool {
	return plan.hasVar(variable) && plan.vars[variable] == RESOLVED
}

func (plan *queryPlan) addTopLevel(variable string) {
	plan.vars[variable] = RESOLVED
}

func (plan *queryPlan) addLink(parent, child string) {
	plan.vars[child] = parent
}
