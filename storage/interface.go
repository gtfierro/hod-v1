package storage

import (
	"github.com/gtfierro/hod/config"
)

type StorageProvider interface {
	// sets up the structure needed for the database
	Initialize(name string, cfg *config.Config) error
	// closes the storage provider when HodDB is done
	Close() error
	// create a transaction for a new version of the database; always operates on the most recent version
	OpenTransaction() (Transaction, error)
	// open the specified version of the database
	OpenVersion(version uint64) (Traversable, error)
}

type Traversable interface {
	Has(bucket HodNamespace, key []byte) (exists bool, err error)
	Get(bucket HodNamespace, key []byte) (value []byte, err error)
	Put(bucket HodNamespace, key, value []byte) (err error)
	Iterate(bucket HodNamespace) Iterator
	Release()
}

type Transaction interface {
	Traversable
	Commit() error
}

type Iterator interface {
	Next() bool
	Key() []byte
	Value() []byte
}
