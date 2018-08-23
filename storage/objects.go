//go:generate msgp
package storage

//TODO: switch to iterator-type interface rather than creating lists

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/gtfierro/btree"
)

// Version represents the state of a Brick model at a given timestamp
type Version struct {
	Timestamp uint64
	Name      string
}

// Return the string representation of the version
func (v Version) String() string {
	return fmt.Sprintf("<Version %s.%d>", v.Name, v.Timestamp)
}

func (v Version) Empty() bool {
	return v.Timestamp == 0
}

// HashKey is the primary key for Brick entities. It maps one-to-one with a URI
type HashKey [8]byte

// this is the empty key
var EmptyKey = HashKey{}

// HashKey has a 4-byte prefix
type KeyType [4]byte

var (
	// belongs in PK bucket
	PK KeyType = [4]byte{0, 0, 0, 0}
	// belongs in URI bucket
	URI KeyType = [4]byte{0, 0, 0, 1}
	// belongs in Entity bucket
	ENTITY KeyType = [4]byte{0, 0, 0, 2}
	// belongs in Extended bucket
	EXTENDED KeyType = [4]byte{0, 0, 0, 3}
	// belongs in Predicate bucket
	PREDICATE KeyType = [4]byte{0, 0, 0, 4}
)

// returns the type of the HashKey
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

// returns a copy of the HashKey as a certain type
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

// Implement Less() for btree
func (hk HashKey) Less(than btree.Item, ctx interface{}) bool {
	t := than.(HashKey)
	return hk.LessThan(t)
}

// Implement LessThan() for btree
func (hk HashKey) LessThan(other HashKey) bool {
	return binary.LittleEndian.Uint32(hk[4:]) < binary.LittleEndian.Uint32(other[4:])
}

type HashKeyGenerator struct {
	num uint32
	pfx uint32
}

func NewHashKeyGenerator(prefix uint32) *HashKeyGenerator {
	return &HashKeyGenerator{
		num: 0,
		pfx: prefix,
	}
}

func (gen *HashKeyGenerator) GetKey() HashKey {
	var key HashKey
	gen.num++
	binary.LittleEndian.PutUint32(key[:], gen.pfx)
	binary.LittleEndian.PutUint32(key[len(key)-4:], gen.num)
	return key
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

func (e *BytesEntity) Copy() Entity {
	e2 := NewEntity(e.Key())
	for k, v := range e.InEdges {
		e2.InEdges[k] = make([]HashKey, len(v))
		copy(e2.InEdges[k], v)
	}
	for k, v := range e.OutEdges {
		e2.OutEdges[k] = make([]HashKey, len(v))
		copy(e2.OutEdges[k], v)
	}
	return e2
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

func (e *BytesPredicateEntity) Copy() PredicateEntity {
	e2 := NewPredicateEntity(e.Key())
	for k, v := range e.Subjects {
		e2.Subjects[k] = make(map[string]uint32)
		for vk, vv := range v {
			e2.Subjects[k][vk] = vv
		}
	}
	for k, v := range e.Objects {
		e2.Objects[k] = make(map[string]uint32)
		for vk, vv := range v {
			e2.Objects[k][vk] = vv
		}
	}
	return e2
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

func (e *BytesEntityExtendedIndex) Copy() EntityExtendedIndex {
	e2 := NewEntityExtendedIndex(e.Key())
	for k, v := range e.InPlusEdges {
		e2.InPlusEdges[k] = make([]HashKey, len(v))
		copy(e2.InPlusEdges[k], v)
	}
	for k, v := range e.OutPlusEdges {
		e2.OutPlusEdges[k] = make([]HashKey, len(v))
		copy(e2.OutPlusEdges[k], v)
	}
	return e2
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
