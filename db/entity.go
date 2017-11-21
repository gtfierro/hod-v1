//go:generate msgp
package db

import (
	"encoding/binary"

	"github.com/mitghi/btree"
)

type PredIndex map[string]*PredicateEntity
type RelshipIndex map[string]string
type NamespaceIndex map[string]string

type Entity struct {
	PK Key `msg:"p"`
	// note: we have to use string keys to get msgp to work
	InEdges  map[string][]Key `msg:"i"`
	OutEdges map[string][]Key `msg:"o"`
}

func NewEntity() *Entity {
	return &Entity{
		InEdges:  make(map[string][]Key),
		OutEdges: make(map[string][]Key),
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
	if edgeList, found = e.InEdges[string(predicate[:])]; !found {
		// if we don't, then create a new one and put the endpoint in it
		edgeList = []Key{endpoint}
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
func (e *Entity) AddOutEdge(predicate, endpoint Key) bool {
	var (
		edgeList []Key
		found    bool
	)
	// check if we already have an edgelist for the given predicate
	if edgeList, found = e.OutEdges[string(predicate[:])]; !found {
		// if we don't, then create a new one and put the endpoint in it
		edgeList = []Key{endpoint}
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
	PK Key `msg:"p"`
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

func (e *PredicateEntity) AddSubjectObject(subject, object Key) {
	// if we have the subject
	if _, found := e.Subjects[string(subject[:])]; found {
		// find the map of related objects
		e.Subjects[string(subject[:])][string(object[:])] = 0
	} else {
		e.Subjects[string(subject[:])] = map[string]uint32{string(object[:]): 0}
	}

	if _, found := e.Objects[string(object[:])]; found {
		// find the map of related objects
		e.Objects[string(object[:])][string(subject[:])] = 0
	} else {
		e.Objects[string(object[:])] = map[string]uint32{string(subject[:]): 0}
	}
}

func (e *PredicateEntity) Dump(db *DB) {
	fmt.Printf("dump predicate> %s %p\n", db.MustGetURI(e.PK), e)
	for sub, objmap := range e.Subjects {
		var s, o Key
		s.FromSlice([]byte(sub))
		fmt.Println("   subject>", db.MustGetURI(s))
		for obj := range objmap {
			o.FromSlice([]byte(obj))
			fmt.Println("     object>", db.MustGetURI(o))
		}
	}
}

type EntityExtendedIndex struct {
	PK           Key              `msg:"p"`
	InPlusEdges  map[string][]Key `msg:"i+"`
	OutPlusEdges map[string][]Key `msg:"o+"`
}

func NewEntityExtendedIndex() *EntityExtendedIndex {
	return &EntityExtendedIndex{
		InPlusEdges:  make(map[string][]Key),
		OutPlusEdges: make(map[string][]Key),
	}
}

// returns true if we added an endpoint; false if it was already there
func (e *EntityExtendedIndex) AddOutPlusEdge(predicate, endpoint Key) bool {
	var (
		edgeList []Key
		found    bool
	)
	// check if we already have an edgelist for the given predicate
	if edgeList, found = e.OutPlusEdges[string(predicate[:])]; !found {
		// if we don't, then create a new one and put the endpoint in it
		edgeList = []Key{endpoint}
		e.OutPlusEdges[string(predicate[:])] = edgeList
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
	e.OutPlusEdges[string(predicate[:])] = edgeList
	return true
}

// returns true if we added an endpoint; false if it was already there
func (e *EntityExtendedIndex) AddInPlusEdge(predicate, endpoint Key) bool {
	var (
		edgeList []Key
		found    bool
	)
	// check if we already have an edgelist for the given predicate
	if edgeList, found = e.InPlusEdges[string(predicate[:])]; !found {
		// if we don't, then create a new one and put the endpoint in it
		edgeList = []Key{endpoint}
		e.InPlusEdges[string(predicate[:])] = edgeList
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
	e.InPlusEdges[string(predicate[:])] = edgeList
	return true
}
