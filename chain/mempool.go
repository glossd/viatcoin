package chain

import (
	"fmt"
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

func markIngested(ts []Transaction) {
	for _, t := range ts {
		memPool.Delete(t.ID)
		var transferSum Coin
		for _, tf := range t.Transfers {
			deposit(tf.To, tf.Amount)
			transferSum += tf.Amount
		}
		withdraw(t.From, transferSum)
	}
}

func deposit(addr string, amount Coin) {
	amounts, _ := wallets.Load(addr)
	wallets.Store(addr, append(amounts, int64(amount)))
}

func withdraw(addr string, amount Coin) {
	amounts, _ := wallets.Load(addr)
	wallets.Store(addr, append(amounts, -int64(amount))) // fixme conversion
}

// In case a block gets reverted by a longer chain.
func markEgested(ts []Transaction) {
	for _, t := range ts {
		var transferSum Coin
		for _, tf := range t.Transfers {
			withdraw(tf.To, tf.Amount)
			transferSum += tf.Amount
		}
		deposit(t.From, transferSum)

		memPool.Store(t.ID, t)
	}
}
