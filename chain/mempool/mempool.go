package mempool

import (
	"github.com/glossd/viatcoin/chain"
	"slices"
	"sync"
)

var memPool = []chain.Transaction{}
var lock sync.RWMutex

func Push(t chain.Transaction) {
	lock.Lock()
	defer lock.Unlock()
	memPool = append(memPool, t)
}

func Pop(num int) []chain.Transaction {
	lock.RLock()
	defer lock.RUnlock()
	idx := num
	if len(memPool) < idx {
		idx = len(memPool)
	}
	return memPool[:]
}

func Delete(ts []chain.Transaction) {
	lock.Lock()
	defer lock.Unlock()
	var set map[string]bool
	for _, t := range ts {
		set[t.Hash] = true
	}

	slices.DeleteFunc(memPool, func(e chain.Transaction) bool {
		return set[e.Hash]
	})
}
