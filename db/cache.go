package db

import (
	"github.com/coocood/freecache"
	"github.com/gtfierro/hod/turtle"
	"sync"
)

// cache overlay for databases
type dbcache struct {
	entityHashCache   *freecache.Cache
	entityObjectCache *freecache.Cache
	entityIndexCache  map[Key]*EntityExtendedIndex
	uriCache          map[Key]turtle.URI
	predCache         map[Key]*PredicateEntity
	sync.RWMutex
}

// size in mb of each cache
func newCache(maxsize int) *dbcache {
	return &dbcache{
		entityHashCache:   freecache.NewCache(maxsize * 1024 * 1024), // in MB
		entityObjectCache: freecache.NewCache(maxsize * 1024 * 1024),
		entityIndexCache:  make(map[Key]*EntityExtendedIndex),
		uriCache:          make(map[Key]turtle.URI),
		predCache:         make(map[Key]*PredicateEntity),
	}
}

func (cache dbcache) getHash(uri turtle.URI) (Key, bool) {
	var rethash Key
	if hash, err := cache.entityHashCache.Get(uri.Bytes()); err != nil {
		return rethash, false
	} else {
		copy(rethash[:], hash)
		return rethash, true
	}
}

func (cache *dbcache) setHash(uri turtle.URI, hash Key) {
	cache.entityHashCache.Set(uri.Bytes(), hash[:], -1) // no expiry
	cache.setURI(hash, uri)
}

func (cache *dbcache) getURI(hash Key) (turtle.URI, bool) {
	uri, found := cache.uriCache[hash]
	return uri, found
}

func (cache *dbcache) setURI(hash Key, uri turtle.URI) {
	cache.uriCache[hash] = uri
}

func (cache *dbcache) getEntityByHash(hash Key) (*Entity, bool) {
	bytes, err := cache.entityObjectCache.Get(hash[:])
	if err != nil {
		return nil, false
	}
	ent := NewEntity()
	if _, err = ent.UnmarshalMsg(bytes); err != nil {
		return nil, false
	}

	return ent, true
}

func (cache *dbcache) setEntityBytesByHash(hash Key, entbytes []byte) {
	cache.entityObjectCache.Set(hash[:], entbytes, -1) // no expiry
}

func (cache *dbcache) getExtendedIndexByHash(hash Key) (*EntityExtendedIndex, bool) {
	ext, found := cache.entityIndexCache[hash]
	return ext, found
}

func (cache *dbcache) setExtendedIndexByHash(hash Key, ext *EntityExtendedIndex) {
	cache.entityIndexCache[hash] = ext
}

func (cache *dbcache) getPredicateByHash(hash Key) (*PredicateEntity, bool) {
	pred, found := cache.predCache[hash]
	return pred, found
}

func (cache *dbcache) setPredicateByHash(hash Key, pred *PredicateEntity) {
	cache.predCache[hash] = pred
}
