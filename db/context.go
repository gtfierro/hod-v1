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
	_traverseOrder   *varlist
	traverseVars     map[string]*list.Element
	linkedValueCache map[Key]map[string]*pointerTree
	tupleCache       map[string][]map[string]turtle.URI
	vardepth         map[string]int
	varpos           map[string]int
	// embedded query plan
	*queryPlan
}

func newQueryContext(plan *queryPlan, db *DB) *queryContext {
	ctx := &queryContext{
		candidates:       make(map[string]*pointerTree),
		chains:           make(map[Key]map[string]*linkRecord),
		queryPlan:        plan,
		_traverseOrder:   newvarlist(),
		traverseVars:     make(map[string]*list.Element),
		linkedValueCache: make(map[Key]map[string]*pointerTree),
		tupleCache:       make(map[string][]map[string]turtle.URI),
		vardepth:         make(map[string]int),
		varpos:           make(map[string]int),
		db:               db,
	}
	return ctx
}

func (ctx *queryContext) dumpVarCounts() {
	for varname, tree := range ctx.candidates {
		fmt.Println("var count", varname, tree.Len())
	}
}

func (ctx *queryContext) dumpChildren() {
	for varname, tree := range ctx.candidates {
		if tree.Len() == 0 {
			continue
		}
		fmt.Println("var ", varname, "has children")
		i := tree.Max()
		for varname, links := range ctx.chains[i.PK] {
			fmt.Println("   =>", varname, len(links.links))
		}

	}
}

func (ctx *queryContext) dumpTraverseOrder() {
	length := len(ctx._traverseOrder.list)
	for i := 0; i < length-1; i++ {
		cur := ctx._traverseOrder.list[i]
		next := ctx._traverseOrder.list[i+1]
		fmt.Println(cur.value, "next =>", next.value)
	}
	fmt.Println("children")
	ctx.dumpChildren()
	// NEW STUFF
	fmt.Println("-----")
	for varname, entry := range ctx._traverseOrder.lookup {
		if entry._prev != nil {
			fmt.Println(varname, "<=", entry._prev.value)
		} else {
			fmt.Println(varname)
		}
	}
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

	if !ctx._traverseOrder.has(varname) {
		ctx._traverseOrder.pushBack(varname)
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

	if !ctx._traverseOrder.has(varname) {
		ctx._traverseOrder.pushBack(varname)
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

	if !ctx._traverseOrder.has(varname) {
		ctx._traverseOrder.pushBack(varname)
	}
}
func (ctx *queryContext) addReachable(parent *Entity, parentVar string, reachable *pointerTree, reachableVar string) {
	log.Error("add", reachableVar, "reachable from", parentVar)
	chain, found := ctx.chains[parent.PK][reachableVar]
	if !found {
		chain = &linkRecord{me: parent.PK}
	}
	reachable.mergeOntoLinkRecord(chain)
	if _, found := ctx.chains[parent.PK]; !found {
		ctx.chains[parent.PK] = make(map[string]*linkRecord)
	}
	ctx.chains[parent.PK][reachableVar] = chain

	if !ctx._traverseOrder.has(parentVar) {
		ctx._traverseOrder.pushBack(parentVar)
	}
	if ctx._traverseOrder.has(reachableVar) {
		ctx._traverseOrder.moveAfter(reachableVar, parentVar)
	} else {
		ctx._traverseOrder.insertAfter(reachableVar, parentVar)
	}

	ctx.dumpTraverseOrder()
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

	if !ctx._traverseOrder.has(parentVar) {
		ctx._traverseOrder.pushBack(parentVar)
	}
	if ctx._traverseOrder.has(reachableVar) {
		ctx._traverseOrder.moveAfter(reachableVar, parentVar)
	} else {
		ctx._traverseOrder.insertAfter(reachableVar, parentVar)
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
		results [][]turtle.URI
	)
	return results
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

func (ctx *queryContext) expandtuples2() [][]turtle.URI {
	var (
		startvar string
		results  [][]turtle.URI
		//tuples   []map[string]turtle.URI
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

	idx := 0
	startidx := ctx._traverseOrder.indexes[startvar]
	for i := startidx; i < len(ctx._traverseOrder.list); i++ {
		ctx.varpos[ctx._traverseOrder.list[i].value] = i
		idx++
	}

	//// determine the position in the results row for the given variable
	//i := 0
	//elem := ctx.traverseVars[startvar]
	//for elem != nil {
	//	ctx.varpos[elem.Value.(string)] = i
	//	i++
	//	elem = elem.Next()
	//}

	topVarTree := ctx.candidates[startvar]
	if topVarTree == nil {
		return results // fail early
	}

	max := topVarTree.Max()
	iter := func(ent *Entity) bool {
		results = append(results, ctx.expandEntity(startvar, ent)...)
		return ent != max
	}
	topVarTree.Iter(iter)

	log.Error(len(results))
	return results
}

// generate all the paths from a given entity.
// If this is recursive, then we are going to have a LOT of mini-allocations. Ideally, we can just do this
// top-level method on the first tree and then be "done".

// TODO: problem is we need to keep track of the "last" item added to each row and keep track of the order
//t hey were added in. Either have a convenience method that goes backwards from the last traverseVar to find
// the item in the row, or create a wrapper c lass that keeps track of that
// (look at lastFilledIndex)
func (ctx *queryContext) expandEntity(varname string, entity *Entity) [][]turtle.URI {
	var (
		rows         [][]turtle.URI
		nextVariable string
	)
	if entity == nil || entity.PK == emptyHash {
		return rows
	}

	var varorder []string
	for _, val := range ctx._traverseOrder.list {
		varorder = append(varorder, val.value)
	}

	//elem := ctx.traverseVars[varname]
	//for elem != nil {
	//	varorder = append(varorder, elem.Value.(string))
	//	elem = elem.Next()
	//}

	lastFilledIndex := func(row []Key) int {
		for i := len(varorder) - 1; i >= 0; i-- {
			if row[ctx.varpos[varorder[i]]] != emptyHash {
				//log.Debug("nonempty hash at i", i, "varord", varorder[i], "varpos", ctx.varpos[varorder[i]])
				return ctx.varpos[varorder[i]]
			}
		}
		return -1
	}

	//log.Debug("URI", ctx.db.MustGetURI(entity.PK), varname, varorder)
	newRow := func() []Key {
		row := make([]Key, len(varorder))
		row[ctx.varpos[varname]] = entity.PK
		return row
	}

	expand := func(row []Key) []turtle.URI {
		newrow := make([]turtle.URI, len(ctx.selectVars))
		for i, v := range ctx.selectVars {
			newrow[i] = ctx.db.MustGetURI(row[ctx.varpos[v]])
		}
		return newrow
	}

	stack := list.New()
	stack.PushFront(newRow())
	for stack.Len() > 0 {
		// get the row
		row := stack.Remove(stack.Front()).([]Key)
		// if it is full, then we add it to the list to be returned
		// because it is complete
		if rowIsFull(row) {
			rows = append(rows, expand(row))
			continue
		}
		// get the position of last filled element
		lastIdx := lastFilledIndex(row)
		log.Error(ctx._traverseOrder.buildVarOrder())
		if lastIdx+1 == len(varorder) {
			log.Error("on", varorder[lastIdx], lastIdx, varorder)
			log.Error("row is", expand(row))
			continue
		}
		nextVariable = varorder[lastIdx+1]
		// varorder[lastIdx] is which variable we are "on" right now
		// varorder[lastIdx+1] is the variable we want to traverse next. Follow this link

		// We are trying to follow the traverse order here (kept in structs varorder
		// and ctx.varpos for index- and varname- keyed indexes).
		// We want to figure out where we are in the traversal list, and follow the linked
		// values for the NEXT variable on the traversal list from the element where we
		// currently are.

		// get all of the links
		children := ctx.chains[row[lastIdx]]
		//log.Debug(varorder)
		//log.Debug("at", varorder[lastIdx])
		//for v, l := range children {
		//	log.Debug(v, len(l.links))
		//}
		//log.Debug("at", lastIdx, varorder[lastIdx])
		//log.Debug("next", lastIdx+1, varorder[lastIdx+1])

		childLinks, found := children[nextVariable]
		// if the links didn't contain an entry for the next traverse variable, then we backtrack
		// up the list of traversed variables to check for downward links for the variable
		// we want to traverse next
		_tmpidx := lastIdx
		log.Debug(varorder)
		for !found && _tmpidx > 0 {
			log.Debug("could not find", nextVariable, "from", varorder[_tmpidx])
			_tmpidx--
			prevVariable := varorder[_tmpidx]
			prevIndex := ctx.varpos[prevVariable]
			children = ctx.chains[row[prevIndex]]
			childLinks, found = children[nextVariable]
		}
		//for childLinks == nil && _tmpidx > 0 { // next var is not reachable from where we are, so we backtrack
		//	_tmpidx--
		//	log.Debug(varorder, _tmpidx, ctx.db.MustGetURI(row[_tmpidx]))
		//	if children, found = ctx.chains[row[_tmpidx]]; found {
		//		log.Debug("found at", varorder[_tmpidx])
		//		childLinks = children[varorder[lastIdx+1]]
		//	}
		//}
		if childLinks == nil {
			log.Warning("giveup")
			continue // no hope
		}
		//log.Debug("next has", len(childLinks.links))
		for _, child := range childLinks.links {
			row[lastIdx+1] = child.me
			newrow := newRow()
			copy(newrow, row)
			stack.PushBack(newrow)
		}
	}

	return rows
}
