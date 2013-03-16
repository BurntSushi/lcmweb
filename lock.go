package main

import (
	"log"
	"sync"
)

type locks struct {
	locks   map[string]*sync.Mutex
	locksRW *sync.RWMutex
}

var locker *locks

func init() {
	locker = &locks{
		locks:   make(map[string]*sync.Mutex),
		locksRW: new(sync.RWMutex),
	}
}

func (lker *locks) getLock(key string) (*sync.Mutex, bool) {
	lker.locksRW.RLock()
	defer lker.locksRW.RUnlock()

	lock, ok := lker.locks[key]
	return lock, ok
}

func (lker *locks) deleteLock(key string) {
	lker.locksRW.Lock()
	defer lker.locksRW.Unlock()

	if _, ok := lker.locks[key]; ok {
		delete(lker.locks, key)
	}
}

func (lker *locks) lock(key string) {
	lk, ok := lker.getLock(key)
	if !ok {
		lker.locksRW.Lock()
		lk = new(sync.Mutex)
		lker.locks[key] = lk
		lker.locksRW.Unlock()
	}
	lk.Lock()
}

func (lker *locks) unlock(key string) {
	lk, ok := lker.getLock(key)
	if !ok {
		log.Printf("Lock for key '%s' not initialized.", key)
		panic(e("BUG: Lock not initialized."))
	}
	lk.Unlock()
}
