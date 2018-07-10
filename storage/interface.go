package storage

import (
	"github.com/gtfierro/hod/config"
)

type StorageProvider interface {
	Initialize(name string, cfg *config.Config) error
	Close() error
	OpenTransaction() (Transaction, error)
	OpenVersion(version uint64) (Traversable, error)
}

type Traversable interface {
	Has(bucket HodBucket, key []byte) (exists bool, err error)
	Get(bucket HodBucket, key []byte) (value []byte, err error)
	Put(bucket HodBucket, key, value []byte) (err error)
	Iterate(bucket HodBucket) Iterator
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

/*
Storage API
- open database (name)
- close database
- buckets
- for each bucket
    - get/put key value
    - iterate
- transactions
*/
