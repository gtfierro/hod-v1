package db

type keymap struct {
	m map[Key]struct{}
}

func newKeymap() *keymap {
	return &keymap{
		m: make(map[Key]struct{}),
	}
}

func (pt *keymap) Add(ent Key) {
	pt.m[ent] = struct{}{}
}

func (pt *keymap) Has(ent Key) bool {
	_, found := pt.m[ent]
	return found
}

func (pt *keymap) Len() int {
	return len(pt.m)
}

func (pt *keymap) Iter(iter func(ent Key)) {
	for k := range pt.m {
		iter(k)
	}
}

func (pt *keymap) Delete(k Key) {
	delete(pt.m, k)
}

func (pt *keymap) DeleteMax() Key {
	max := pt.Max()
	delete(pt.m, max)
	return max
}

func (pt *keymap) Max() Key {
	var max Key
	for k := range pt.m {
		max = k
		break
	}
	return max
}
