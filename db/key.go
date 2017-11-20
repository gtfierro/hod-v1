//go:generate msgp
package db

import (
	"encoding/binary"

	"github.com/mitghi/btree"
)

type Key [4]byte

func (k Key) Less(than btree.Item, ctx interface{}) bool {
	t := than.(Key)
	return k.LessThan(t)
}

func (k Key) LessThan(other Key) bool {
	return binary.LittleEndian.Uint32(k[:]) < binary.LittleEndian.Uint32(other[:])
}

func (k *Key) FromSlice(src []byte) {
	copy(k[:], src)
}

func (k Key) String() string {
	return string(k[:])
}
