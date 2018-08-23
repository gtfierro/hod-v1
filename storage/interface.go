package storage

import (
	"errors"
	"time"

	"github.com/gtfierro/hod/config"
	"github.com/gtfierro/hod/turtle"
)

// ErrNotFound entity not found error
var ErrNotFound = errors.New("Not found")

// ErrGraphNotFound graph/version not found error
var ErrGraphNotFound = errors.New("Graph not found")

// The StorageProvider interface defines the methods required by HodDB in order to
// store and retrieve Brick Models from physical media
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
	// The returned transaction should be read-only
	OpenVersion(ver Version) (Transaction, error)

	// lists versions of the graph with the given name
	ListVersions(name string) ([]Version, error)

	// returns the latest version of the given graph
	CurrentVersion(name string) (Version, error)

	// returns the version of the given graph at the given timestamp
	VersionAt(name string, timestamp time.Time) (Version, error)

	// returns the version active before the one active at the given timestamp
	VersionBefore(name string, timestamp time.Time) (Version, error)

	// returns the version active after the one active at the given timestamp
	VersionAfter(name string, timestamp time.Time) (Version, error)

	// list all stored versions
	Graphs() ([]Version, error)

	// return the set of saved abbreviation -> namespace URI mappings
	GetNamespaces() (mapping map[string]string, err error)

	// save a new abbreviation -> namespace URI mapping (e.g. brick -> https://brickschema.org/schema/1.0.3/Brick#)
	SaveNamespace(abbreviation string, uri string) error
}

// Transaction defines the generic interface for read-only and read-write transactions for a StorageProvider
type Transaction interface {
	// Commit the transaction
	Commit() error
	// Release the transaction (read-only) or discard the transaction (rw)
	Release()
	// return the current Version of the Transaction
	Version() Version

	// retrive the HashKey for the given URI
	GetHash(turtle.URI) (HashKey, error)

	// retrive the URI for the given HashKey
	GetURI(HashKey) (turtle.URI, error)

	// retrive the Entity object for the given HashKey
	GetEntity(HashKey) (Entity, error)

	// retrive the ExtendedIndex object for the given HashKey
	GetExtendedIndex(HashKey) (EntityExtendedIndex, error)

	// retrive the PredicateEntity object for the given HashKey
	GetPredicate(HashKey) (PredicateEntity, error)

	// get the inverse edge for the given URI, if it exists
	GetReversePredicate(turtle.URI) (turtle.URI, bool)

	// call the provided function for each Entity object in the graph
	IterateAllEntities(func(HashKey, Entity) bool) error

	// store the URI and return the HashKey
	PutURI(turtle.URI) (HashKey, error)
	// store the Entity object, mapped by HashKey (Entity.Key())
	PutEntity(Entity) error
	// store the ExtendedIndex object, mapped by HashKey (ExtendedIndex.Key())
	PutExtendedIndex(EntityExtendedIndex) error
	// store the PredicateEntity object, mapped by HashKey (PredicateEntity.Key())
	PutPredicate(PredicateEntity) error
	// store the two URIs as inverses of each other
	PutReversePredicate(turtle.URI, turtle.URI) error
}

// Entity defines the read/write methods for the virtual Entity object in a Brick graph
type Entity interface {
	// returns the primary key for this Entity
	Key() HashKey
	// returns the byte representation of this Entity object
	Bytes() []byte
	// initializes this Enttity object from a serialized form
	FromBytes([]byte) error
	// copies the object
	Copy() Entity
	// add incoming predicate, subject. Returns false if already exists
	AddInEdge(predicate, endpoint HashKey) bool
	// add outgoing predicate, object. Returns false if already exists
	AddOutEdge(predicate, endpoint HashKey) bool
	// List all subjects with the given predicate to this entity
	ListInEndpoints(predicate HashKey) []HashKey
	// List all objects this entity has the given predicate to
	ListOutEndpoints(predicate HashKey) []HashKey
	// List all predicates this entity has
	GetAllPredicates() []HashKey
}

// PredicateEntity defines the read/write methods for the virtual PredicateEntity object in a Brick graph
type PredicateEntity interface {
	// returns the primary key for this PredicateEntity
	Key() HashKey
	// returns the byte representation of this PredicateEntity object
	Bytes() []byte
	// initializes this PredicateEntity object from its serialized form
	FromBytes([]byte) error
	// copies the object
	Copy() PredicateEntity
	// add subject/object pair that uses this predicate. Returns false if already exists
	AddSubjectObject(subject, object HashKey) bool
	// list all objects for this predicate with the given subject
	GetObjects(subject HashKey) []HashKey
	// list all subjects for this predicate with the given object
	GetSubjects(object HashKey) []HashKey
	// lits all objects for this predicate
	GetAllObjects() []HashKey
	// lits all subjects for this predicate
	GetAllSubjects() []HashKey
}

// EntityExtendedIndex defines the read/write methods for the virtual EntityExtendedIndex object in a Brick graph
type EntityExtendedIndex interface {
	// returns the primary key for this ExtendedIndex
	Key() HashKey
	// returns byte representation of this ExtendedIndex object
	Bytes() []byte
	// initializes this ExtendedIndex object from its serialized form
	FromBytes([]byte) error
	// copies the object
	Copy() EntityExtendedIndex
	// add 1+ predicate edge to this entity (from subject)
	AddInPlusEdge(predicate, endpoint HashKey) bool
	// add 1+ predicate edge to this entity (to object)
	AddOutPlusEdge(predicate, endpoint HashKey) bool
	// list all reachable entities 1+ edges away from this entity using the given incoming predicate
	ListInPlusEndpoints(predicate HashKey) []HashKey
	// list all reachable entities 1+ edges away from this entity using the given outgoing predicate
	ListOutPlusEndpoints(predicate HashKey) []HashKey
}
