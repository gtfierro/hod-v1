// this file contains the set of query operators generated by the query planner
package db

import (
	"fmt"

	"github.com/gtfierro/hod/query"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

type operation interface {
	run(ctx *queryContext) error
	String() string
	SortKey() string
	GetTerm() *queryTerm
}

// ?subject predicate object
// Find all subjects part of triples with the given predicate and object
type resolveSubject struct {
	term *queryTerm
}

func (rs *resolveSubject) String() string {
	return fmt.Sprintf("[resolveSubject %s]", rs.term)
}

func (rs *resolveSubject) SortKey() string {
	return rs.term.Subject.String()
}

func (rs *resolveSubject) GetTerm() *queryTerm {
	return rs.term
}

func (rs *resolveSubject) run(ctx *queryContext) error {
	// fetch the object from the graph
	object, err := ctx.db.GetEntity(rs.term.Object)
	if err != nil && err != leveldb.ErrNotFound {
		return errors.Wrap(err, fmt.Sprintf("%+v", rs.term))
	} else if err == leveldb.ErrNotFound {
		return nil
	}
	subjectVar := rs.term.Subject.String()
	// get all subjects reachable from the given object along the path
	subjects := ctx.db.getSubjectFromPredObject(object.PK, rs.term.Path)

	ctx.addOrFilterVariable(subjectVar, hashTreeToPointerTree(ctx.db, subjects))

	return nil
}

// object predicate ?object
// Find all objects part of triples with the given predicate and subject
type resolveObject struct {
	term *queryTerm
}

func (ro *resolveObject) String() string {
	return fmt.Sprintf("[resolveObject %s]", ro.term)
}

func (ro *resolveObject) SortKey() string {
	return ro.term.Object.String()
}

func (ro *resolveObject) GetTerm() *queryTerm {
	return ro.term
}

func (ro *resolveObject) run(ctx *queryContext) error {
	// fetch the subject from the graph
	subject, err := ctx.db.GetEntity(ro.term.Subject)
	if err != nil && err != leveldb.ErrNotFound {
		return errors.Wrap(err, fmt.Sprintf("%+v", ro.term))
	} else if err == leveldb.ErrNotFound {
		return nil
	}
	objectVar := ro.term.Object.String()
	// get all objects reachable from the given subject along the path
	objects := ctx.db.getObjectFromSubjectPred(subject.PK, ro.term.Path)

	ctx.addOrFilterVariable(objectVar, hashTreeToPointerTree(ctx.db, objects))

	return nil
}

// object ?predicate object
// Find all predicates part of triples with the given subject and subject
type resolvePredicate struct {
	term *queryTerm
}

func (op *resolvePredicate) String() string {
	return fmt.Sprintf("[resolvePredicate %s]", op.term)
}

func (op *resolvePredicate) SortKey() string {
	return op.term.Path[0].Predicate.String()
}

func (op *resolvePredicate) GetTerm() *queryTerm {
	return op.term
}

func (op *resolvePredicate) run(ctx *queryContext) error {
	// fetch the subject from the graph
	subject, err := ctx.db.GetEntity(op.term.Subject)
	if err != nil && err != leveldb.ErrNotFound {
		return errors.Wrap(err, fmt.Sprintf("%+v", op.term))
	} else if err == leveldb.ErrNotFound {
		return nil
	}
	// now get object
	object, err := ctx.db.GetEntity(op.term.Object)
	if err != nil && err != leveldb.ErrNotFound {
		return errors.Wrap(err, fmt.Sprintf("%+v", op.term))
	} else if err == leveldb.ErrNotFound {
		return nil
	}

	predicateVar := op.term.Path[0].Predicate.String()
	// get all preds w/ the given end object, starting from the given subject

	predicates := ctx.db.getPredicateFromSubjectObject(subject, object)

	ctx.addOrFilterVariable(predicateVar, hashTreeToPointerTree(ctx.db, predicates))
	return nil
}

// ?sub pred ?obj
// Find all subjects and objects that have the given relationship
type restrictSubjectObjectByPredicate struct {
	term                *queryTerm
	parentVar, childVar string
}

func (rso *restrictSubjectObjectByPredicate) String() string {
	return fmt.Sprintf("[restrictSubObjByPred %s]", rso.term)
}

func (rso *restrictSubjectObjectByPredicate) SortKey() string {
	return rso.parentVar
}

func (rso *restrictSubjectObjectByPredicate) GetTerm() *queryTerm {
	return rso.term
}

// this forms a linking between the subject and object vars; for each
// subject, we want to have the set of objects that 'follow' from it.
// A variable can be in various states:
//  - unresolved (we don't know what the variable is)
//  - resolved, unconnected (we have proposal values for the variable, but they aren't
//      associated with any other variable)
//  - resolved, connected (we have proposal values for the variable, and they are linked
//      to another variable)

func (rso *restrictSubjectObjectByPredicate) run(ctx *queryContext) error {
	var (
		subjectVar = rso.term.Subject.String()
		objectVar  = rso.term.Object.String()
		subTree, _ = ctx.getValues(subjectVar)
		objTree, _ = ctx.getValues(objectVar)
	)
	// we add the objects on to each subject
	if rso.parentVar == subjectVar {
		// iterate through current subjects
		max := subTree.Max()
		iter := func(subject *Entity) bool {
			objects := hashTreeToPointerTree(ctx.db, ctx.db.getObjectFromSubjectPred(subject.PK, rso.term.Path))
			ctx.addOrMergeVariable(objectVar, objects)
			// now add the links. From subject var, the links are the object results
			ctx.addReachable(subject, subjectVar, objects, objectVar)
			return subject != max
		}
		subTree.Iter(iter)

	} else if rso.parentVar == objectVar {
		// iterate through current objects
		max := objTree.Max()
		iter := func(object *Entity) bool {
			subjects := hashTreeToPointerTree(ctx.db, ctx.db.getSubjectFromPredObject(object.PK, rso.term.Path))
			ctx.addOrMergeVariable(subjectVar, subjects)
			ctx.addReachable(object, objectVar, subjects, subjectVar)
			return object != max
		}
		objTree.Iter(iter)
	} else {
		log.Fatal("unfamiliar situation")
	}

	return nil
}

// ?sub pred ?obj, but we have already resolved the object
// For each of the current
type resolveSubjectFromVarObject struct {
	term *queryTerm
}

func (rsv *resolveSubjectFromVarObject) String() string {
	return fmt.Sprintf("[resolveSubFromVarObj %s]", rsv.term)
}

func (rsv *resolveSubjectFromVarObject) SortKey() string {
	return rsv.term.Object.String()
}

func (rsv *resolveSubjectFromVarObject) GetTerm() *queryTerm {
	return rsv.term
}

// Use this when we have subject and object variables, but only object has been filled in
func (rsv *resolveSubjectFromVarObject) run(ctx *queryContext) error {
	var (
		objectVar  = rsv.term.Object.String()
		subjectVar = rsv.term.Subject.String()
		objTree, _ = ctx.getValues(objectVar)
	)
	max := objTree.Max()
	iter := func(object *Entity) bool {
		subjects := hashTreeToPointerTree(ctx.db, ctx.db.getSubjectFromPredObject(object.PK, rsv.term.Path))
		ctx.addOrMergeVariable(subjectVar, subjects)
		ctx.addReachable(object, objectVar, subjects, subjectVar)
		return object != max
	}
	objTree.Iter(iter)
	return nil
}

type resolveObjectFromVarSubject struct {
	term *queryTerm
}

func (rov *resolveObjectFromVarSubject) String() string {
	return fmt.Sprintf("[resolveObjFromVarSub %s]", rov.term)
}

func (rov *resolveObjectFromVarSubject) SortKey() string {
	return rov.term.Subject.String()
}

func (rov *resolveObjectFromVarSubject) GetTerm() *queryTerm {
	return rov.term
}

func (rov *resolveObjectFromVarSubject) run(ctx *queryContext) error {
	var (
		objectVar  = rov.term.Object.String()
		subjectVar = rov.term.Subject.String()
		subTree, _ = ctx.getValues(subjectVar)
	)
	max := subTree.Max()
	iter := func(subject *Entity) bool {
		objects := hashTreeToPointerTree(ctx.db, ctx.db.getObjectFromSubjectPred(subject.PK, rov.term.Path))
		ctx.addOrMergeVariable(objectVar, objects)
		ctx.addReachable(subject, subjectVar, objects, objectVar)
		return subject != max
	}
	subTree.Iter(iter)
	return nil
}

type resolveObjectFromVarSubjectPred struct {
	term *queryTerm
}

func (op *resolveObjectFromVarSubjectPred) String() string {
	return fmt.Sprintf("[resolveObjFromVarSubPred %s]", op.term)
}

func (op *resolveObjectFromVarSubjectPred) SortKey() string {
	return op.term.Subject.String()
}

func (op *resolveObjectFromVarSubjectPred) GetTerm() *queryTerm {
	return op.term
}

// TODO: implement resolveObjectFromVarSubjectPred
// ?s ?p o
func (rov *resolveObjectFromVarSubjectPred) run(ctx *queryContext) error {
	return nil
}

type resolveSubjectObjectFromPred struct {
	term *queryTerm
}

func (rso *resolveSubjectObjectFromPred) run(ctx *queryContext) error {
	subsobjs := ctx.db.getSubjectObjectFromPred(rso.term.Path)
	subjectVar := rso.term.Subject.String()
	objectVar := rso.term.Object.String()
	subjects := newPointerTree(3)
	objects := newPointerTree(3)
	for _, sopair := range subsobjs {
		subjects.Add(ctx.db.MustGetEntityFromHash(sopair[0]))
		objects.Add(ctx.db.MustGetEntityFromHash(sopair[1]))
	}
	ctx.addOrFilterVariable(subjectVar, subjects)
	ctx.addOrFilterVariable(objectVar, objects)
	return nil
}

type resolveSubjectPredFromObject struct {
	term *queryTerm
}

func (op *resolveSubjectPredFromObject) String() string {
	return fmt.Sprintf("[resolveSubPredFromObj %s]", op.term)
}

func (op *resolveSubjectPredFromObject) SortKey() string {
	return op.term.Path[0].Predicate.String()
}

func (op *resolveSubjectPredFromObject) GetTerm() *queryTerm {
	return op.term
}

// we have an object and want to find subjects/predicates that connect to it.
// If we have partially resolved the predicate, then we iterate through those connected to
// the known object and then pull the associated subjects. We then filter those subjects
// by anything we've already resolved.
// If we have *not* resolved the predicate, then this is easy: just graph traverse from the object
func (op *resolveSubjectPredFromObject) run(ctx *queryContext) error {
	var (
		tree *pointerTree
	)
	subjectVar := op.term.Subject.String()
	predicateVar := op.term.Path[0].Predicate.String()

	// fetch the object from the graph
	object, err := ctx.db.GetEntity(op.term.Object)
	if err != nil && err != leveldb.ErrNotFound {
		return errors.Wrap(err, fmt.Sprintf("%+v", op.term))
	} else if err == leveldb.ErrNotFound {
		return nil
	}
	candidateSubjects := newPointerTree(2)
	// get all predicates from it
	predicates := hashTreeToPointerTree(ctx.db, ctx.db.getPredicatesFromObject(object))
	// for each subject reachable from each predicate, add the predicate as aa
	// dependent of the subject
	predmax := predicates.Max()
	iterpred := func(predicate *Entity) bool {
		path := []query.PathPattern{{Predicate: ctx.db.MustGetURI(predicate.PK), Pattern: query.PATTERN_SINGLE}}
		subjects := hashTreeToPointerTree(ctx.db, ctx.db.getSubjectFromPredObject(object.PK, path))
		max := subjects.Max()
		iter := func(ent *Entity) bool {
			tree = ctx.getLinkedValues(ent)
			tree.Add(predicate)
			candidateSubjects.Add(ent) // subject
			ctx.addReachable(ent, subjectVar, tree, predicateVar)
			return ent != max
		}
		subjects.Iter(iter)
		return predicate != predmax
	}
	predicates.Iter(iterpred)

	// need to merge w/ the subjects we've already gotten
	ctx.addOrFilterVariable(subjectVar, candidateSubjects)

	return nil
}

type resolvePredObjectFromSubject struct {
	term *queryTerm
}

func (op *resolvePredObjectFromSubject) String() string {
	return fmt.Sprintf("[resolvePredObjectFromSubject %s]", op.term)
}

func (op *resolvePredObjectFromSubject) SortKey() string {
	return op.term.Path[0].Predicate.String()
}

func (op *resolvePredObjectFromSubject) GetTerm() *queryTerm {
	return op.term
}

func (op *resolvePredObjectFromSubject) run(ctx *queryContext) error {
	var (
		tree *pointerTree
	)
	objectVar := op.term.Object.String()
	predicateVar := op.term.Path[0].Predicate.String()

	// fetch the subject from the graph
	subject, err := ctx.db.GetEntity(op.term.Subject)
	if err != nil && err != leveldb.ErrNotFound {
		return errors.Wrap(err, fmt.Sprintf("%+v", op.term))
	} else if err == leveldb.ErrNotFound {
		return nil
	}
	candidateObjects := newPointerTree(2)
	// get all predicates from it
	predicates := hashTreeToPointerTree(ctx.db, ctx.db.getPredicatesFromSubject(subject))
	predmax := predicates.Max()
	iterpred := func(predicate *Entity) bool {
		path := []query.PathPattern{{Predicate: ctx.db.MustGetURI(predicate.PK), Pattern: query.PATTERN_SINGLE}}
		objects := hashTreeToPointerTree(ctx.db, ctx.db.getObjectFromSubjectPred(subject.PK, path))
		max := objects.Max()
		iter := func(ent *Entity) bool {
			tree = ctx.getLinkedValues(ent)
			tree.Add(predicate)
			candidateObjects.Add(ent) // object
			ctx.addReachable(ent, objectVar, tree, predicateVar)
			return ent != max
		}
		objects.Iter(iter)
		return predicate != predmax
	}
	predicates.Iter(iterpred)

	// need to merge w/ the objects we've already gotten
	ctx.addOrFilterVariable(objectVar, candidateObjects)

	return nil
}

// TODO: implement these for ?s ?p ?o constructs
// TODO: also requires query planner
type resolveVarTripleFromSubject struct {
	term *queryTerm
}

func (op *resolveVarTripleFromSubject) String() string {
	return fmt.Sprintf("[resolveVarTripleFromSubject %s]", op.term)
}

func (op *resolveVarTripleFromSubject) SortKey() string {
	return op.term.Subject.String()
}

func (op *resolveVarTripleFromSubject) GetTerm() *queryTerm {
	return op.term
}

// ?s ?p ?o; start from s
func (op *resolveVarTripleFromSubject) run(ctx *queryContext) error {
	// for all subjects, find all predicates and objects. Note: these predicates
	// and objects may be partially evaluated already
	var (
		subjectVar                     = op.term.Subject.String()
		objectVar                      = op.term.Object.String()
		predicateVar                   = op.term.Path[0].Predicate.String()
		subjects, _                    = ctx.getValues(subjectVar)
		knownPredicates, hadPredicates = ctx.getValues(predicateVar)
		candidateObjects               = newPointerTree(2)
		candidatePredicates            = newPointerTree(2)
	)

	maxSub := subjects.Max()
	var predKey Key
	subjectIter := func(subject *Entity) bool {
		linkedPredicates := newPointerTree(2)
		for edge, objectList := range subject.OutEdges {
			predKey.FromSlice([]byte(edge))
			predicate := ctx.db.MustGetEntityFromHash(predKey)
			if hadPredicates && !knownPredicates.Has(predicate) {
				continue // skip
			}
			candidatePredicates.Add(predicate)
			linkedPredicates.Add(predicate)
			linkedObjects := newPointerTree(2)
			for _, objectKey := range objectList {
				object := ctx.db.MustGetEntityFromHash(objectKey)
				candidateObjects.Add(object)
				linkedObjects.Add(object)
			}
			ctx.addReachable(predicate, predicateVar, linkedObjects, objectVar)
		}
		ctx.addReachable(subject, subjectVar, linkedPredicates, predicateVar)
		return subject != maxSub
	}
	subjects.Iter(subjectIter)
	ctx.addOrMergeVariable(objectVar, candidateObjects)
	ctx.addOrMergeVariable(predicateVar, candidatePredicates)
	return nil
}

type resolveVarTripleFromObject struct {
	term *queryTerm
}

func (op *resolveVarTripleFromObject) String() string {
	return fmt.Sprintf("[resolveVarTripleFromObject %s]", op.term)
}

func (op *resolveVarTripleFromObject) SortKey() string {
	return op.term.Object.String()
}

func (op *resolveVarTripleFromObject) GetTerm() *queryTerm {
	return op.term
}

// ?s ?p ?o; start from o
func (op *resolveVarTripleFromObject) run(ctx *queryContext) error {
	var (
		subjectVar                     = op.term.Subject.String()
		objectVar                      = op.term.Object.String()
		predicateVar                   = op.term.Path[0].Predicate.String()
		objects, _                     = ctx.getValues(objectVar)
		knownPredicates, hadPredicates = ctx.getValues(predicateVar)
		candidateSubjects              = newPointerTree(2)
		candidatePredicates            = newPointerTree(2)
	)

	maxObj := objects.Max()
	var predKey Key
	objectIter := func(object *Entity) bool {
		linkedPredicates := newPointerTree(2)
		for edge, subjectList := range object.InEdges {
			predKey.FromSlice([]byte(edge))
			predicate := ctx.db.MustGetEntityFromHash(predKey)
			if hadPredicates && !knownPredicates.Has(predicate) {
				continue // skip
			}
			candidatePredicates.Add(predicate)
			linkedPredicates.Add(predicate)
			linkedSubjects := newPointerTree(2)
			for _, subjectKey := range subjectList {
				subject := ctx.db.MustGetEntityFromHash(subjectKey)
				candidateSubjects.Add(subject)
				linkedSubjects.Add(subject)
			}
			ctx.addReachable(predicate, predicateVar, linkedSubjects, subjectVar)
		}
		ctx.addReachable(object, objectVar, linkedPredicates, predicateVar)
		return object != maxObj
	}
	objects.Iter(objectIter)
	ctx.addOrMergeVariable(subjectVar, candidateSubjects)
	ctx.addOrMergeVariable(predicateVar, candidatePredicates)
	return nil
}

type resolveVarTripleFromPredicate struct {
	term *queryTerm
}

func (op *resolveVarTripleFromPredicate) String() string {
	return fmt.Sprintf("[resolveVarTripleFromPredicate %s]", op.term)
}

func (op *resolveVarTripleFromPredicate) SortKey() string {
	return op.term.Path[0].Predicate.String()
}

func (op *resolveVarTripleFromPredicate) GetTerm() *queryTerm {
	return op.term
}

// ?s ?p ?o; start from s
func (op *resolveVarTripleFromPredicate) run(ctx *queryContext) error {
	return nil
}

type resolveVarTripleAll struct {
	term *queryTerm
}

func (op *resolveVarTripleAll) String() string {
	return fmt.Sprintf("[resolveVarTripleAll %s]", op.term)
}

func (op *resolveVarTripleAll) SortKey() string {
	return op.term.Subject.String()
}

func (op *resolveVarTripleAll) GetTerm() *queryTerm {
	return op.term
}

// ?s ?p ?o; start from s
func (op *resolveVarTripleAll) run(ctx *queryContext) error {
	return nil
}
