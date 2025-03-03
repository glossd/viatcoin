package chain

import (
	"fmt"
	"reflect"

	"github.com/glossd/viatcoin/chain/util"
)

// bitcoin is using levelDB. It's persistent kv-storage sorted by keys.
var memPool = util.Map[string, Transaction]{}

// set of unspent transaction. Bitcoin has unspent outputs.
var unspentTxs = util.Map[string, bool]{}

func Push(t Transaction) error {
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

func Get(hash string) (Transaction, bool) {
	return memPool.Load(hash)
}

func Top(num int) []Transaction {
	var res []Transaction
	memPool.Range(func(key string, value Transaction) bool {
		if num == 0 {
			return false
		}
		res = append(res, value)
		num--
		return true
	})

	return res
}

func ExistsUnspent(ts []Transaction) error {
	for _, t := range ts {
		mt, ok := Get(t.Hash)
		if !ok {
			return fmt.Errorf("transaction not found: %s", t.Hash)
		}
		if reflect.DeepEqual(t, mt) {
			return fmt.Errorf("transaction doesn't match: %s", t.Hash)
		}
		if _, ok := unspentTxs.Load(t.Hash); !ok {
			return fmt.Errorf("already spent transaction: %s", t.Hash)
		}
	}
	return nil
}

func Delete(ts []Transaction) {
	for _, t := range ts {
		memPool.Delete(t.Hash)
		unspentTxs.Delete(t.PreviousHash)
		unspentTxs.Store(t.Hash, true)
	}
}
