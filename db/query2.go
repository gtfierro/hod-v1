package db

import (
	"container/list"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/google/btree"
)

type VariableEntity struct {
	PK [4]byte
	// a link has key: variable name, values: set of VariableEntities
	Links map[string]*btree.BTree
}

func (ve VariableEntity) Less(than btree.Item) bool {
	t := than.(*VariableEntity)
	return binary.LittleEndian.Uint32(ve.PK[:]) < binary.LittleEndian.Uint32(t.PK[:])
}

func (db *DB) executeQuery2(run *queryRun) {
	// our query plan defines tree of how different query
	// terms depend on each other. We traverse this in a BFS
	// (breadth-first search) order to resolve the terms
	stack := list.New()
	for _, r := range run.plan.roots {
		stack.PushFront(r)
	}
	for stack.Len() > 0 {
		node := stack.Remove(stack.Front()).(*queryTerm)
		db.runFilterTerm2(run, node)
		// add node children to back of stack
		for _, c := range node.children {
			stack.PushBack(c)
		}
	}
	for variable, res := range run.vars {
		fmt.Printf("var %s has count %d\n", variable, res.Len())
	}
}

func (db *DB) runFilterTerm2(run *queryRun, term *queryTerm) error {
	// first determine if we have any variables in the term. This determines
	// what actions we take
	fmt.Println("run filter", term)
	var (
		subjectIsVariable = strings.HasPrefix(term.Subject.Value, "?")
		objectIsVariable  = strings.HasPrefix(term.Object.Value, "?")
	)
	// This is an odd scenario, that probably involves us having a variable for
	// a predicate, which we don't support yet
	if !subjectIsVariable && !objectIsVariable {
		log.Warningf("THIS IS WEIRD")
		return nil
	} else if !subjectIsVariable {
		// here, object is a variable that is anchored with some predicate to
		// the known subject. We grab the Subject entity from the graph store
		subject, err := db.GetEntity(term.Subject)
		if err != nil {
			return err
		}

		// now we check to see if we already have a set of proposal entities
		// for the 'object' variable. If we do, then we will restrict those
		// by the constraints of this query term. Else, we will grab ALL of the
		// objects from the graph that fit the constraints of this term

		// grab the set of objects that are valid from the subject hash
		// resolve the subject entity to get the hash
		reachableObjects := db.getObjectFromSubjectPred(subject.PK, term.Path)
		// if we have a proposal set already, then we check if we are in
		// the list of reachable Objects. If we aren't, then we ditch that entity
		keepProposals := btree.New(3)
		if proposals, found := run.vars[term.Subject.String()]; found {
			iter := func(i btree.Item) bool {
				entity := i.(*VariableEntity)
				// check if 'entity' is in the result set.
				// If it is, then we keep the proposal; else, we don't
				if reachableObjects.Has(Item(entity.PK)) {
					keepProposals.ReplaceOrInsert(entity)
				}
				return i != proposals.Max()
			}
			proposals.Ascend(iter)
		} else {
			// here, we don't have a set of proposal objects, so we add ALL reachable
			// objects as proposal VariableEntities
			keepProposals = hashTreeToEntityTree(reachableObjects)
		}
		run.vars[term.Object.String()] = keepProposals
	} else if !objectIsVariable {
		// basically copy the logic from above
		object, err := db.GetEntity(term.Object)
		if err != nil {
			return err
		}
		reachableSubjects := db.getSubjectFromPredObject(object.PK, term.Path)
		fmt.Println("reachable subject", reachableSubjects.Len())
		keepProposals := btree.New(3)
		if proposals, found := run.vars[term.Object.String()]; found {
			iter := func(i btree.Item) bool {
				entity := i.(*VariableEntity)
				// check if 'entity' is in the result set.
				// If it is, then we keep the proposal; else, we don't
				if reachableSubjects.Has(Item(entity.PK)) {
					keepProposals.ReplaceOrInsert(entity)
				}
				return i != proposals.Max()
			}
			proposals.Ascend(iter)
		} else {
			// here, we don't have a set of proposal objects, so we add ALL reachable
			// objects as proposal VariableEntities
			keepProposals = hashTreeToEntityTree(reachableSubjects)
			fmt.Println("proposals", keepProposals.Len())
		}
		run.vars[term.Subject.String()] = keepProposals
	} else {
		// if both the subject and object are variables, then there are 4 scenarios:
		// 1: we have results for S but not O (e.g. S was a variable that we already have some results for)
		// 2. we have results for O but not S
		// 3. we have results for BOTH S and O
		// 4. we do NOT have results for either S or O
		// If scenario 4, then the query is not solveable, because if we are at this point,
		// then we should have filled at least one of the variables
		subTree, have_sub := run.vars[term.Subject.String()]
		objTree, have_obj := run.vars[term.Object.String()]
		if have_sub {
			have_sub = subTree.Len() > 0
		}
		if have_obj {
			have_obj = objTree.Len() > 0
		}
		fmt.Println("sub, obj", have_sub, have_obj)
		if have_obj && have_sub {
			// if we have results for both S and O, then we have a scenario something like:
			//   ?room rdf:type brick:Room
			//   ?zone rdf:type brick:HVAC_Zone
			//   ?room bf:hasPart ?zone
			// Here, we have pulled in the types for both of ?room and ?zone, and we need to restrict
			// the result sets by pairs that are connected by hasPart.
			// We loop through the smaller of the two result sets (say S is smaller than O). For each
			// subject in S, we find all objects in O that have the desired relationship and connect those
			// with that subject. When finished, we "delete" O.
			keepSubjects := btree.New(3)
			keepObjects := btree.New(3)
			if subTree.Len() <= objTree.Len() {
				iter := func(i btree.Item) bool {
					subject := i.(*VariableEntity)
					// get all of the objects appropriately connected to our subject
					results := db.getObjectFromSubjectPred(subject.PK, term.Path)
					// terminate early if we have no results. Make sure we don't add this subject to any result set
					if results.Len() == 0 {
						return i != subTree.Max()
					}
					// filter 'results' by the objects we have in objTree
					subject.Links[term.Object.String()] = intersectTrees(objTree, hashTreeToEntityTree(results))
					keepSubjects.ReplaceOrInsert(subject)
					return i != subTree.Max()
				}
				subTree.Ascend(iter)
				run.vars[term.Subject.String()] = keepSubjects
				delete(run.vars, term.Object.String())
			} else {
				iter := func(i btree.Item) bool {
					object := i.(*VariableEntity)
					// get all of the subjects appropriately connected to our object
					results := db.getSubjectFromPredObject(object.PK, term.Path)
					// terminate early if we have no results. Make sure we don't add this object to any result set
					if results.Len() == 0 {
						return i != subTree.Max()
					}
					// filter 'results' by the subjects we have in objTree
					object.Links[term.Subject.String()] = intersectTrees(subTree, hashTreeToEntityTree(results))
					keepObjects.ReplaceOrInsert(object)
					return i != objTree.Max()
				}
				objTree.Ascend(iter)
				run.vars[term.Object.String()] = keepObjects
				delete(run.vars, term.Subject.String())
			}
		} else if have_obj {
			// we have a set of object proposals. For each of them, we find all matching
			// subjects with the requisite path that terminate at the given object, and
			// attach those subjects to that variable
			objectProposals := run.vars[term.Object.String()]
			iter := func(i btree.Item) bool {
				object := i.(*VariableEntity)
				subjects := db.getSubjectFromPredObject(object.PK, term.Path)
				object.Links[term.Subject.String()] = hashTreeToEntityTree(subjects)
				return i != objectProposals.Max()
			}
			objectProposals.Ascend(iter)
		} else if have_sub {
			subjectProposals := run.vars[term.Subject.String()]
			iter := func(i btree.Item) bool {
				subject := i.(*VariableEntity)
				objects := db.getObjectFromSubjectPred(subject.PK, term.Path)
				subject.Links[term.Object.String()] = hashTreeToEntityTree(objects)
				return i != subjectProposals.Max()
			}
			subjectProposals.Ascend(iter)
		}
	}
	return nil
}
