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
	chains           map[Key]*linkRecord
	db               *DB
	traverseOrder    *list.List
	traverseVars     map[string]*list.Element
	linkedValueCache map[Key]*pointerTree
	tupleCache       map[string][]map[string]turtle.URI
	// embedded query plan
	*queryPlan
}

func newQueryContext(plan *queryPlan, db *DB) *queryContext {
	ctx := &queryContext{
		candidates:       make(map[string]*pointerTree),
		chains:           make(map[Key]*linkRecord),
		queryPlan:        plan,
		traverseOrder:    list.New(),
		traverseVars:     make(map[string]*list.Element),
		linkedValueCache: make(map[Key]*pointerTree),
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

// returns the set of reachable values from the given entity
func (ctx *queryContext) getLinkedValues(ent *Entity) *pointerTree {
	if tree, found := ctx.linkedValueCache[ent.PK]; found {
		return tree
	}
	var res = newPointerTree(3)
	chain := ctx.chains[ent.PK]
	if chain != nil {
		for _, link := range chain.links {
			res.Add(ctx.db.MustGetEntityFromHash(link.me))
		}
	}
	ctx.linkedValueCache[ent.PK] = res
	return res
}

// if values don't exist for the variable w/n this context, then we just add these values
// if values DO already exist, then we take the intersection
func (ctx *queryContext) addOrFilterVariable(varname string, values *pointerTree) {
	if oldValues, exists := ctx.candidates[varname]; exists {
		ctx.candidates[varname] = intersectPointerTrees(oldValues, values)
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
	chain, found := ctx.chains[parent.PK]
	if !found {
		chain = &linkRecord{me: parent.PK}
	}
	reachable.mergeOntoLinkRecord(chain)
	ctx.chains[parent.PK] = chain

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
	link, found := ctx.chains[ent.PK]
	if !found || link == nil {
		return false
	}
	return len(link.links) > 0
}

// gets the name of the next variable
func (ctx *queryContext) getChild(varname string) string {
	elem := ctx.traverseVars[varname].Next()
	if elem != nil {
		return elem.Value.(string)
	}
	return ""
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
	childName := ctx.getChild(name)
	if !ctx.entityHasFollowers(ent) || childName == "" {
		ret = append(ret, map[string]turtle.URI{name: uri})
	} else {
		// loop through the values of the child var
		childValues := ctx.getLinkedValues(ent)
		candidateChildValues, hasRestrictions := ctx.getValues(childName)
		max := childValues.Max()
		iter := func(child *Entity) bool {
			if hasRestrictions && !candidateChildValues.Has(child) {
				return child != max
			}
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

func (qp *queryPlan) findVarDepth(target string) int {
	var depth = 0
	start := qp.vars[target]
	for start != RESOLVED {
		start = qp.vars[start]
		depth += 1
	}
	return depth
}

func (qp *queryPlan) Len() int {
	return len(qp.operations)
}
func (qp *queryPlan) Swap(i, j int) {
	qp.operations[i], qp.operations[j] = qp.operations[j], qp.operations[i]
}
func (qp *queryPlan) Less(i, j int) bool {
	iDepth := qp.findVarDepth(qp.operations[i].SortKey())
	jDepth := qp.findVarDepth(qp.operations[j].SortKey())
	return iDepth < jDepth
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
