package storage

import (
	"encoding/binary"

	"github.com/gtfierro/hod/turtle"
	"github.com/zhangxinngang/murmur"
)

func hashURI(u turtle.URI, dest *HashKey, salt uint64) {
	var hash uint32
	if salt > 0 {
		saltbytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(saltbytes, salt)
		hash = murmur.Murmur3(append(u.Bytes(), saltbytes...))
	} else {
		hash = murmur.Murmur3(u.Bytes())
	}
	binary.LittleEndian.PutUint32(dest[4:], hash)
}
