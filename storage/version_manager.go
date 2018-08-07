package storage

import (
//	"context"
//	"sync"
//	"time"
//
//	"golang.org/x/sync/semaphore"
)

//type managed interface {
//	Close() error
//	Open() error
//}
//
//// has configurable N slots for open databases; databases are closed after M minutes of inactivity
//// Block extra requests until a spare slot is opened
//type versionManager struct {
//	max_open        int
//	sem             *semaphore.Weighted
//	timeout         time.Duration
//	loaded_versions map[Version]*time.Timer
//	sync.RWMutex
//}
//
//func initializeVersionManager(max_open int, timeout time.Duration) *versionManager {
//	return &versionManager{
//		max_open:        max_open,
//		timeout:         timeout,
//		loaded_versions: make(map[Version]*time.Timer),
//		sem:             semaphore.NewWeighted(max_open),
//	}
//}
//
//func (vm *versionManager) open(ctx context.Context, version Version, cb func()) error {
//	vm.RLock()
//	if timer, found := vm.loaded_versions[version]; found {
//		timer.Reset()
//		vm.RUnlock()
//		return
//	}
//	vm.RUnlock()
//
//	if err := vm.sem.Acquire(ctx, 1); err != nil {
//		return err
//	}
//
//	vm.Lock()
//	if _, found := vm.loaded_versions[version]; found {
//		vm.Unlock()
//		return
//	}
//	vm.loaded_versions[version] = time.AfterFunc(vm.timeout, func() {
//		cb()
//		vm.sem.Release(1)
//	})
//	vm.Unlock()
//
//	return nil
//}
