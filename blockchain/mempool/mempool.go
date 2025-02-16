package mempool

import "github.com/glossd/viatcoin/blockchain"

var memPool = []blockchain.Transaction{}

func Push(t blockchain.Transaction) {
	memPool = append(memPool, t)
}

func Pop() (blockchain.Transaction, bool) {
	if len(memPool) == 0 {
		return blockchain.Transaction{}, false
	}
	t := memPool[0]
	memPool = memPool[1:]
	return t, true
}
