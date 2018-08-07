package db

import (
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

func (pt *keymap) Add(ent storage.HashKey) {
	pt.m[ent] = struct{}{}
}

func (pt *keymap) Has(ent storage.HashKey) bool {
	_, found := pt.m[ent]
	return found
}

func (pt *keymap) Len() int {
	return len(pt.m)
}

func (pt *keymap) Iter(iter func(ent storage.HashKey)) {
	for k := range pt.m {
		iter(k)
	}
}

func (pt *keymap) Delete(k storage.HashKey) {
	delete(pt.m, k)
}

func (pt *keymap) DeleteMax() storage.HashKey {
	max := pt.Max()
	delete(pt.m, max)
	return max
}

func (pt *keymap) Max() storage.HashKey {
	var max storage.HashKey
	for k := range pt.m {
		max = k
		break
	}
	return max
}
