package db

import (
	"sync"
	"sync/atomic"
	//"time"

	"github.com/gtfierro/hod/turtle"
)

// cache overlay for databases
type dbcache struct {
	entityHashCache   map[turtle.URI]Key
	uriCache          map[Key]turtle.URI
	entityObjectCache map[Key]*Entity
	entityIndexCache  map[Key]*EntityExtendedIndex
	predCache         map[Key]*PredicateEntity
	hit               uint64
	total             uint64
	pendingEvict      chan Key
	sync.RWMutex
}

// size in mb of each cache
func newCache(maxsize int) *dbcache {
	c := &dbcache{
		entityHashCache:   make(map[turtle.URI]Key),
		entityObjectCache: make(map[Key]*Entity),
		entityIndexCache:  make(map[Key]*EntityExtendedIndex),
		uriCache:          make(map[Key]turtle.URI),
		predCache:         make(map[Key]*PredicateEntity),
		pendingEvict:      make(chan Key, 1e6),
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

func (cache *dbcache) getHash(uri turtle.URI) (Key, bool) {
	cache.RLock()
	defer cache.RUnlock()
	hash, found := cache.entityHashCache[uri]
	cache.markHitOrMiss(found)
	return hash, found
}

func (cache *dbcache) evict(hash Key) {
	cache.pendingEvict <- hash
}

func (cache *dbcache) setHash(uri turtle.URI, hash Key) {
	cache.Lock()
	cache.entityHashCache[uri] = hash
	cache.Unlock()
}

func (cache *dbcache) getURI(hash Key) (turtle.URI, bool) {
	cache.RLock()
	defer cache.RUnlock()
	uri, found := cache.uriCache[hash]
	cache.markHitOrMiss(found)
	return uri, found
}

func (cache *dbcache) setURI(hash Key, uri turtle.URI) {
	cache.Lock()
	cache.uriCache[hash] = uri
	cache.Unlock()
}

func (cache *dbcache) getEntityByHash(hash Key) (*Entity, bool) {
	cache.RLock()
	defer cache.RUnlock()
	ent, found := cache.entityObjectCache[hash]
	cache.markHitOrMiss(found)
	return ent, found
}

func (cache *dbcache) setEntityByHash(hash Key, ent *Entity) {
	cache.Lock()
	cache.entityObjectCache[hash] = ent
	cache.Unlock()
}

func (cache *dbcache) getExtendedIndexByHash(hash Key) (*EntityExtendedIndex, bool) {
	cache.RLock()
	defer cache.RUnlock()
	ext, found := cache.entityIndexCache[hash]
	cache.markHitOrMiss(found)
	return ext, found
}

func (cache *dbcache) setExtendedIndexByHash(hash Key, ext *EntityExtendedIndex) {
	cache.Lock()
	cache.entityIndexCache[hash] = ext
	cache.Unlock()
}

func (cache *dbcache) getPredicateByHash(hash Key) (*PredicateEntity, bool) {
	cache.RLock()
	defer cache.RUnlock()
	pred, found := cache.predCache[hash]
	cache.markHitOrMiss(found)
	return pred, found
}

func (cache *dbcache) setPredicateByHash(hash Key, pred *PredicateEntity) {
	cache.Lock()
	cache.predCache[hash] = pred
	cache.Unlock()
}
