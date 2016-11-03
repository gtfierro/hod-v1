//go:generate msgp
package db

import ()

type Entity struct {
	PK [4]byte `msg:"p"`
	// note: we have to use string keys to get msgp to work
	Edges map[string][][4]byte `msg:"e"`
}

func NewEntity() *Entity {
	return &Entity{
		Edges: make(map[string][][4]byte),
	}
}

func (e *Entity) AddEdge(predicate, endpoint [4]byte) {
	var (
		edgeList [][4]byte
		found    bool
	)
	// check if we already have an edgelist for the given predicate
	if edgeList, found = e.Edges[string(predicate[:])]; !found {
		// if we don't, then create a new one and put the endpoint in it
		edgeList = [][4]byte{endpoint}
		e.Edges[string(predicate[:])] = edgeList
		return
	}
	// else, we check if our endpoint is already in the edge list
	for _, edge := range edgeList {
		// if it is, return
		if edge == endpoint {
			return
		}
	}
	// else, we add it into the edge list and return
	edgeList = append(edgeList, endpoint)
	e.Edges[string(predicate[:])] = edgeList
	return
}

type PredicateEntity struct {
	PK [4]byte `msg:"p"`
	// note: we have to use string keys to get msgp to work
	Subjects [][4]byte `msg:"s"`
	Objects  [][4]byte `msg:"o"`
}

func NewPredicateEntity() *PredicateEntity {
	return &PredicateEntity{
		Subjects: [][4]byte{},
		Objects:  [][4]byte{},
	}
}

func (e *PredicateEntity) AddSubjectObject(subject, object [4]byte) {
	for _, ent := range e.Subjects {
		if ent == subject {
			goto object
		}
	}
	e.Subjects = append(e.Subjects, subject)
object:
	for _, ent := range e.Objects {
		if ent == object {
			return
		}
	}
	e.Objects = append(e.Objects, object)
}
