package bw2bind

import (
	"encoding/binary"
	"os"

	"github.com/dgraph-io/badger"
	"github.com/zhangxinngang/murmur"
	"gopkg.in/vmihailenco/msgpack.v3"
)

// this is a simple WAL backed by a directory provided as a configuration argument to newWal.
// Each message to be published is defined by a PublishParams struct. The WAL stores each pending
// PublishParams struct in a key-value store keyed by the murmur3 hash of the msgpack-serialized
// PublishParams struct. When the client receives notification from the local bw2 agent that
// the message has been published, the WAL removes the entry.
//
// When the WAL is created, it automatically iterates through the unsent messages and attempts
// to send them again. Currently all of these messages are copied into memory. Each message is then
// re-submitted through the above process (this means that the 'replay' is also durable).
type wal struct {
	db  *badger.DB
	dir string
}

func newWal(dir string) (*wal, error) {
	if err := os.MkdirAll(dir, os.ModeDir|0700); err != nil {
		return nil, err
	}

	opts := badger.DefaultOptions
	opts.Dir = dir
	opts.ValueDir = dir
	opts.SyncWrites = true
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	w := &wal{
		db:  db,
		dir: dir,
	}

	return w, nil
}

func (w *wal) add(params PublishParams) ([]byte, error) {
	seqnobytes := make([]byte, 4)
	err := w.db.Update(func(txn *badger.Txn) error {
		b, err := msgpack.Marshal(params)
		if err != nil {
			return err
		}

		binary.LittleEndian.PutUint32(seqnobytes, murmur.Murmur3(b))

		return txn.Set(seqnobytes, b)
	})
	return seqnobytes, err
}

func (w *wal) done(hash []byte) error {
	return w.db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(hash)
		if err == badger.ErrKeyNotFound {
			return nil // nothing to delete
		}
		if err != nil {
			return err
		}
		return txn.Delete(hash)
	})
}

func (w *wal) pending() ([]*PublishParams, error) {
	var topublish []*PublishParams
	err := w.db.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			v, err := item.Value()
			if err != nil {
				return err
			}
			var p PublishParams
			if err := msgpack.Unmarshal(v, &p); err != nil {
				return err
			}
			topublish = append(topublish, &p)

			txn.Delete(item.Key())
		}
		return nil
	})
	return topublish, err
}
