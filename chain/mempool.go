package chain

import (
	"fmt"
	"reflect"

	"github.com/glossd/viatcoin/chain/util"
)

// bitcoin is using levelDB. It's persistent kv-storage sorted by keys.
var memPool = util.Map[string, Transaction]{}

// set of unspent transaction. Bitcoin has unspent outputs.
var unspentTxs = util.Map[string, Transaction]{}

func Push(t Transaction) error {
	err := verifyTx(t)
	if err != nil {
		return err
	}

	memPool.LoadOrStore(t.Hash, t)
	return nil
}

func verifyTx(t Transaction) error {
	if err := t.Verify(); err != nil {
		return err
	}

	previous, ok := GetUnspent(t.PreviousHash)
	if !ok {
		return fmt.Errorf("previous transaction is not unspent: %s", t.PreviousHash)
	}

	if previous.Balance >= t.Balance {
		return fmt.Errorf("previous transaction balance must be less than old one")
	}
	return nil
}

func Get(hash string) (Transaction, bool) {
	return memPool.Load(hash)
}

func GetUnspent(hash string) (Transaction, bool) {
	return unspentTxs.Load(hash)
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
		if _, ok := unspentTxs.Load(t.PreviousHash); !ok {
			return fmt.Errorf("already spent transaction: %s", t.Hash)
		}
	}
	return nil
}

func MarkIngested(ts []Transaction) {
	for _, t := range ts {
		memPool.Delete(t.Hash)
		unspentTxs.Delete(t.PreviousHash)
		unspentTxs.Store(t.Hash, t)
	}
}
