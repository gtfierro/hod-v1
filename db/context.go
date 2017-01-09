package db

// queryContext
type queryContext struct {
	candidates map[string]*pointerTree
	chains     map[Key]linkRecord
	db         *DB
	// embedded query plan
	*queryPlan
}

func newQueryContext(plan *queryPlan, db *DB) *queryContext {
	ctx := &queryContext{
		queryPlan: plan,
		db:        db,
	}
	return ctx
}

// now we need to plan out the set of actions for adding/filtering vars on the query context

// if values don't exist for the variable w/n this context, then we just add these values
// if values DO already exist, then we take the intersection
func (ctx *queryContext) addOrFilterVariable(varname string, values *pointerTree) {
	if oldValues, exists := ctx.candidates[varname]; exists {
		ctx.candidates[varname] = intersectPointerTrees(oldValues, values)
	} else {
		ctx.candidates[varname] = values
	}
}

func (ctx *queryContext) linkValues(parent *Entity, reachable pointerTree) {
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
