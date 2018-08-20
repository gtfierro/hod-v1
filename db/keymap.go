package db

import (
	"encoding/binary"
	"github.com/gtfierro/hod/storage"
)

type keymap struct {
	m map[storage.HashKey]struct{}
}

func newKeymap() *keymap {
	return &keymap{
		m: make(map[storage.HashKey]struct{}),
	}
}

func (km *keymap) Add(ent storage.HashKey) {
	km.m[ent] = struct{}{}
}

func (km *keymap) Has(ent storage.HashKey) bool {
	_, found := km.m[ent]
	return found
}

func (km *keymap) Len() int {
	return len(km.m)
}

func (km *keymap) Iter(iter func(ent storage.HashKey)) {
	for k := range km.m {
		iter(k)
	}
}

func (km *keymap) Delete(k storage.HashKey) {
	delete(km.m, k)
}

func (km *keymap) DeleteMax() storage.HashKey {
	max := km.Max()
	delete(km.m, max)
	return max
}

func (km *keymap) Max() storage.HashKey {
	var max storage.HashKey
	for k := range km.m {
		if max.LessThan(k) {
			max = k
		}
	}
	return max
}

func generateKeyMap(num, offset int) *keymap {
	km := newKeymap()
	for i := 0; i < num; i++ {
		var key storage.HashKey
		binary.LittleEndian.PutUint32(key[len(key)-4:], uint32(offset+i))
		km.Add(key)
	}
	return km
}
