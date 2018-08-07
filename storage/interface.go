package storage

import (
	"errors"

	"github.com/gtfierro/hod/config"
	"github.com/gtfierro/hod/turtle"
)

var ErrNotFound = errors.New("Not found")
var ErrGraphNotFound = errors.New("Graph not found")

//const LATEST = Version(0)

type StorageProvider interface {
	// provides configuration information to the storage provider.
	// When this method returns, the storage provider is allowed to be used
	Initialize(cfg *config.Config) error

	// Closes the storage provider. Further calls to the storage provider
	// should return an error
	Close() error

	// Adds a new graph to the storage provider under the given name,
	// Returns the version of the graph and a boolean 'exists' value
	// that is true if the database already existed.
	AddGraph(name string) (Version, bool, error)

	// creates and returns a new writable version of the graph with the given name.
	// This version will not be available until it is committed
	CreateVersion(name string) (Transaction, error)

	// returns the given version of the graph with the given name; returns an error if the version doesn't exist
	OpenVersion(ver Version) (Transaction, error)

	// lists versions of the graph with the given name
	ListVersions(name string) ([]Version, error)

	Graphs() ([]Version, error)
}

type Traversable interface {
	GetHash(turtle.URI) (HashKey, error)
	GetURI(HashKey) (turtle.URI, error)
	GetEntity(HashKey) (Entity, error)
	GetExtendedIndex(HashKey) (EntityExtendedIndex, error)
	GetPredicate(HashKey) (PredicateEntity, error)
	GetReversePredicate(turtle.URI) (turtle.URI, bool)
	IterateAllEntities(func(HashKey, Entity) bool) error
}

type Transaction interface {
	Traversable
	Commit() error
	Release()
	Version() Version

	PutURI(turtle.URI) (HashKey, error)
	PutEntity(Entity) error
	PutExtendedIndex(EntityExtendedIndex) error
	PutPredicate(PredicateEntity) error
	PutReversePredicate(turtle.URI, turtle.URI) error
}

type Snapshot interface {
	Traversable
	Release()
	Version() Version
	Commit() error
}

type Entity interface {
	Key() HashKey
	Bytes() []byte
	FromBytes([]byte) error
	AddInEdge(predicate, endpoint HashKey) bool
	AddOutEdge(predicate, endpoint HashKey) bool
	ListInEndpoints(predicate HashKey) []HashKey
	ListOutEndpoints(predicate HashKey) []HashKey
	GetAllPredicates() []HashKey
}

type PredicateEntity interface {
	Key() HashKey
	Bytes() []byte
	FromBytes([]byte) error
	AddSubjectObject(subject, object HashKey) bool
	GetObjects(subject HashKey) []HashKey
	GetSubjects(object HashKey) []HashKey
	GetAllObjects() []HashKey
	GetAllSubjects() []HashKey
}

type EntityExtendedIndex interface {
	Key() HashKey
	Bytes() []byte
	FromBytes([]byte) error
	AddInPlusEdge(predicate, endpoint HashKey) bool
	AddOutPlusEdge(predicate, endpoint HashKey) bool
	ListInPlusEndpoints(predicate HashKey) []HashKey
	ListOutPlusEndpoints(predicate HashKey) []HashKey
}
