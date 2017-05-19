//go:generate msgp
package db

import (
	"encoding/binary"

	"github.com/mitghi/btree"
)

type Key [4]byte

func (k Key) Less(than btree.Item, ctx interface{}) bool {
	t := than.(Key)
	return binary.LittleEndian.Uint32(k[:]) < binary.LittleEndian.Uint32(t[:])
}

func (k *Key) FromSlice(src []byte) {
	copy(k[:], src)
}

func (k Key) String() string {
	return string(k[:])
}

func (k Key) Uint32() uint32 {
	return binary.LittleEndian.Uint32(k[:])
}

func (k *Key) FromUint32(s uint32) {
	binary.LittleEndian.PutUint32(k[:], s)
}
