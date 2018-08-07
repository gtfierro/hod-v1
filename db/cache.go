package db

import (
	"sync"
	"sync/atomic"

	"github.com/gtfierro/hod/storage"
	"github.com/gtfierro/hod/turtle"
)

// cache overlay for databases
type dbcache struct {
	entityHashCache   map[turtle.URI]storage.HashKey
	uriCache          map[storage.HashKey]turtle.URI
	entityObjectCache map[storage.HashKey]storage.Entity
	entityIndexCache  map[storage.HashKey]storage.EntityExtendedIndex
	predCache         map[storage.HashKey]storage.PredicateEntity
	hit               uint64
	total             uint64
	pendingEvict      chan storage.HashKey
	sync.RWMutex
}

// size in mb of each cache
func newCache(maxsize int) *dbcache {
	c := &dbcache{
		entityHashCache:   make(map[turtle.URI]storage.HashKey),
		entityObjectCache: make(map[storage.HashKey]storage.Entity),
		entityIndexCache:  make(map[storage.HashKey]storage.EntityExtendedIndex),
		uriCache:          make(map[storage.HashKey]turtle.URI),
		predCache:         make(map[storage.HashKey]storage.PredicateEntity),
		pendingEvict:      make(chan storage.HashKey, 1e6),
	}

	//var timer = time.NewTimer(5 * time.Second)
	go func() {
		for {
			select {
			case hash := <-c.pendingEvict:
				//timer = time.NewTimer(5 * time.Second)
				c.Lock()
				delete(c.entityObjectCache, hash)
				delete(c.entityIndexCache, hash)
				delete(c.predCache, hash)
				//for hash := range c.pendingEvict {
				//	delete(c.entityObjectCache, hash)
				//	delete(c.entityIndexCache, hash)
				//	delete(c.predCache, hash)
				//}
				c.Unlock()
				//case <-timer.C:
				//				timer.Reset(5 * time.Second)
			}
		}
	}()

	return c
}

func (cache *dbcache) markHitOrMiss(b bool) {
	if b {
		atomic.AddUint64(&cache.hit, 1)
	}
	atomic.AddUint64(&cache.total, 1)
}

func (cache *dbcache) getHash(uri turtle.URI) (storage.HashKey, bool) {
	cache.RLock()
	defer cache.RUnlock()
	hash, found := cache.entityHashCache[uri]
	cache.markHitOrMiss(found)
	return hash, found
}

func (cache *dbcache) evict(hash storage.HashKey) {
	cache.pendingEvict <- hash
}

func (cache *dbcache) setHash(uri turtle.URI, hash storage.HashKey) {
	cache.Lock()
	cache.entityHashCache[uri] = hash
	cache.Unlock()
}

func (cache *dbcache) getURI(hash storage.HashKey) (turtle.URI, bool) {
	cache.RLock()
	defer cache.RUnlock()
	uri, found := cache.uriCache[hash]
	cache.markHitOrMiss(found)
	return uri, found
}

func (cache *dbcache) setURI(hash storage.HashKey, uri turtle.URI) {
	cache.Lock()
	cache.uriCache[hash] = uri
	cache.Unlock()
}

func (cache *dbcache) getEntityByHash(hash storage.HashKey) (storage.Entity, bool) {
	cache.RLock()
	defer cache.RUnlock()
	ent, found := cache.entityObjectCache[hash]
	cache.markHitOrMiss(found)
	return ent, found
}

func (cache *dbcache) setEntityByHash(hash storage.HashKey, ent storage.Entity) {
	cache.Lock()
	cache.entityObjectCache[hash] = ent
	cache.Unlock()
}

func (cache *dbcache) getExtendedIndexByHash(hash storage.HashKey) (storage.EntityExtendedIndex, bool) {
	cache.RLock()
	defer cache.RUnlock()
	ext, found := cache.entityIndexCache[hash]
	cache.markHitOrMiss(found)
	return ext, found
}

func (cache *dbcache) setExtendedIndexByHash(hash storage.HashKey, ext storage.EntityExtendedIndex) {
	cache.Lock()
	cache.entityIndexCache[hash] = ext
	cache.Unlock()
}

func (cache *dbcache) getPredicateByHash(hash storage.HashKey) (storage.PredicateEntity, bool) {
	cache.RLock()
	defer cache.RUnlock()
	pred, found := cache.predCache[hash]
	cache.markHitOrMiss(found)
	return pred, found
}

func (cache *dbcache) setPredicateByHash(hash storage.HashKey, pred storage.PredicateEntity) {
	cache.Lock()
	cache.predCache[hash] = pred
	cache.Unlock()
}
