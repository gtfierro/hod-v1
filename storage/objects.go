//go:generate msgp
package storage

//TODO: switch to iterator-type interface rather than creating lists

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/gtfierro/btree"
)

type Version struct {
	Timestamp uint64
	Name      string
}

func (v Version) String() string {
	return fmt.Sprintf("<Version %s.%d>", v.Name, v.Timestamp)
}

type HashKey [8]byte

var EmptyKey = HashKey{}

type KeyType [4]byte

var (
	PK        KeyType = [4]byte{0, 0, 0, 0}
	URI       KeyType = [4]byte{0, 0, 0, 1}
	ENTITY    KeyType = [4]byte{0, 0, 0, 2}
	EXTENDED  KeyType = [4]byte{0, 0, 0, 3}
	PREDICATE KeyType = [4]byte{0, 0, 0, 4}
)

func (hk HashKey) Type() KeyType {
	switch {
	case bytes.Equal(hk[:], PK[:]):
		return PK
	case bytes.Equal(hk[:], URI[:]):
		return URI
	case bytes.Equal(hk[:], ENTITY[:]):
		return ENTITY
	case bytes.Equal(hk[:], EXTENDED[:]):
		return EXTENDED
	case bytes.Equal(hk[:], PREDICATE[:]):
		return PREDICATE
	}
	return PK
}

func (hk HashKey) AsType(t KeyType) (newkey HashKey) {
	copy(newkey[4:], hk[4:])
	switch t {
	case PK:
		copy(newkey[:4], PK[:])
	case URI:
		copy(newkey[:4], URI[:])
	case ENTITY:
		copy(newkey[:4], ENTITY[:])
	case EXTENDED:
		copy(newkey[:4], EXTENDED[:])
	case PREDICATE:
		copy(newkey[:4], PREDICATE[:])
	}
	return newkey
}

func (hk HashKey) Less(than btree.Item, ctx interface{}) bool {
	t := than.(HashKey)
	return hk.LessThan(t)
}

func (hk HashKey) LessThan(other HashKey) bool {
	return binary.LittleEndian.Uint32(hk[4:]) < binary.LittleEndian.Uint32(other[4:])
}

/*
 ******************************
 * Bytes Entity
 ******************************
 */

type BytesEntity struct {
	PK HashKey `msg:"p"`
	// note: we have to use string keys to get msgp to work
	InEdges  map[string][]HashKey `msg:"i"`
	OutEdges map[string][]HashKey `msg:"o"`
}

func NewEntity(key HashKey) *BytesEntity {
	return &BytesEntity{
		PK:       key.AsType(PK),
		InEdges:  make(map[string][]HashKey),
		OutEdges: make(map[string][]HashKey),
	}
}

func (e *BytesEntity) Key() HashKey {
	return e.PK
}

func (e *BytesEntity) Bytes() []byte {
	b, err := e.MarshalMsg(nil)
	if err != nil {
		panic(err)
	}
	return b
}

func (e *BytesEntity) FromBytes(b []byte) error {
	_, err := e.UnmarshalMsg(b)
	return err
}

// returns true if we added an endpoint; false if it was already there
func (e *BytesEntity) AddInEdge(predicate, endpoint HashKey) bool {
	var (
		edgeList []HashKey
		found    bool
	)
	// check if we already have an edgelist for the given predicate
	if edgeList, found = e.InEdges[string(predicate[:])]; !found {
		// if we don't, then create a new one and put the endpoint in it
		edgeList = []HashKey{endpoint}
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
func (e *BytesEntity) AddOutEdge(predicate, endpoint HashKey) bool {
	var (
		edgeList []HashKey
		found    bool
	)
	// check if we already have an edgelist for the given predicate
	if edgeList, found = e.OutEdges[string(predicate[:])]; !found {
		// if we don't, then create a new one and put the endpoint in it
		edgeList = []HashKey{endpoint}
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

func (e *BytesEntity) ListInEndpoints(predicate HashKey) []HashKey {
	return e.InEdges[string(predicate[:])]
}

func (e *BytesEntity) ListOutEndpoints(predicate HashKey) []HashKey {
	return e.OutEdges[string(predicate[:])]
}

func (e *BytesEntity) GetAllPredicates() (preds []HashKey) {
	for pred := range e.InEdges {
		var hk HashKey
		copy(hk[:], []byte(pred))
		preds = append(preds, hk)
	}
	for pred := range e.OutEdges {
		var hk HashKey
		copy(hk[:], []byte(pred))
		preds = append(preds, hk)
	}
	return
}

/*
 ******************************
 * Bytes Predicate Entity
 ******************************
 */
type BytesPredicateEntity struct {
	PK HashKey `msg:"p"`
	// note: we have to use string keys to get msgp to work
	Subjects map[string]map[string]uint32 `msg:"s"`
	Objects  map[string]map[string]uint32 `msg:"o"`
}

func NewPredicateEntity(key HashKey) *BytesPredicateEntity {
	return &BytesPredicateEntity{
		PK:       key.AsType(PK),
		Subjects: make(map[string]map[string]uint32),
		Objects:  make(map[string]map[string]uint32),
	}
}

func (e *BytesPredicateEntity) Key() HashKey {
	return e.PK
}

func (e *BytesPredicateEntity) Bytes() []byte {
	b, err := e.MarshalMsg(nil)
	if err != nil {
		panic(err)
	}
	return b
}

func (e *BytesPredicateEntity) FromBytes(b []byte) error {
	_, err := e.UnmarshalMsg(b)
	return err
}

// adds subject/object to Predicate index entry. Returns true if this changed the entity
func (e *BytesPredicateEntity) AddSubjectObject(subject, object HashKey) bool {
	changed := false
	// if we have the subject
	if submap, found := e.Subjects[string(subject[:])]; found {
		// find the map of related objects
		if _, foundobj := submap[string(object[:])]; !foundobj {
			e.Subjects[string(subject[:])][string(object[:])] = 0
			changed = true
		}
	} else {
		e.Subjects[string(subject[:])] = map[string]uint32{string(object[:]): 0}
		changed = true
	}

	if objmap, found := e.Objects[string(object[:])]; found {
		// find the map of related objects
		if _, foundsub := objmap[string(subject[:])]; !foundsub {
			e.Objects[string(object[:])][string(subject[:])] = 0
			changed = true
		}
	} else {
		e.Objects[string(object[:])] = map[string]uint32{string(subject[:]): 0}
		changed = true
	}

	return changed
}

func (e *BytesPredicateEntity) GetObjects(subject HashKey) (entities []HashKey) {
	if objects, found := e.Subjects[string(subject[:])]; found {
		for objectstr := range objects {
			var hk HashKey
			copy(hk[:], []byte(objectstr))
			entities = append(entities, hk)
		}
	}
	return
}

func (e *BytesPredicateEntity) GetSubjects(object HashKey) (entities []HashKey) {
	if subjects, found := e.Objects[string(object[:])]; found {
		for subjectstr := range subjects {
			var hk HashKey
			copy(hk[:], []byte(subjectstr))
			entities = append(entities, hk)
		}
	}
	return
}

func (e *BytesPredicateEntity) GetAllObjects() (entities []HashKey) {
	for object := range e.Objects {
		var hk HashKey
		copy(hk[:], []byte(object))
		entities = append(entities, hk)
	}
	return
}

func (e *BytesPredicateEntity) GetAllSubjects() (entities []HashKey) {
	for subject := range e.Subjects {
		var hk HashKey
		copy(hk[:], []byte(subject))
		entities = append(entities, hk)
	}
	return
}

/*
 ******************************
 * Bytes Entity Extended
 ******************************
 */
type BytesEntityExtendedIndex struct {
	PK           HashKey              `msg:"p"`
	InPlusEdges  map[string][]HashKey `msg:"i+"`
	OutPlusEdges map[string][]HashKey `msg:"o+"`
}

func NewEntityExtendedIndex(key HashKey) *BytesEntityExtendedIndex {
	return &BytesEntityExtendedIndex{
		PK:           key.AsType(PK),
		InPlusEdges:  make(map[string][]HashKey),
		OutPlusEdges: make(map[string][]HashKey),
	}
}

func (e *BytesEntityExtendedIndex) Key() HashKey {
	return e.PK
}

func (e *BytesEntityExtendedIndex) Bytes() []byte {
	// TODO: implement
	b, err := e.MarshalMsg(nil)
	if err != nil {
		panic(err)
	}
	return b
}

func (e *BytesEntityExtendedIndex) FromBytes(b []byte) error {
	_, err := e.UnmarshalMsg(b)
	return err
}

// returns true if we added an endpoint; false if it was already there
func (e *BytesEntityExtendedIndex) AddOutPlusEdge(predicate, endpoint HashKey) bool {
	var (
		edgeList []HashKey
		found    bool
	)
	// check if we already have an edgelist for the given predicate
	if edgeList, found = e.OutPlusEdges[string(predicate[:])]; !found {
		// if we don't, then create a new one and put the endpoint in it
		edgeList = []HashKey{endpoint}
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
func (e *BytesEntityExtendedIndex) AddInPlusEdge(predicate, endpoint HashKey) bool {
	var (
		edgeList []HashKey
		found    bool
	)
	// check if we already have an edgelist for the given predicate
	if edgeList, found = e.InPlusEdges[string(predicate[:])]; !found {
		// if we don't, then create a new one and put the endpoint in it
		edgeList = []HashKey{endpoint}
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

func (e *BytesEntityExtendedIndex) ListInPlusEndpoints(predicate HashKey) (entities []HashKey) {
	return e.InPlusEdges[string(predicate[:])]
}

func (e *BytesEntityExtendedIndex) ListOutPlusEndpoints(predicate HashKey) (entities []HashKey) {
	return e.OutPlusEdges[string(predicate[:])]
}
