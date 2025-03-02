package mempool

import (
	"fmt"
	"github.com/glossd/viatcoin/chain"
	"sync"
)

// bitcoin is using levelDB. It's persistent kv-storage sorted by keys.
var memPool = sync.Map{}

func Push(t chain.Transaction) error {
	if err := t.Verify(); err != nil {
		return err
	}

	previous, ok := Get(t.PreviousHash)
	if !ok {
		return fmt.Errorf("transaction not found with previous hash: %s", t.PreviousHash)
	}

	if previous.Balance >= t.Balance {
		return fmt.Errorf("previous transaction balance must be less than old one")
	}

	memPool.LoadOrStore(t.Hash, t)
	return nil
}

func Get(hash string) (chain.Transaction, bool) {
	t, ok := memPool.Load(hash)
	if !ok {
		return chain.Transaction{}, false
	}
	return t.(chain.Transaction), true
}

func Top(num int) []chain.Transaction {
	var res []chain.Transaction
	memPool.Range(func(key, value any) bool {
		if num == 0 {
			return false
		}
		res = append(res, value.(chain.Transaction))
		num--
		return true
	})

	return res
}

func Delete(ts []chain.Transaction) {
	for _, t := range ts {
		memPool.Delete(t.Hash)
	}
}
