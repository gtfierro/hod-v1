//go:generate msgp
package db

import ()

type PredIndex map[string]*PredicateEntity
type RelshipIndex map[string]string
type NamespaceIndex map[string]string

type Entity struct {
	PK [4]byte `msg:"p"`
	// note: we have to use string keys to get msgp to work
	InEdges  map[string][][4]byte `msg:"ein"`
	OutEdges map[string][][4]byte `msg:"eout"`
}

func NewEntity() *Entity {
	return &Entity{
		InEdges:  make(map[string][][4]byte),
		OutEdges: make(map[string][][4]byte),
	}
}

// returns true if we added an endpoint; false if it was already there
func (e *Entity) AddInEdge(predicate, endpoint [4]byte) bool {
	var (
		edgeList [][4]byte
		found    bool
	)
	// check if we already have an edgelist for the given predicate
	if edgeList, found = e.InEdges[string(predicate[:])]; !found {
		// if we don't, then create a new one and put the endpoint in it
		edgeList = [][4]byte{endpoint}
		e.InEdges[string(predicate[:])] = edgeList
		return true
	}
	// else, we check if our endpoint is already in the edge list
	for _, edge := range edgeList {
		// if it is, return
		if edge == endpoint {
			return false
		}
	}
	// else, we add it into the edge list and return
	edgeList = append(edgeList, endpoint)
	e.InEdges[string(predicate[:])] = edgeList
	return true
}

// returns true if we added an endpoint; false if it was already there
func (e *Entity) AddOutEdge(predicate, endpoint [4]byte) bool {
	var (
		edgeList [][4]byte
		found    bool
	)
	// check if we already have an edgelist for the given predicate
	if edgeList, found = e.OutEdges[string(predicate[:])]; !found {
		// if we don't, then create a new one and put the endpoint in it
		edgeList = [][4]byte{endpoint}
		e.OutEdges[string(predicate[:])] = edgeList
		return true
	}
	// else, we check if our endpoint is already in the edge list
	for _, edge := range edgeList {
		// if it is, return
		if edge == endpoint {
			return false
		}
	}
	// else, we add it into the edge list and return
	edgeList = append(edgeList, endpoint)
	e.OutEdges[string(predicate[:])] = edgeList
	return true
}

type PredicateEntity struct {
	PK [4]byte `msg:"p"`
	// note: we have to use string keys to get msgp to work
	Subjects map[string]map[string]uint32 `msg:"s"`
	Objects  map[string]map[string]uint32 `msg:"o"`
}

func NewPredicateEntity() *PredicateEntity {
	return &PredicateEntity{
		Subjects: make(map[string]map[string]uint32),
		Objects:  make(map[string]map[string]uint32),
	}
}

func (e *PredicateEntity) AddSubjectObject(subject, object [4]byte) {
	// if we have the subject
	if ms, found := e.Subjects[string(subject[:])]; found {
		// find the map of related objects
		ms[string(object[:])] = 0
	} else {
		e.Subjects[string(subject[:])] = map[string]uint32{string(object[:]): 0}
	}

	if ms, found := e.Objects[string(object[:])]; found {
		// find the map of related objects
		ms[string(subject[:])] = 0
	} else {
		e.Objects[string(object[:])] = map[string]uint32{string(subject[:]): 0}
	}
}
