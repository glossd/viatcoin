package miner

import (
	"fmt"
	"math"

	"github.com/glossd/viatcoin/chain"
)

func StartMining(pk *chain.PrivateKey) {
	lb := chain.GetLastBlock()
	txs := chain.Top(999)
	// todo add coinbase transaction
	// swap := txs[0]
	// chain.NewTransactionCoinbase()
	block := searchForValidBlock(lb, txs)
	err := chain.Broadcast(block)
	if err != nil {
		fmt.Printf("broadcasting valid block failed: %s\n", err)
	}
	StartMining(pk)
}

func searchForValidBlock(last chain.Block, txs []chain.Transaction) chain.Block {
	b := chain.NewBlock(last.PreviousHash, txs)
	// todo add coinbase transaction
	n, ok := bruteForceNonce(b)
	if ok {
		b.Nonce = n
		return b
	} else {
		// new timestamp will change the hashes
		return searchForValidBlock(last, txs)
	}
}

func bruteForceNonce(b chain.Block) (uint32, bool) {
	for {
		if b.Valid() {
			return b.Nonce, true
		}
		if math.MaxUint32 == b.Nonce {
			return 0, false
		}
	}
}
