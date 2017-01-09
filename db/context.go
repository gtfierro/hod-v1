package db

import (
	turtle "github.com/gtfierro/hod/goraptor"
)

// queryContext
type queryContext struct {
	candidates map[string]*pointerTree
	chains     map[Key]*linkRecord
	db         *DB
	// embedded query plan
	*queryPlan
}

func newQueryContext(plan *queryPlan, db *DB) *queryContext {
	ctx := &queryContext{
		candidates: make(map[string]*pointerTree),
		chains:     make(map[Key]*linkRecord),
		queryPlan:  plan,
		db:         db,
	}
	return ctx
}

// now we need to plan out the set of actions for adding/filtering vars on the query context

// returns the set of current guesses for the given variable
func (ctx *queryContext) getValues(varname string) *pointerTree {
	if tree, found := ctx.candidates[varname]; found && tree != nil {
		return tree
	}
	ctx.candidates[varname] = newPointerTree(3)
	return ctx.candidates[varname]
}

// if values don't exist for the variable w/n this context, then we just add these values
// if values DO already exist, then we take the intersection
func (ctx *queryContext) addOrFilterVariable(varname string, values *pointerTree) {
	if oldValues, exists := ctx.candidates[varname]; exists {
		ctx.candidates[varname] = intersectPointerTrees(oldValues, values)
	} else {
		ctx.candidates[varname] = values
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
}

func (ctx *queryContext) addReachable(parent *Entity, reachable *pointerTree) {
	chain, found := ctx.chains[parent.PK]
	if !found {
		chain = &linkRecord{me: parent.PK}
	}
	reachable.mergeOntoLinkRecord(chain)
	ctx.chains[parent.PK] = chain
}

func (ctx *queryContext) expandTuples() [][]turtle.URI {
	var (
		startvar string
		results  [][]turtle.URI
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
	max := topVarTree.Max()
	iter := func(ent *Entity) bool {
		results = append(results, []turtle.URI{ctx.db.MustGetURI(ent.PK)})
		return ent != max
	}
	topVarTree.Iter(iter)
	// now for each of these, we traverse the link records
	//length := 1
	//for idx, value := range results {
	//	log.Debug(idx, value, ctx.chains[value[length]])
	//}
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
	vars       map[string]string
}

func newQueryPlan(dg *dependencyGraph) *queryPlan {
	plan := &queryPlan{
		selectVars: dg.selectVars,
		dg:         dg,
		vars:       make(map[string]string),
	}
	return plan
}

func (qp *queryPlan) dumpVarchain() {
	for k, v := range qp.vars {
		log.Debug(k, "=>", v)
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
