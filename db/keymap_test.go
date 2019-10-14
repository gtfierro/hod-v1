package db

import (
	"encoding/binary"
	"github.com/gtfierro/hod/storage"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKeyMap(t *testing.T) {
	assert := assert.New(t)

	km1 := newKeymap()
	assert.NotNil(km1)

	// test add
	key1 := storage.HashKey{0, 0, 0, 0, 0, 0, 0, 1}
	km1.Add(key1)
	_, found := km1.m[key1]
	assert.True(found, "Key should be found")
	assert.True(km1.Has(key1), "Key should be found")
	assert.Equal(km1.Len(), 1)

	key2 := storage.HashKey{0, 0, 0, 0, 0, 0, 0, 2}
	km1.Add(key2)
	_, found = km1.m[key2]
	assert.True(found, "Key should be found")
	assert.True(km1.Has(key2), "Key should be found")
	assert.Equal(km1.Len(), 2)

	// add again idempotently
	key3 := storage.HashKey{0, 0, 0, 0, 0, 0, 0, 2}
	km1.Add(key3)
	_, found = km1.m[key3]
	assert.True(found, "Key should be found")
	assert.True(km1.Has(key3), "Key should be found")
	assert.Equal(km1.Len(), 2)

	// test delete
	km1.Delete(key2)
	_, found = km1.m[key2]
	assert.False(found, "Key should not be found")
	assert.False(km1.Has(key2), "Key should not be found")
	assert.Equal(km1.Len(), 1)

	// add it back and delete max
	km1.Add(key2)
	_, found = km1.m[key2]
	assert.True(found, "Key should be found")
	assert.True(km1.Has(key2), "Key should be found")
	assert.Equal(km1.Len(), 2)

	assert.Equal(km1.Max(), key2)
	max1 := km1.DeleteMax()
	assert.Equal(km1.Max(), key1)
	assert.Equal(max1, key2)
	assert.Equal(km1.Len(), 1)
	_, found = km1.m[key2]
	assert.False(found, "Key should not be found")
	assert.False(km1.Has(key2), "Key should not be found")
}

func BenchmarkKeyMapAdd(b *testing.B) {
	b.ReportAllocs()
	km := newKeymap()
	for i := 0; i < b.N; i++ {
		var key storage.HashKey
		binary.LittleEndian.PutUint32(key[:], uint32(i))
		km.Add(key)
	}
}

func BenchmarkKeyMapAdd1(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		generateKeyMap(500, 0)
	}
}
