package miner

import (
	"math"

	"github.com/glossd/viatcoin/chain"
)

type StartConfig struct {
	Pk           *chain.PrivateKey // required
	Network      chain.Net         // defaults to Mainnet
	PreviousHash string            // optional if pk doesn't have any transactions
}

func StartMining(cfg StartConfig) {
	if cfg.Pk == nil {
		panic("private key isn't specified")
	}
	lb := chain.GetLastBlock()
	txs := chain.Top(999)

	newBalance := chain.GetMinerReward()
	if cfg.PreviousHash != "" {
		t, ok := chain.GetUnspent(cfg.PreviousHash)
		if !ok {
			panic("previous transaction is not unspent")
		}
		newBalance += t.Balance
	}

	coinbaseTx, err := chain.NewTransaction(cfg.PreviousHash, newBalance, cfg.Pk.PublicKey().Address(cfg.Network)).Sign(cfg.Pk)
	if err != nil {
		panic("failed to sign coinbase transaction" + err.Error())
	}
	swap := txs[0]
	txs[0] = coinbaseTx
	txs = append(txs, swap)
	block := searchForValidBlock(lb, txs)
	err = chain.Broadcast(block)
	if err != nil {
		panic("broadcasting valid block failed: %s" + err.Error())
	}
	cfg.PreviousHash = coinbaseTx.Hash
	StartMining(cfg)
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
