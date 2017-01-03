package db

import (
	"container/list"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/google/btree"
	turtle "github.com/gtfierro/hod/goraptor"
	"github.com/gtfierro/hod/query"
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

func (db *DB) getQueryResults(q query.Query) [][]turtle.URI {
	if db.showQueryPlan {
		fmt.Println("-------------- start query plan -------------")
	}
	// start timer
	planStart := time.Now()

	// form dependency graph and build query plan out of it
	dg := db.formDependencyGraph(q)
	qp := db.formQueryPlan(dg)

	if db.showQueryPlan {
		for _, op := range qp.operations {
			log.Notice("op", op)
		}
	}
	if db.showQueryPlanLatencies {
		log.Infof("Formed execution plan in %s", time.Since(planStart))
	}
	if db.showQueryPlan {
		fmt.Println("-------------- end query plan -------------")
	}

	runStart := time.Now()
	rm := db.executeQueryPlan(qp)
	if db.showQueryLatencies {
		log.Infof("Ran query in %s", time.Since(runStart))
	}

	runStart = time.Now()
	results := db.expandTuples(rm, qp.selectVars, q.Select.Partial, q.Select.Limit)
	if db.showQueryLatencies {
		log.Infof("Expanded tuples in %s", time.Since(runStart))
	}
	return results
}

// We need an execution plan for the list of filters contained in a query. How do we do this?
func (db *DB) formDependencyGraph(q query.Query) *dependencyGraph {
	dg := makeDependencyGraph(q)
	terms := make([]*queryTerm, len(q.Where.Filters))
	for i, f := range q.Where.Filters {
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

	originalLength := len(terms)
	for len(terms) > 0 {
		// first find all the terms with 0 or 1 unresolved variable terms
		var added = []*queryTerm{}
		for _, term := range terms {
			if numUnresolved(term) < 2 {
				if len(dg.roots) == 0 {
					dg.addRootTerm(term)
				} else {
					dg.addChild(term)
				}
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
		if len(terms) == originalLength {
			// we don't have any root elements. Need to consider 2-variable terms
			added = []*queryTerm{}
			for _, term := range terms {
				if numUnresolved(term) == 2 {
					dg.addRootTerm(term)
					added = append(added, term)
					break
				}
			}
			terms = filterTermList(terms, added)
		}
	}
	if db.showDependencyGraph {
		dg.dump()
	}
	return dg
}

func (db *DB) executeQueryPlan(qp *queryPlan) *resultMap {
	rm := newResultMap()
	rm.varOrder = qp.varOrder
	var err error
	for _, op := range qp.operations {
		now := time.Now()
		rm, err = op.run(db, qp.varOrder, rm)
		if db.showQueryPlanLatencies {
			fmt.Println(op, time.Since(now))
		}
		if err != nil {
			log.Fatal(err)
		}
	}
	return rm
}

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

	var traversed = traversedBTreePool.Get()
	defer traversedBTreePool.Put(traversed)

	for stack.Len() > 0 {
		entity := stack.Remove(stack.Front()).(*Entity)
		if traversed.Has(Item(entity.PK)) {
			continue
		}
		traversed.ReplaceOrInsert(Item(entity.PK))
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
				results.ReplaceOrInsert(Item(entityHash))
			}
			// because this is one hop, we don't add any new entities to the stack
		case query.PATTERN_ZERO_ONE:
			results.ReplaceOrInsert(Item(entity.PK))
			endpoints, found := entity.InEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range endpoints {
				results.ReplaceOrInsert(Item(entityHash))
			}
			// because this is one hop, we don't add any new entities to the stack
		case query.PATTERN_ZERO_PLUS:
			results.ReplaceOrInsert(Item(entity.PK))
			endpoints, found := entity.InEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range endpoints {
				nextEntity := db.MustGetEntityFromHash(entityHash)
				if !results.Has(Item(nextEntity.PK)) {
					searchstack.PushBack(nextEntity)
				}
				results.ReplaceOrInsert(Item(entityHash))
				stack.PushBack(nextEntity)
			}
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

	var traversed = traversedBTreePool.Get()
	defer traversedBTreePool.Put(traversed)

	for stack.Len() > 0 {
		entity := stack.Remove(stack.Front()).(*Entity)
		if traversed.Has(Item(entity.PK)) {
			continue
		}
		traversed.ReplaceOrInsert(Item(entity.PK))
		switch pattern.Pattern {
		case query.PATTERN_SINGLE:
			// [found] indicates whether or not we have any edges with the given pattern
			endpoints, found := entity.OutEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range endpoints {
				results.ReplaceOrInsert(Item(entityHash))
			}
			// because this is one hop, we don't add any new entities to the stack
		case query.PATTERN_ZERO_ONE:
			// this does not require the pattern to exist, so we add the current entity plus any
			// connected by the appropriate edge
			results.ReplaceOrInsert(Item(entity.PK))
			endpoints, found := entity.OutEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range endpoints {
				results.ReplaceOrInsert(Item(entityHash))
			}
			// because this is one hop, we don't add any new entities to the stack
		case query.PATTERN_ZERO_PLUS:
			results.ReplaceOrInsert(Item(entity.PK))
			endpoints, found := entity.OutEdges[string(predHash[:])]
			// this requires the pattern to exist, so we skip if we have no edges of that name
			if !found {
				continue
			}
			// here, these entities are all connected by the required predicate
			for _, entityHash := range endpoints {
				nextEntity := db.MustGetEntityFromHash(entityHash)
				if !results.Has(Item(nextEntity.PK)) {
					searchstack.PushBack(nextEntity)
				}
				results.ReplaceOrInsert(Item(entityHash))
				stack.PushBack(nextEntity)
			}
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

	stack := list.New()
	stack.PushFront(objEntity)

	var traversed = traversedBTreePool.Get()
	defer traversedBTreePool.Put(traversed)

	for idx, segment := range path {
		reachable := btree.New(2)
		for stack.Len() > 0 {
			entity := stack.Remove(stack.Front()).(*Entity)
			// if we have already traversed this entity, skip it
			if traversed.Has(Item(entity.PK)) {
				continue
			}
			// mark this entity as traversed
			traversed.ReplaceOrInsert(Item(entity.PK))
			db.followPathFromObject(entity, reachable, stack, segment)
		}

		// if we aren't done, then we push these items onto the stack
		if idx < len(path)-1 {
			max := reachable.Max()
			iter := func(i btree.Item) bool {
				ent, err := db.GetEntityFromHash(i.(Item))
				if err != nil {
					log.Error(err)
					return false
				}
				stack.PushBack(ent)
				return i != max
			}
			reachable.Ascend(iter)
		} else {
			return reachable
		}
	}
	return btree.New(2)
}

// Given object and predicate, get all subjects
func (db *DB) getObjectFromSubjectPred(subjectHash [4]byte, path []query.PathPattern) *btree.BTree {
	subEntity, err := db.GetEntityFromHash(subjectHash)
	if err != nil {
		panic(err)
	}

	// stack of entities to search
	stack := list.New()
	stack.PushFront(subEntity)
	var traversed = traversedBTreePool.Get()
	defer traversedBTreePool.Put(traversed)

	// we have our starting entity; follow the first segment of the path and save everything we can reach from there.
	// Then, from that set, search the second segment of the path, etc. We save the last reachable set
	for idx, segment := range path {
		reachable := btree.New(2)
		for stack.Len() > 0 {
			entity := stack.Remove(stack.Front()).(*Entity)
			// if we have already traversed this entity, skip it
			if traversed.Has(Item(entity.PK)) {
				continue
			}
			// mark this entity as traversed
			traversed.ReplaceOrInsert(Item(entity.PK))
			db.followPathFromSubject(entity, reachable, stack, segment)
		}

		// if we aren't done, then we push these items onto the stack
		if idx < len(path)-1 {
			max := reachable.Max()
			iter := func(i btree.Item) bool {
				ent, err := db.GetEntityFromHash(i.(Item))
				if err != nil {
					log.Error(err)
					return false
				}
				stack.PushBack(ent)
				return i != max
			}
			reachable.Ascend(iter)
		} else {
			return reachable
		}
	}
	return btree.New(2)
}

// Given a predicate, it returns pairs of (subject, object) that are connected by that relationship
func (db *DB) getSubjectObjectFromPred(path []query.PathPattern) (soPair [][][4]byte) {
	//pe, found := db.predIndex[pattern.Predicate]
	//if !found {
	//	panic(fmt.Sprintf("Cannot find predicate %s", pattern.Predicate))
	//}
	//for subject, objectMap := range pe.Subjects {
	//	for object := range objectMap {
	//		var sh, oh [4]byte
	//		copy(sh[:], subject)
	//		copy(oh[:], object)
	//		soPair = append(soPair, [][4]byte{sh, oh})
	//	}
	//}
	return soPair
}
