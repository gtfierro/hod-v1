package db

import (
	"container/list"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/google/btree"
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
	plan      *queryPlan
	variables map[string]*btree.BTree
}

func makeQueryRun(plan *queryPlan) *queryRun {
	qr := &queryRun{
		plan:      plan,
		variables: make(map[string]*btree.BTree),
	}
	for _, v := range plan.selectVars {
		qr.variables[v] = btree.New(3)
	}
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
	qp := db.formExecutionPlan(q)
	log.Infof("Formed execution plan in %s", time.Since(planStart))
	fmt.Println("-------------- end query plan -------------")

	runStart := time.Now()
	run := makeQueryRun(qp)
	db.executeQuery(run)
	log.Infof("Ran query in %s", time.Since(runStart))

	for _, varName := range q.Select.Variables {
		resultTree := run.variables[varName.String()]
		if q.Select.Count {
			fmt.Println(varName, resultTree.Len())
		} else {
			iter := func(i btree.Item) bool {
				uri := db.MustGetURI(i.(Item))
				fmt.Println(varName, uri.String())
				return i != resultTree.Max()
			}
			resultTree.Ascend(iter)
		}
	}
}

// We need an execution plan for the list of filters contained in a query. How do we do this?
func (db *DB) formExecutionPlan(q query.Query) *queryPlan {
	qp := makeQueryPlan(q)
	terms := make([]*queryTerm, len(q.Where))
	for i, f := range q.Where {
		terms[i] = qp.makeQueryTerm(f)
	}

	for len(terms) > 0 {
		// first find all the terms with 0 or 1 unresolved variable terms
		var added = []*queryTerm{}
		for _, term := range terms {
			if term.numUnresolved() < 2 {
				qp.addRootTerm(term)
				added = append(added, term)
			}
		}
		// remove the terms that we added to the root set
		terms = filterTermList(terms, added)
		added = []*queryTerm{}
		for _, term := range terms {
			if qp.addChild(term) {
				added = append(added, term)
			}
		}
		terms = filterTermList(terms, added)
	}
	qp.dump()
	return qp
}

// okay how do we run the execution plan?
// There's actually some ambiguity here: originally, the plan was to throw all of the matched
// variables into Btrees, and then recover what the returned tuples are; however, this isn't straightforward
// because you only want to create the tuples that were found as a result of the query, which is going to be
// a subset of the full connectivity between tuples in the graph. So, what's the approach?
// Proposal 1: first run the query to get 'pools' of valid entities, then 're-run' the query, restricting results
//             by the sets of entities that exist in the pools
// Proposal 2: when we execute the query by following the query plan, rather than running on the full graph, we
//             make sure to associate our result sets with the chain of terms and variables and entities we have
//             traversed so far. The challenge here is how do we do this kind of associated store.
// I like proposal 2 better, because it re-does less work. So, how do we do it?
// When we resolve a variable (lets start with one of the root terms), we get a set of 'proposal' entities. Right now
// we just throw these into a big tree and treat it as a 'set'. Rather, instead of just storing the entity, we need to store
// a structure that has the entity along with the set of paths of relationships to other variables that come from that entity.
// The end result is we get a list of tuples of all variables in the query, and then we can take the subset of variables
// mentioned in the select clause and uniquify the results
func (db *DB) executeQuery(run *queryRun) {
	// first, resolve all the roots and store the intermediate results

	stack := list.New()
	for _, r := range run.plan.roots {
		stack.PushFront(r)
	}
	for stack.Len() > 0 {
		node := stack.Remove(stack.Front()).(*queryTerm)
		db.runFilterTerm(run, node)
		// add node children to back of stack
		for _, c := range node.children {
			stack.PushBack(c)
		}
	}
	for variable, res := range run.variables {
		fmt.Printf("var %s has count %d\n", variable, res.Len())
	}

}

func (db *DB) runFilterTerm(run *queryRun, term *queryTerm) error {
	var (
		subjectIsVariable = strings.HasPrefix(term.Subject.Value, "?")
		objectIsVariable  = strings.HasPrefix(term.Object.Value, "?")
	)
	if !subjectIsVariable && !objectIsVariable {
		log.Warningf("THIS IS WEIRD")
		return nil
		//log.Noticef("S/O anchored: S: %s, O: %s", term.Subject.String(), term.Object.String())
		//results := db.getSubjectObjectFromPred(term.Path[0])
		//log.Infof("Got %d results", len(results))
	} else if !subjectIsVariable {
		log.Noticef("S anchored: S: %s, O: %s", term.Subject.String(), term.Object.String())
		entity, err := db.GetEntity(term.Subject)
		if err != nil {
			return err
		}
		results := db.getObjectFromSubjectPred(entity.PK, term.Path)
		if tree, found := run.variables[term.Object.String()]; found {
			mergeTrees(tree, results)
		} else {
			tree := btree.New(3)
			mergeTrees(tree, results)
			run.variables[term.Object.String()] = tree
		}
	} else if !objectIsVariable {
		log.Noticef("O anchored: S: %s, O: %s", term.Subject.String(), term.Object.String())
		entity, err := db.GetEntity(term.Object)
		if err != nil {
			return err
		}

		results := db.getSubjectFromPredObject(entity.PK, term.Path)
		if tree, found := run.variables[term.Subject.String()]; found {
			mergeTrees(tree, results)
		} else {
			tree := btree.New(3)
			mergeTrees(tree, results)
			run.variables[term.Subject.String()] = tree
		}
	} else {
		// if both the subject and object are variables, then there are 4 scenarios:
		// 1: we have results for S but not O (e.g. S was a variable that we already have some results for)
		// 2. we have results for O but not S
		// 3. we have results for BOTH S and O
		// 4. we do NOT have results for either S or O
		// If scenario 4, then the query is not solveable, because if we are at this point,
		// then we should have filled at least one of the variables
		subTree, have_sub := run.variables[term.Subject.String()]
		objTree, have_obj := run.variables[term.Object.String()]
		if have_sub {
			have_sub = subTree.Len() > 0
		}
		if have_obj {
			have_obj = objTree.Len() > 0
		}
		log.Debug("have s?", have_sub, "have o?", have_obj)
		if have_sub && have_obj {
			// what do we do here? We restrict the sets to those pairs of subject/object that are connected
			// by the provided predicate path
			// How do we do this? We iterate through the SHORTER of the two variable trees. For example, lets say
			// this is the subject variable tree
			// for sub in subjectTree:
			//   foundObjects = sub.findPath(path)
			//   if len(foundObjects) > 0:
			//      keepSubs.appenD(sub)
			//      keepObjects.append(foundobjects...)
			//TODO: DOES THIS EVEN WORK. Check counts before/after
			keepSubjects := btree.New(3)
			keepObjects := btree.New(3)
			log.Warningf("subject len %d, object len %d", subTree.Len(), objTree.Len())
			if subTree.Len() <= objTree.Len() {
				iter := func(i btree.Item) bool {
					subject, err := db.GetEntityFromHash(i.(Item))
					if err != nil {
						log.Error(err)
					}
					results := db.getObjectFromSubjectPred(subject.PK, term.Path)
					if results.Len() > 0 {
						keepSubjects.ReplaceOrInsert(i)
						mergeTrees(keepObjects, results)
					}
					return i != subTree.Max()
				}
				subTree.Ascend(iter)
			} else {
				iter := func(i btree.Item) bool {
					object, err := db.GetEntityFromHash(i.(Item))
					if err != nil {
						log.Error(err)
					}
					results := db.getSubjectFromPredObject(object.PK, term.Path)
					if results.Len() > 0 {
						keepObjects.ReplaceOrInsert(i)
						mergeTrees(keepSubjects, results)
					}
					return i != objTree.Max()
				}
				objTree.Ascend(iter)
			}
			log.Warningf("subject len %d, object len %d", keepSubjects.Len(), keepObjects.Len())
			//TODO: the other way, traversing object tree
			run.variables[term.Subject.String()] = keepSubjects
			run.variables[term.Object.String()] = keepObjects
		} else if have_obj {
			// in this scenario, we have a set of object entities, and we want to find the set of subject entities
			// that map to them using the given path
			subTree = btree.New(3)
			iter := func(i btree.Item) bool {
				object, err := db.GetEntityFromHash(i.(Item))
				if err != nil {
					log.Error(err)
				}
				results := db.getSubjectFromPredObject(object.PK, term.Path)
				mergeTrees(subTree, results)
				return i != objTree.Max()
			}
			objTree.Ascend(iter)
			run.variables[term.Subject.String()] = subTree
		} else if have_sub {
			objTree = btree.New(3)
			iter := func(i btree.Item) bool {
				subject, err := db.GetEntityFromHash(i.(Item))
				if err != nil {
					log.Error(err)
				}
				results := db.getObjectFromSubjectPred(subject.PK, term.Path)
				mergeTrees(objTree, results)
				return i != subTree.Max()
			}
			subTree.Ascend(iter)
			run.variables[term.Subject.String()] = objTree
		} else {
			log.Warning("WHY ARE WE HERE")
		}
		log.Noticef("not anchored!: S: %s, O: %s", term.Subject.String(), term.Object.String())
	}
	return nil
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
//func (db *DB) getSubjectFromPredObject(objectHash [4]byte, pattern query.PathPattern) [][4]byte {
//	// get the object, look in its "in" edges for the path pattern
//	objEntity, err := db.GetEntityFromHash(objectHash)
//	if err != nil {
//		panic(err)
//	}
//	// get predicate hash
//	predHash, err := db.GetHash(pattern.Predicate)
//	if err != nil {
//		panic(err)
//	}
//	return objEntity.InEdges[string(predHash[:])]
//}

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
