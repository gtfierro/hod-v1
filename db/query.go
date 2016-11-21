package db

import (
	"container/list"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/google/btree"
	turtle "github.com/gtfierro/hod/goraptor"
	query "github.com/gtfierro/hod/query"
	"github.com/syndtr/goleveldb/leveldb"
)

// make a set of structs that capture what these queries want to do

// So, how do queries work?
// We have a list of filters, each of which has a subject, list of predicate things, and
// an object. Any of these might be variables, which we can distinguish by having a "?"
// in front of the value.
//
// First we "clean" these by making sure that they have their full
// namespaces rather than the prefix

type Item [4]byte

func (i Item) Less(than btree.Item) bool {
	t := than.(Item)
	return binary.LittleEndian.Uint32(i[:]) < binary.LittleEndian.Uint32(t[:])
}

type queryRun struct {
	plan      *dependencyGraph
	variables map[string]*btree.BTree
	vars      map[string]*btree.BTree
}

func makeQueryRun(plan *queryPlan) *queryRun {
	qr := &queryRun{
		//plan:      plan,
		variables: make(map[string]*btree.BTree),
		vars:      make(map[string]*btree.BTree),
	}
	//for _, v := range plan.selectVars {
	//	qr.variables[v] = btree.New(3)
	//	qr.vars[v] = btree.New(3)
	//}
	return qr
}

func (db *DB) RunQuery(q query.Query) {
	// "clean" the query by expanding out the prefixes
	for idx, filter := range q.Where {
		if !strings.HasPrefix(filter.Subject.Value, "?") {
			if full, found := db.namespaces[filter.Subject.Namespace]; found {
				filter.Subject.Namespace = full
			}
			q.Where[idx] = filter
		}
		if !strings.HasPrefix(filter.Object.Value, "?") {
			if full, found := db.namespaces[filter.Object.Namespace]; found {
				filter.Object.Namespace = full
			}
			q.Where[idx] = filter
		}
		for idx2, pred := range filter.Path {
			if !strings.HasPrefix(pred.Predicate.Value, "?") {
				if full, found := db.namespaces[pred.Predicate.Namespace]; found {
					pred.Predicate.Namespace = full
				}
				filter.Path[idx2] = pred
			}
		}
		q.Where[idx] = filter
	}

	fmt.Println("-------------- start query plan -------------")
	planStart := time.Now()
	dg := db.formDependencyGraph(q)
	qp := db.formQueryPlan(dg)
	for _, op := range qp.operations {
		log.Debug("op", op)
	}
	log.Infof("Formed execution plan in %s", time.Since(planStart))
	fmt.Println("-------------- end query plan -------------")

	runStart := time.Now()
	rm := db.executeQueryPlan(qp)
	log.Infof("Ran query in %s", time.Since(runStart))

	runStart = time.Now()
	results := db.expandTuples(rm, qp.selectVars)
	log.Infof("Expanded tuples in %s", time.Since(runStart))
	if q.Select.Count {
		fmt.Println(len(results))
	} else {
		for _, r := range results {
			fmt.Println(r)
		}
	}
	return
}

// retrieves for each of the variables in the vars, get each of its Links, etc etc
func (db *DB) getTuples(qr *queryRun) [][]turtle.URI {
	var tuples []map[string]turtle.URI
	for varname, tree := range qr.vars {
		fmt.Println("varname", varname, "has", tree.Len(), "entries")
		iter := func(i btree.Item) bool {
			entity := i.(*VariableEntity)
			newtups := db._getTuplesFromLinks(varname, entity)
			tuples = append(tuples, newtups...)
			return i != tree.Max()
		}
		tree.Ascend(iter)
		//for stack.Len() > 0 {
		//	row := make(map[string]turtle.URI)
		//	entity := stack.Remove(stack.Front()).(*VariableEntity)
		//	row[varname] = db.MustGetURI(entity.PK)
		//	for linkname, tree := range entity.Links {
		//		fmt.Println("=>", "var", varname, "link", linkname, "num", tree.Len())
		//	}
		//}
	}
	var results [][]turtle.URI
tupleLoop:
	for _, tup := range tuples {
		var row []turtle.URI
		for _, varname := range qr.plan.selectVars {
			if _, found := tup[varname]; !found {
				continue tupleLoop
			}
			row = append(row, tup[varname])
		}
		results = append(results, row)
	}
	return results
}

func (db *DB) _getTuplesFromLinks(name string, ve *VariableEntity) []map[string]turtle.URI {
	uri := db.MustGetURI(ve.PK)
	var ret []map[string]turtle.URI
	if len(ve.Links) == 0 {
		ret = append(ret, map[string]turtle.URI{name: uri})
	} else {
		for lname, etree := range ve.Links {
			vars := make(map[string]turtle.URI)
			vars[name] = uri
			iter := func(i btree.Item) bool {
				entity := i.(*VariableEntity)
				for _, m := range db._getTuplesFromLinks(lname, entity) {
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

// We need an execution plan for the list of filters contained in a query. How do we do this?
func (db *DB) formDependencyGraph(q query.Query) *dependencyGraph {
	dg := makeDependencyGraph(q)
	terms := make([]*queryTerm, len(q.Where))
	for i, f := range q.Where {
		terms[i] = dg.makeQueryTerm(f)
	}

	numUnresolved := func(qt *queryTerm) int {
		num := 0
		for _, v := range qt.variables {
			if !dg.variables[v] {
				num++
			}
		}
		return num
	}

	for len(terms) > 0 {
		// first find all the terms with 0 or 1 unresolved variable terms
		var added = []*queryTerm{}
		for _, term := range terms {
			if numUnresolved(term) < 2 {
				dg.addRootTerm(term)
				added = append(added, term)
			}
		}
		// remove the terms that we added to the root set
		terms = filterTermList(terms, added)
		added = []*queryTerm{}
		for _, term := range terms {
			if dg.addChild(term) {
				added = append(added, term)
			}
		}
		terms = filterTermList(terms, added)
	}
	dg.dump()
	return dg
}

func (db *DB) executeQueryPlan(qp *queryPlan) *resultMap {
	rm := newResultMap()
	rm.varOrder = qp.varOrder
	var err error
	for _, op := range qp.operations {
		rm, err = op.run(db, qp.varOrder, rm)
		if err != nil {
			log.Fatal(err)
		}
	}
	//for vname := range qp.varOrder.vars {
	//	fmt.Println(vname, rm.getVariableChain(vname))
	//	for re := range rm.iterVariable(vname) {
	//		fmt.Println(vname, re)
	//	}
	//}
	for vname, tree := range rm.vars {
		fmt.Println(vname, tree.Len())
	}
	return rm
}

// TODO: change to use compound predicates
// takes the inverse of every relationship. If no inverse exists, returns nil
func (db *DB) reversePathPattern(path []query.PathPattern) []query.PathPattern {
	var reverse = make([]query.PathPattern, len(path))
	for idx, pred := range path {
		if inverse, found := db.relationships[pred.Predicate]; found {
			pred.Predicate = inverse
			reverse[idx] = pred
		} else {
			return nil
		}
	}
	return reverse
}

// follow the pattern from the given object's InEdges, placing the results in the btree
func (db *DB) followPathFromObject(object *Entity, results *btree.BTree, searchstack *list.List, pattern query.PathPattern) {
	stack := list.New()
	stack.PushFront(object)

	predHash, err := db.GetHash(pattern.Predicate)
	if err != nil && err == leveldb.ErrNotFound {
		log.Infof("Adding unseen predicate %s", pattern.Predicate)
		var hashdest [4]byte
		if err := db.insertEntity(pattern.Predicate, hashdest[:]); err != nil {
			panic(fmt.Errorf("Could not insert entity %s (%v)", pattern.Predicate, err))
		}
	} else if err != nil {
		panic(fmt.Errorf("Not found: %v (%s)", pattern.Predicate, err))
	}

	for stack.Len() > 0 {
		entity := stack.Remove(stack.Front()).(*Entity)
		switch pattern.Pattern {
		case query.PATTERN_SINGLE:
			// [found] indicates whether or not we have any edges with the given pattern
			edges, found := entity.InEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range edges {
				nextEntity := db.MustGetEntityFromHash(entityHash)
				results.ReplaceOrInsert(Item(nextEntity.PK))
				searchstack.PushBack(nextEntity)
			}
			// because this is one hop, we don't add any new entities to the stack
		case query.PATTERN_ZERO_ONE:
			log.Notice("PATH ?", pattern)
			// this does not require the pattern to exist, so we add ALL entities connected
			// by ALL edges
			for _, endpointHashList := range entity.InEdges {
				for _, entityHash := range endpointHashList {
					nextEntity := db.MustGetEntityFromHash(entityHash)
					results.ReplaceOrInsert(Item(nextEntity.PK))
					searchstack.PushBack(nextEntity)
				}
			}
			// because this is one hop, we don't add any new entities to the stack
		case query.PATTERN_ZERO_PLUS:
			log.Notice("PATH *", pattern)
		case query.PATTERN_ONE_PLUS:
			edges, found := entity.InEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range edges {
				nextEntity := db.MustGetEntityFromHash(entityHash)
				results.ReplaceOrInsert(Item(nextEntity.PK))
				searchstack.PushBack(nextEntity)
				// also make sure to add this to the stack so we can search
				stack.PushBack(nextEntity)
			}
		}
	}
}

// follow the pattern from the given subject's OutEdges, placing the results in the btree
func (db *DB) followPathFromSubject(subject *Entity, results *btree.BTree, searchstack *list.List, pattern query.PathPattern) {
	stack := list.New()
	stack.PushFront(subject)

	predHash, err := db.GetHash(pattern.Predicate)
	if err != nil {
		panic(err)
	}

	for stack.Len() > 0 {
		entity := stack.Remove(stack.Front()).(*Entity)
		switch pattern.Pattern {
		case query.PATTERN_SINGLE:
			// [found] indicates whether or not we have any edges with the given pattern
			edges, found := entity.OutEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range edges {
				nextEntity := db.MustGetEntityFromHash(entityHash)
				results.ReplaceOrInsert(Item(nextEntity.PK))
				searchstack.PushBack(nextEntity)
			}
			// because this is one hop, we don't add any new entities to the stack
		case query.PATTERN_ZERO_ONE:
			// this does not require the pattern to exist, so we add ALL entities connected
			// by ALL edges
			for _, endpointHashList := range entity.OutEdges {
				for _, entityHash := range endpointHashList {
					nextEntity := db.MustGetEntityFromHash(entityHash)
					results.ReplaceOrInsert(Item(nextEntity.PK))
					searchstack.PushBack(nextEntity)
				}
			}
			// because this is one hop, we don't add any new entities to the stack
		case query.PATTERN_ZERO_PLUS:
		case query.PATTERN_ONE_PLUS:
			edges, found := entity.OutEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range edges {
				nextEntity := db.MustGetEntityFromHash(entityHash)
				results.ReplaceOrInsert(Item(nextEntity.PK))
				searchstack.PushBack(nextEntity)
				// also make sure to add this to the stack so we can search
				stack.PushBack(nextEntity)
			}
		}
	}
}

func (db *DB) getSubjectFromPredObject(objectHash [4]byte, path []query.PathPattern) *btree.BTree {
	// first get the initial object entity from the db
	// then we're going to conduct a BFS search starting from this entity looking for all entities
	// that have the required path sequence. We place the results in a BTree to maintain uniqueness

	// So how does this traversal actually work?
	// At each 'step', we are looking at an entity and some offset into the path.

	// get the object, look in its "in" edges for the path pattern
	objEntity, err := db.GetEntityFromHash(objectHash)
	if err != nil {
		panic(err)
	}

	results := btree.New(2)

	stack := list.New()
	stack.PushFront(objEntity)

	for stack.Len() > 0 {
		entity := stack.Remove(stack.Front()).(*Entity)
		for _, pat := range path {
			db.followPathFromObject(entity, results, stack, pat)
		}
	}

	return results
}

// TODO: change to use compound predicates
// Given object and predicate, get all subjects
func (db *DB) getObjectFromSubjectPred(subjectHash [4]byte, path []query.PathPattern) *btree.BTree {
	subEntity, err := db.GetEntityFromHash(subjectHash)
	if err != nil {
		panic(err)
	}

	results := btree.New(2)

	stack := list.New()
	stack.PushFront(subEntity)

	for stack.Len() > 0 {
		entity := stack.Remove(stack.Front()).(*Entity)
		for _, pat := range path {
			db.followPathFromSubject(entity, results, stack, pat)
		}
	}

	return results
}

// TODO: change to use compound predicates
// Given a predicate, it returns pairs of (subject, object) that are connected by that relationship
func (db *DB) getSubjectObjectFromPred(pattern query.PathPattern) (soPair [][][4]byte) {
	pe, found := db.predIndex[pattern.Predicate]
	if !found {
		panic(fmt.Sprintf("Cannot find predicate %s", pattern.Predicate))
	}
	for subject, objectMap := range pe.Subjects {
		for object := range objectMap {
			var sh, oh [4]byte
			copy(sh[:], subject)
			copy(oh[:], object)
			soPair = append(soPair, [][4]byte{sh, oh})
		}
	}
	return soPair
}
