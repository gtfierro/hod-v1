package db

import (
	"fmt"

	sparql "github.com/gtfierro/hod/lang/ast"
	"github.com/gtfierro/hod/storage"
	"github.com/pkg/errors"
	logrus "github.com/sirupsen/logrus"
)

type operation interface {
	run(ctx *queryContext) error
	String() string
	SortKey() string
	GetTerm() queryTerm
}

// ?subject predicate object
// Find all subjects part of triples with the given predicate and object
type resolveSubject struct {
	term queryTerm
}

func (rs *resolveSubject) String() string {
	return fmt.Sprintf("[resolveSubject %s]", rs.term)
}

func (rs *resolveSubject) SortKey() string {
	return rs.term.Subject.String()
}

func (rs *resolveSubject) GetTerm() queryTerm {
	return rs.term
}

func (rs *resolveSubject) run(ctx *queryContext) error {
	// fetch the object from the graph
	object, err := ctx.tx.getHash(rs.term.Object)
	if err != nil && err != storage.ErrNotFound {
		return errors.Wrap(err, fmt.Sprintf("%+v", rs.term))
	} else if err == storage.ErrNotFound {
		return nil
	}
	subjectVar := rs.term.Subject.String()
	// get all subjects reachable from the given object along the path
	subjects, err := ctx.tx.getSubjectFromPredObject(object, rs.term.Predicates)
	if err != nil {
		return err
	}

	if !ctx.defined(subjectVar) {
		// if not defined, then we put this into the relation
		ctx.defineVariable(subjectVar, subjects)
		ctx.rel.add1Value(subjectVar, subjects)
	} else {
		// if it *is* already defined, then we intersect the values by joining
		ctx.unionDefinitions(subjectVar, subjects)

		newrel := NewRelation([]string{subjectVar})
		newrel.add1Value(subjectVar, subjects)

		ctx.rel.join(newrel, []string{subjectVar}, ctx)
	}

	return nil
}

// object predicate ?object
// Find all objects part of triples with the given predicate and subject
type resolveObject struct {
	term queryTerm
}

func (ro *resolveObject) String() string {
	return fmt.Sprintf("[resolveObject %s]", ro.term)
}

func (ro *resolveObject) SortKey() string {
	return ro.term.Object.String()
}

func (ro *resolveObject) GetTerm() queryTerm {
	return ro.term
}

func (ro *resolveObject) run(ctx *queryContext) error {
	// fetch the subject from the graph
	subject, err := ctx.tx.getHash(ro.term.Subject)
	if err != nil && err != storage.ErrNotFound {
		return errors.Wrap(err, fmt.Sprintf("%+v", ro.term))
	} else if err == storage.ErrNotFound {
		return nil
	}
	objectVar := ro.term.Object.String()
	objects := ctx.tx.getObjectFromSubjectPred(subject, ro.term.Predicates)
	if err != nil {
		return err
	}

	if !ctx.defined(objectVar) {
		ctx.defineVariable(objectVar, objects)
		ctx.rel.add1Value(objectVar, objects)
	} else {
		ctx.unionDefinitions(objectVar, objects)

		newrel := NewRelation([]string{objectVar})
		newrel.add1Value(objectVar, objects)

		ctx.rel.join(newrel, []string{objectVar}, ctx)
	}

	return nil
}

// object ?predicate object
// Find all predicates part of triples with the given subject and subject
type resolvePredicate struct {
	term queryTerm
}

func (op *resolvePredicate) String() string {
	return fmt.Sprintf("[resolvePredicate %s]", op.term)
}

func (op *resolvePredicate) SortKey() string {
	return op.term.Predicates[0].Predicate.String()
}

func (op *resolvePredicate) GetTerm() queryTerm {
	return op.term
}

func (op *resolvePredicate) run(ctx *queryContext) error {
	// fetch the subject from the graph
	subject, err := ctx.tx.getEntityByURI(op.term.Subject)
	if err != nil && err != storage.ErrNotFound {
		return errors.Wrap(err, fmt.Sprintf("%+v", op.term))
	} else if err == storage.ErrNotFound {
		return nil
	}
	// now get object
	object, err := ctx.tx.getEntityByURI(op.term.Object)
	if err != nil && err != storage.ErrNotFound {
		return errors.Wrap(err, fmt.Sprintf("%+v", op.term))
	} else if err == storage.ErrNotFound {
		return nil
	}

	predicateVar := op.term.Predicates[0].Predicate.String()
	// get all preds w/ the given end object, starting from the given subject

	predicates := ctx.tx.getPredicateFromSubjectObject(subject, object)

	// new stuff
	if !ctx.defined(predicateVar) {
		ctx.defineVariable(predicateVar, predicates)
		ctx.rel.add1Value(predicateVar, predicates)
	} else {
		ctx.unionDefinitions(predicateVar, predicates)

		newrel := NewRelation([]string{predicateVar})
		newrel.add1Value(predicateVar, predicates)

		ctx.rel.join(newrel, []string{predicateVar}, ctx)
	}

	return nil
}

// ?sub pred ?obj
// Find all subjects and objects that have the given relationship
type restrictSubjectObjectByPredicate struct {
	term                queryTerm
	parentVar, childVar string
}

func (rso *restrictSubjectObjectByPredicate) String() string {
	return fmt.Sprintf("[restrictSubObjByPred %s]", rso.term)
}

func (rso *restrictSubjectObjectByPredicate) SortKey() string {
	return rso.parentVar
}

func (rso *restrictSubjectObjectByPredicate) GetTerm() queryTerm {
	return rso.term
}

func (rso *restrictSubjectObjectByPredicate) run(ctx *queryContext) error {
	var (
		subjectVar = rso.term.Subject.String()
		objectVar  = rso.term.Object.String()
	)

	// this operator takes existing values for subjects and objects and finds the pairs of them that
	// are connected by the path defined by rso.term.Predicates.

	var rsopRelation *Relation
	var relationContents [][]storage.HashKey
	var joinOn []string
	var itererr error

	// use whichever variable has already been joined on, which means
	// that there are values in the relation that we can join with
	if ctx.hasJoined(subjectVar) {
		joinOn = []string{subjectVar}
		subjects := ctx.getValuesForVariable(subjectVar)

		rsopRelation = NewRelation([]string{subjectVar, objectVar})

		subjects.Iter(func(subject storage.HashKey) {
			reachableObjects := ctx.tx.getObjectFromSubjectPred(subject, rso.term.Predicates)
			// we restrict the values in reachableObjects to those that we already have inside 'objectVar'
			ctx.restrictToResolved(objectVar, reachableObjects)
			reachableObjects.Iter(func(objectKey storage.HashKey) {
				relationContents = append(relationContents, []storage.HashKey{subject, objectKey})
			})
		})
		rsopRelation.add2Values(subjectVar, objectVar, relationContents)

	} else if ctx.hasJoined(objectVar) {
		joinOn = []string{objectVar}
		objects := ctx.getValuesForVariable(objectVar)

		rsopRelation = NewRelation([]string{objectVar, subjectVar})

		objects.Iter(func(object storage.HashKey) {
			reachableSubjects, err := ctx.tx.getSubjectFromPredObject(object, rso.term.Predicates)
			if err != nil {
				itererr = err
				return
			}
			ctx.restrictToResolved(subjectVar, reachableSubjects)

			reachableSubjects.Iter(func(subjectKey storage.HashKey) {
				relationContents = append(relationContents, []storage.HashKey{object, subjectKey})
			})
		})
		rsopRelation.add2Values(objectVar, subjectVar, relationContents)
	} else if ctx.cardinalityUnique(subjectVar) < ctx.cardinalityUnique(objectVar) {
		// we start with whichever has fewer values (subject or object). For each of them, we search
		// the graph for reachable endpoints (object or subject) on the provided path (rso.term.Predicates)
		// neither is joined
		joinOn = []string{subjectVar}
		subjects := ctx.getValuesForVariable(subjectVar)

		rsopRelation = NewRelation([]string{subjectVar, objectVar})

		subjects.Iter(func(subject storage.HashKey) {
			reachableObjects := ctx.tx.getObjectFromSubjectPred(subject, rso.term.Predicates)
			ctx.restrictToResolved(objectVar, reachableObjects)

			reachableObjects.Iter(func(objectKey storage.HashKey) {
				relationContents = append(relationContents, []storage.HashKey{subject, objectKey})
			})
		})
		rsopRelation.add2Values(subjectVar, objectVar, relationContents)
	} else {
		joinOn = []string{objectVar}
		objects := ctx.getValuesForVariable(objectVar)

		rsopRelation = NewRelation([]string{objectVar, subjectVar})

		objects.Iter(func(object storage.HashKey) {
			reachableSubjects, err := ctx.tx.getSubjectFromPredObject(object, rso.term.Predicates)
			if err != nil {
				itererr = err
				return
			}
			ctx.restrictToResolved(subjectVar, reachableSubjects)

			reachableSubjects.Iter(func(subjectKey storage.HashKey) {
				relationContents = append(relationContents, []storage.HashKey{object, subjectKey})
			})
		})
		rsopRelation.add2Values(objectVar, subjectVar, relationContents)
	}

	if itererr != nil {
		return itererr
	}

	ctx.rel.join(rsopRelation, joinOn, ctx)
	ctx.markJoined(subjectVar)
	ctx.markJoined(objectVar)

	return nil
}

// ?sub pred ?obj, but we have already resolved the object
// For each of the current
type resolveSubjectFromVarObject struct {
	term queryTerm
}

func (rsv *resolveSubjectFromVarObject) String() string {
	return fmt.Sprintf("[resolveSubFromVarObj %s]", rsv.term)
}

func (rsv *resolveSubjectFromVarObject) SortKey() string {
	return rsv.term.Object.String()
}

func (rsv *resolveSubjectFromVarObject) GetTerm() queryTerm {
	return rsv.term
}

// Use this when we have subject and object variables, but only object has been filled in
func (rsv *resolveSubjectFromVarObject) run(ctx *queryContext) error {
	var (
		objectVar  = rsv.term.Object.String()
		subjectVar = rsv.term.Subject.String()
	)

	var rsopRelation = NewRelation([]string{objectVar, subjectVar})
	var relationContents [][]storage.HashKey
	var itererr error

	newSubjects := newKeymap()

	objects := ctx.getValuesForVariable(objectVar)
	objects.Iter(func(object storage.HashKey) {
		reachableSubjects, err := ctx.tx.getSubjectFromPredObject(object, rsv.term.Predicates)
		if err != nil {
			itererr = err
			return
		}
		ctx.restrictToResolved(subjectVar, reachableSubjects)

		reachableSubjects.Iter(func(subjectKey storage.HashKey) {
			newSubjects.Add(subjectKey)
			relationContents = append(relationContents, []storage.HashKey{object, subjectKey})
		})
	})
	if itererr != nil {
		return itererr
	}

	rsopRelation.add2Values(objectVar, subjectVar, relationContents)
	ctx.rel.join(rsopRelation, rsopRelation.keys[:1], ctx) // join on objectVar
	ctx.markJoined(subjectVar)
	ctx.markJoined(objectVar)
	ctx.unionDefinitions(subjectVar, newSubjects)

	return nil
}

type resolveObjectFromVarSubject struct {
	term queryTerm
}

func (rov *resolveObjectFromVarSubject) String() string {
	return fmt.Sprintf("[resolveObjFromVarSub %s]", rov.term)
}

func (rov *resolveObjectFromVarSubject) SortKey() string {
	return rov.term.Subject.String()
}

func (rov *resolveObjectFromVarSubject) GetTerm() queryTerm {
	return rov.term
}

func (rov *resolveObjectFromVarSubject) run(ctx *queryContext) error {
	var (
		objectVar  = rov.term.Object.String()
		subjectVar = rov.term.Subject.String()
	)

	var rsopRelation = NewRelation([]string{subjectVar, objectVar})
	var relationContents [][]storage.HashKey

	newObjects := newKeymap()

	subjects := ctx.getValuesForVariable(subjectVar)
	subjects.Iter(func(subject storage.HashKey) {
		reachableObjects := ctx.tx.getObjectFromSubjectPred(subject, rov.term.Predicates)
		ctx.restrictToResolved(objectVar, reachableObjects)

		reachableObjects.Iter(func(objectKey storage.HashKey) {
			newObjects.Add(objectKey)
			relationContents = append(relationContents, []storage.HashKey{subject, objectKey})
		})
	})

	rsopRelation.add2Values(subjectVar, objectVar, relationContents)
	ctx.rel.join(rsopRelation, rsopRelation.keys[:1], ctx) // join on subjectVar
	ctx.markJoined(subjectVar)
	ctx.markJoined(objectVar)
	ctx.unionDefinitions(objectVar, newObjects)

	return nil
}

type resolveObjectFromVarSubjectPred struct {
	term queryTerm
}

func (op *resolveObjectFromVarSubjectPred) String() string {
	return fmt.Sprintf("[resolveObjFromVarSubPred %s]", op.term)
}

func (op *resolveObjectFromVarSubjectPred) SortKey() string {
	return op.term.Subject.String()
}

func (op *resolveObjectFromVarSubjectPred) GetTerm() queryTerm {
	return op.term
}

// ?s ?p o
func (op *resolveObjectFromVarSubjectPred) run(ctx *queryContext) error {
	return nil
}

type resolveSubjectObjectFromPred struct {
	term queryTerm
}

func (op *resolveSubjectObjectFromPred) String() string {
	return fmt.Sprintf("[resolveSubObjFromPred %s]", op.term)
}

func (op *resolveSubjectObjectFromPred) SortKey() string {
	return op.term.Subject.String()
}

func (op *resolveSubjectObjectFromPred) GetTerm() queryTerm {
	return op.term
}

func (op *resolveSubjectObjectFromPred) run(ctx *queryContext) error {
	subsobjs, err := ctx.tx.getSubjectObjectFromPred(op.term.Predicates)
	if err != nil {
		return err
	}
	subjectVar := op.term.Subject.String()
	objectVar := op.term.Object.String()

	if ctx.defined(subjectVar) || ctx.hasJoined(subjectVar) {
		rsopRelation := NewRelation([]string{subjectVar, objectVar})
		rsopRelation.add2Values(subjectVar, objectVar, subsobjs)
		ctx.rel.join(rsopRelation, []string{subjectVar}, ctx)
	} else if ctx.defined(objectVar) || ctx.hasJoined(objectVar) {
		rsopRelation := NewRelation([]string{subjectVar, objectVar})
		rsopRelation.add2Values(subjectVar, objectVar, subsobjs)
		ctx.rel.join(rsopRelation, []string{objectVar}, ctx)
	} else {
		ctx.rel.add2Values(subjectVar, objectVar, subsobjs)
	}
	ctx.markJoined(subjectVar)
	ctx.markJoined(objectVar)

	return nil
}

type resolveSubjectPredFromObject struct {
	term queryTerm
}

func (op *resolveSubjectPredFromObject) String() string {
	return fmt.Sprintf("[resolveSubPredFromObj %s]", op.term)
}

func (op *resolveSubjectPredFromObject) SortKey() string {
	return op.term.Predicates[0].Predicate.String()
}

func (op *resolveSubjectPredFromObject) GetTerm() queryTerm {
	return op.term
}

// we have an object and want to find subjects/predicates that connect to it.
// If we have partially resolved the predicate, then we iterate through those connected to
// the known object and then pull the associated subjects. We then filter those subjects
// by anything we've already resolved.
// If we have *not* resolved the predicate, then this is easy: just graph traverse from the object
func (op *resolveSubjectPredFromObject) run(ctx *queryContext) error {
	subjectVar := op.term.Subject.String()
	predicateVar := op.term.Predicates[0].Predicate.String()

	// fetch the object from the graph
	object, err := ctx.tx.getEntityByURI(op.term.Object)
	if err != nil && err != storage.ErrNotFound {
		return errors.Wrap(err, fmt.Sprintf("%+v", op.term))
	} else if err == storage.ErrNotFound {
		return nil
	}

	// get all predicates from it
	predicates := ctx.tx.getPredicatesFromObject(object)

	var subPredPairs [][]storage.HashKey
	var itererr error
	predicates.Iter(func(predicate storage.HashKey) {
		if !ctx.validValue(predicateVar, predicate) {
			return
		}
		pred, err := ctx.tx.getURI(predicate)
		if err != nil {
			itererr = err
			return
		}
		path := []sparql.PathPattern{{Predicate: pred, Pattern: sparql.PATTERN_SINGLE}}
		subjects, err := ctx.tx.getSubjectFromPredObject(object.Key(), path)
		if err != nil {
			itererr = err
			return
		}

		subjects.Iter(func(subject storage.HashKey) {
			if !ctx.validValue(subjectVar, subject) {
				return
			}
			subPredPairs = append(subPredPairs, []storage.HashKey{subject, predicate})

		})
	})
	if itererr != nil {
		return err
	}

	if ctx.defined(subjectVar) {
		rsopRelation := NewRelation([]string{subjectVar, predicateVar})
		rsopRelation.add2Values(subjectVar, predicateVar, subPredPairs)
		ctx.rel.join(rsopRelation, []string{subjectVar}, ctx)
	} else if ctx.defined(predicateVar) {
		rsopRelation := NewRelation([]string{subjectVar, predicateVar})
		rsopRelation.add2Values(subjectVar, predicateVar, subPredPairs)
		ctx.rel.join(rsopRelation, []string{predicateVar}, ctx)
	} else {
		ctx.rel.add2Values(subjectVar, predicateVar, subPredPairs)
	}

	ctx.defineVariable(predicateVar, newKeymap())
	ctx.defineVariable(subjectVar, newKeymap())

	return nil
}

type resolvePredObjectFromSubject struct {
	term queryTerm
}

func (op *resolvePredObjectFromSubject) String() string {
	return fmt.Sprintf("[resolvePredObjectFromSubject %s]", op.term)
}

func (op *resolvePredObjectFromSubject) SortKey() string {
	return op.term.Predicates[0].Predicate.String()
}

func (op *resolvePredObjectFromSubject) GetTerm() queryTerm {
	return op.term
}

func (op *resolvePredObjectFromSubject) run(ctx *queryContext) error {
	objectVar := op.term.Object.String()
	predicateVar := op.term.Predicates[0].Predicate.String()

	// fetch the subject from the graph
	subject, err := ctx.tx.getEntityByURI(op.term.Subject)
	if err != nil && err != storage.ErrNotFound {
		return errors.Wrap(err, fmt.Sprintf("%+v", op.term))
	} else if err == storage.ErrNotFound {
		return nil
	}

	// We take each reachable predicate (from the subject) and enumerate it with each reachable object
	predicates := ctx.tx.getPredicatesFromSubject(subject)
	var predObjPairs [][]storage.HashKey
	var itererr error
	predicates.Iter(func(predicate storage.HashKey) {
		if !ctx.validValue(predicateVar, predicate) {
			return
		}
		pred, err := ctx.tx.getURI(predicate)
		if err != nil {
			itererr = err
			return
		}
		path := []sparql.PathPattern{{Predicate: pred, Pattern: sparql.PATTERN_SINGLE}}
		if err != nil {
			return
		}
		objects := ctx.tx.getObjectFromSubjectPred(subject.Key(), path)

		objects.Iter(func(object storage.HashKey) {
			if !ctx.validValue(objectVar, object) {
				return
			}
			predObjPairs = append(predObjPairs, []storage.HashKey{predicate, object})
		})
	})
	if itererr != nil {
		return itererr
	}

	var rsopRelation *Relation
	var joinOn []string
	if ctx.hasJoined(predicateVar) {
		joinOn = []string{predicateVar}
		rsopRelation = NewRelation([]string{predicateVar, objectVar})
		rsopRelation.add2Values(predicateVar, objectVar, predObjPairs)
		ctx.rel.join(rsopRelation, joinOn, ctx)
	} else if ctx.hasJoined(objectVar) {
		joinOn = []string{objectVar}
		rsopRelation = NewRelation([]string{objectVar, predicateVar})
		rsopRelation.add2Values(predicateVar, objectVar, predObjPairs)
		ctx.rel.join(rsopRelation, joinOn, ctx)
	} else {
		// if nothing has been joined yet, then we are populating this relation for the first time.
		// from that predicate
		ctx.rel.add2Values(predicateVar, objectVar, predObjPairs)
		ctx.defineVariable(predicateVar, newKeymap())
		ctx.defineVariable(objectVar, newKeymap())
	}
	ctx.markJoined(predicateVar)
	ctx.markJoined(objectVar)

	return nil
}

type resolveVarTripleFromSubject struct {
	term queryTerm
}

func (op *resolveVarTripleFromSubject) String() string {
	return fmt.Sprintf("[resolveVarTripleFromSubject %s]", op.term)
}

func (op *resolveVarTripleFromSubject) SortKey() string {
	return op.term.Subject.String()
}

func (op *resolveVarTripleFromSubject) GetTerm() queryTerm {
	return op.term
}

// ?s ?p ?o; start from s
func (op *resolveVarTripleFromSubject) run(ctx *queryContext) error {
	// for all subjects, find all predicates and objects. Note: these predicates
	// and objects may be partially evaluated already
	var (
		subjectVar   = op.term.Subject.String()
		objectVar    = op.term.Object.String()
		predicateVar = op.term.Predicates[0].Predicate.String()
	)

	var rsopRelation = NewRelation([]string{subjectVar, predicateVar, objectVar})
	var relationContents [][]storage.HashKey

	subjects := ctx.definitions[subjectVar]
	subjects.Iter(func(subjectKey storage.HashKey) {
		subject, err := ctx.tx.getEntityByHash(subjectKey)
		if err != nil {
			logrus.Error(err)
			return
		}
		for _, predKey := range subject.GetAllPredicates() {
			for _, objectKey := range subject.ListOutEndpoints(predKey) {
				relationContents = append(relationContents, []storage.HashKey{subject.Key(), predKey, objectKey})
			}
		}
	})

	rsopRelation.add3Values(subjectVar, predicateVar, objectVar, relationContents)
	ctx.rel.join(rsopRelation, []string{subjectVar}, ctx)
	ctx.markJoined(subjectVar)
	ctx.markJoined(predicateVar)
	ctx.markJoined(objectVar)
	return nil
}

type resolveVarTripleFromObject struct {
	term queryTerm
}

func (op *resolveVarTripleFromObject) String() string {
	return fmt.Sprintf("[resolveVarTripleFromObject %s]", op.term)
}

func (op *resolveVarTripleFromObject) SortKey() string {
	return op.term.Object.String()
}

func (op *resolveVarTripleFromObject) GetTerm() queryTerm {
	return op.term
}

// ?s ?p ?o; start from o
func (op *resolveVarTripleFromObject) run(ctx *queryContext) error {
	var (
		subjectVar   = op.term.Subject.String()
		objectVar    = op.term.Object.String()
		predicateVar = op.term.Predicates[0].Predicate.String()
	)

	var rsopRelation = NewRelation([]string{objectVar, predicateVar, subjectVar})
	var relationContents [][]storage.HashKey

	objects := ctx.definitions[objectVar]
	objects.Iter(func(objectKey storage.HashKey) {
		object, err := ctx.tx.getEntityByHash(objectKey)
		if err != nil {
			logrus.Error(err)
			return
		}

		for _, predKey := range object.GetAllPredicates() {
			for _, subjectKey := range object.ListInEndpoints(predKey) {
				relationContents = append(relationContents, []storage.HashKey{object.Key(), predKey, subjectKey})
			}
		}
	})

	rsopRelation.add3Values(objectVar, predicateVar, subjectVar, relationContents)
	ctx.rel.join(rsopRelation, []string{objectVar}, ctx)
	ctx.markJoined(subjectVar)
	ctx.markJoined(predicateVar)
	ctx.markJoined(objectVar)
	return nil
}

type resolveVarTripleFromPredicate struct {
	term queryTerm
}

func (op *resolveVarTripleFromPredicate) String() string {
	return fmt.Sprintf("[resolveVarTripleFromPredicate %s]", op.term)
}

func (op *resolveVarTripleFromPredicate) SortKey() string {
	return op.term.Predicates[0].Predicate.String()
}

func (op *resolveVarTripleFromPredicate) GetTerm() queryTerm {
	return op.term
}

// ?s ?p ?o; start from p
func (op *resolveVarTripleFromPredicate) run(ctx *queryContext) error {
	var (
		subjectVar   = op.term.Subject.String()
		objectVar    = op.term.Object.String()
		predicateVar = op.term.Predicates[0].Predicate.String()
	)

	var rsopRelation = NewRelation([]string{predicateVar, subjectVar, objectVar})
	var relationContents [][]storage.HashKey

	predicates := ctx.definitions[predicateVar]
	var itererr error
	predicates.Iter(func(predicateKey storage.HashKey) {
		var subjectKey storage.HashKey
		// subsobjs := ctx.getSubjectObjectFromPred(rso.term.Predicates)
		uri, err := ctx.tx.getURI(predicateKey)
		if err != nil {
			itererr = err
			return
		}

		predicate, err := ctx.tx.getPredicateByURI(uri)
		if err != nil {
			itererr = err
			return
		}
		//predicate := ctx.db.predIndex[uri]

		for _, subjectHash := range predicate.GetAllSubjects() {
			for _, objectHash := range predicate.GetObjects(subjectHash) {
				relationContents = append(relationContents, []storage.HashKey{predicateKey, subjectKey, objectHash})
			}
		}

	})
	if itererr != nil {
		return itererr
	}

	ctx.markJoined(predicateVar)
	rsopRelation.add3Values(predicateVar, subjectVar, objectVar, relationContents)
	ctx.rel.join(rsopRelation, []string{predicateVar}, ctx)
	ctx.markJoined(subjectVar)
	ctx.markJoined(predicateVar)
	ctx.markJoined(objectVar)
	return nil

}

type resolveVarTripleAll struct {
	term queryTerm
}

func (op *resolveVarTripleAll) String() string {
	return fmt.Sprintf("[resolveVarTripleAll %s]", op.term)
}

func (op *resolveVarTripleAll) SortKey() string {
	return op.term.Subject.String()
}

func (op *resolveVarTripleAll) GetTerm() queryTerm {
	return op.term
}

func (op *resolveVarTripleAll) run(ctx *queryContext) error {
	var (
		subjectVar   = op.term.Subject.String()
		objectVar    = op.term.Object.String()
		predicateVar = op.term.Predicates[0].Predicate.String()
	)
	var relation = NewRelation([]string{subjectVar, predicateVar, objectVar})
	var content [][]storage.HashKey

	iter := func(subjectHash storage.HashKey, entity storage.Entity) bool {
		for _, predHash := range entity.GetAllPredicates() {
			for _, objectHash := range entity.ListOutEndpoints(predHash) {
				content = append(content, []storage.HashKey{subjectHash, predHash, objectHash})
			}
		}
		return false // continue iter
	}
	if err := ctx.tx.iterAllEntities(iter); err != nil {
		return err
	}

	relation.add3Values(subjectVar, predicateVar, objectVar, content)
	if len(ctx.rel.rows) > 0 {
		panic("This should not happen! Tell Gabe")
	}
	// in this case, we just replace the relation
	ctx.rel = relation
	ctx.markJoined(subjectVar)
	ctx.markJoined(predicateVar)
	ctx.markJoined(objectVar)
	return nil
}
