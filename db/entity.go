//go:generate msgp
package db

import (
	"encoding/binary"

	"github.com/mitghi/btree"
)

type Entity struct {
	PK Key `msg:"p"`
	// note: we have to use string keys to get msgp to work
	InEdges  map[uint32][]Key `msg:"ein"`
	OutEdges map[uint32][]Key `msg:"eout"`
}

func NewEntity() *Entity {
	return &Entity{
		InEdges:  make(map[uint32][]Key),
		OutEdges: make(map[uint32][]Key),
	}
}

func (e *Entity) Less(than btree.Item, ctx interface{}) bool {
	t := than.(*Entity)
	return binary.LittleEndian.Uint32(e.PK[:]) < binary.LittleEndian.Uint32(t.PK[:])
}

// returns true if we added an endpoint; false if it was already there
func (e *Entity) AddInEdge(predicate, endpoint Key) bool {
	var (
		edgeList []Key
		found    bool
	)
	// check if we already have an edgelist for the given predicate
	if edgeList, found = e.InEdges[predicate.Uint32()]; !found {
		// if we don't, then create a new one and put the endpoint in it
		edgeList = []Key{endpoint}
		e.InEdges[predicate.Uint32()] = edgeList
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
	e.InEdges[predicate.Uint32()] = edgeList
	return true
}

// returns true if we added an endpoint; false if it was already there
func (e *Entity) AddOutEdge(predicate, endpoint Key) bool {
	var (
		edgeList []Key
		found    bool
	)
	// check if we already have an edgelist for the given predicate
	if edgeList, found = e.OutEdges[predicate.Uint32()]; !found {
		// if we don't, then create a new one and put the endpoint in it
		edgeList = []Key{endpoint}
		e.OutEdges[predicate.Uint32()] = edgeList
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
	e.OutEdges[predicate.Uint32()] = edgeList
	return true
}

type PredicateEntity struct {
	PK Key `msg:"p"`
	// note: we have to use uint32 keys to get msgp to work
	Subjects map[uint32]map[uint32]uint32 `msg:"s"`
	Objects  map[uint32]map[uint32]uint32 `msg:"o"`
}

func NewPredicateEntity() *PredicateEntity {
	return &PredicateEntity{
		Subjects: make(map[uint32]map[uint32]uint32),
		Objects:  make(map[uint32]map[uint32]uint32),
	}
}

func (e *PredicateEntity) AddSubjectObject(subject, object Key) {
	// if we have the subject
	subjKey := subject.Uint32()
	objKey := object.Uint32()
	if ms, found := e.Subjects[subjKey]; found {
		// find the map of related objects
		ms[objKey] = 0
	} else {
		e.Subjects[subjKey] = map[uint32]uint32{objKey: 0}
	}

	if ms, found := e.Objects[objKey]; found {
		// find the map of related objects
		ms[subjKey] = 0
	} else {
		e.Objects[objKey] = map[uint32]uint32{subjKey: 0}
	}
}
