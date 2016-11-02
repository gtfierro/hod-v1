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
