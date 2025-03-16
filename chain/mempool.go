package chain

import (
	"fmt"
	"reflect"

	"github.com/glossd/viatcoin/chain/util"
)

// bitcoin is using levelDB. It's persistent kv-storage sorted by keys.
// todo after adding miner fee, sort the transaction by it
var memPool = util.Map[string, Transaction]{}

var wallets = util.Map[string, []int64]{}

func Push(t Transaction) error {
	err := verifyTx(t)
	if err != nil {
		return err
	}

	_, ok := memPool.LoadOrStore(t.ID, t)
	if ok {
		return fmt.Errorf("transaciont already exists: %s", t.ID)
	}
	return nil
}

func verifyTx(t Transaction) error {
	if err := t.Verify(); err != nil {
		return err
	}

	mt, ok := memPool.Load(t.ID)
	if !ok {
		return fmt.Errorf("transaction is not in the mempool: %s", t.ID)
	}
	if reflect.DeepEqual(t, mt) {
		return fmt.Errorf("transaction doesn't match the one in mempool: %s", t.ID)
	}

	var fullAmount Coin
	for _, tf := range t.Transfers {
		fullAmount += tf.Amount
	}
	balance := Balance(t.From)
	if fullAmount > balance {
		return fmt.Errorf("the amount in trasfers exceeded wallet balance, required=%f, balance=%f", fullAmount.AsViatcoins(), balance.AsViatcoins())
	}

	return nil
}

func Get(hash string) (Transaction, bool) {
	return memPool.Load(hash)
}

func Balance(address string) Coin {
	amounts, _ := wallets.Load(address)
	var balance Coin
	for _, a := range amounts {
		if a >= 0 {
			balance += Coin(a)
		} else {
			balance -= Coin(-a)
		}
	}
	return balance
}

func Top(num int) []Transaction {
	var res []Transaction
	memPool.Range(func(k string, v Transaction) bool {
		if num == 0 {
			return false
		}
		res = append(res, v)
		num--
		return true
	})
	return res
}

func MarkIngested(ts []Transaction) {
	for _, t := range ts {
		memPool.Delete(t.ID)
		var transferSum int64
		for _, tf := range t.Transfers {
			amounts, _ := wallets.Load(tf.To)
			wallets.Store(tf.To, append(amounts, int64(tf.Amount)))
			transferSum += int64(tf.Amount)
		}
		amounts, _ := wallets.Load(t.From)
		wallets.Store(t.From, append(amounts, -transferSum))
	}
}
