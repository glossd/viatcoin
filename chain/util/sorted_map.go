package util

import (
	"slices"
	"sync"
)

type SortedMap[K comparable, V any] struct {
	mu sync.RWMutex
	inner map[K]V
	order []K
}

func (sm *SortedMap[K, V]) init() {
	if sm.inner == nil {
		sm.inner = make(map[K]V, 16)
	}
	if sm.order == nil {
		sm.order = []K{}
	}
}

// O(n), but this method shouldn't be used in a blockchain :)
func (sm *SortedMap[K, V]) Delete(key K) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.init()

	delete(sm.inner, key)
	i := slices.Index(sm.order, key)
	if i == -1 {
		return
	}
	_ = slices.Delete(sm.order, i, i+1)
}
func (sm *SortedMap[K, V]) Load(key K) (value V, ok bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	sm.init()

	value, ok = sm.inner[key]
	return
}

func (sm *SortedMap[K, V]) LoadIndex(i int) V {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	sm.init()

	return sm.inner[sm.order[i]]
}

func (sm *SortedMap[K, V]) Last() V {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	sm.init()

	return sm.inner[sm.order[len(sm.order)-1]]
}

// sm[i:j]
func (sm *SortedMap[K, V]) LoadRange(i, j int) []V {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	sm.init()

	var res []V
	for _, key := range sm.order[i:j] {
		res = append(res, sm.inner[key])
	}
	return res
}

// sm[i:j]
func (sm *SortedMap[K, V]) LoadRangeSafe(i, j int) []V {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	sm.init()

	var res []V
	if i < 0 {
		i = 0
	}
	if j > len(sm.order) {
		j = len(sm.order)
	} 
	for _, key := range sm.order[i:j] {
		res = append(res, sm.inner[key])
	}
	return res
}

func (sm *SortedMap[K, V]) Store(key K, value V) { 
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.init()

	sm.inner[key] = value
	sm.order = append(sm.order, key)
}

func (sm *SortedMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) { 
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.init()
	
	old , ok := sm.inner[key]
	if ok {
		return old, true
	} else {
		sm.inner[key] = value
		sm.order = append(sm.order, key)
		return value, false
	}
}

func (sm *SortedMap[K, V]) Len() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	sm.init()

	return len(sm.order)
}

func (sm *SortedMap[K, V]) Clear() { 
	sm.mu.Lock()
	defer sm.mu.Unlock()

	clear(sm.inner)
	clear(sm.order)
}
