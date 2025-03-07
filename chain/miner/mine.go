package miner

import (
	"fmt"
	"math"
	"time"

	"github.com/glossd/viatcoin/chain"
)

type StartConfig struct {
	Pk           *chain.PrivateKey // required
	Network      chain.Net         // defaults to Mainnet
	PreviousHash string            // optional if pk doesn't have any transactions
}

func Start(cfg StartConfig) {
	if cfg.Pk == nil {
		panic("private key isn't specified")
	}
	lb := chain.GetLastBlock()
	txs := chain.Top(999)
	difTargetBits := chain.GetDiffuctlyTargetBits()

	minerReward := chain.GetMinerReward()
	newBalance := minerReward
	if cfg.PreviousHash != "" {
		t, ok := chain.GetUnspent(cfg.PreviousHash)
		if !ok {
			panic("previous transaction is not unspent")
		}
		newBalance += t.Balance
	}

	pkAddress := cfg.Pk.PublicKey().Address(cfg.Network)
	coinbaseTx, err := chain.NewTransaction(cfg.PreviousHash, newBalance, pkAddress, pkAddress).Sign(cfg.Pk)
	if err != nil {
		panic("failed to sign coinbase transaction" + err.Error())
	}
	if len(txs) == 0 {
		txs = []chain.Transaction{coinbaseTx}
	} else {
		swap := txs[0]
		txs[0] = coinbaseTx
		txs = append(txs, swap)
	}

	block := searchForValidBlock(lb, txs, difTargetBits)
	err = chain.Broadcast(block)
	if err != nil {
		panic("broadcasting valid block failed: %s" + err.Error())
	} else {
		fmt.Printf("broadcasted block, earned %.2f Viatcoins\n", minerReward.AsViatcoins())
	}
	cfg.PreviousHash = coinbaseTx.Hash
	Start(cfg)
}

func searchForValidBlock(last chain.Block, txs []chain.Transaction, difTarBits uint32) chain.Block {
	b := chain.NewBlock(last.PreviousHash, txs, difTarBits)
	// todo add coinbase transaction
	n, ok := bruteForceNonce(b)
	if ok {
		b.Nonce = n
		return b
	} else {
		fmt.Println("nonce exhausted, changing timestamp")
		return searchForValidBlock(last, txs, difTarBits)
	}
}

func bruteForceNonce(b chain.Block) (uint32, bool) {
	var nonce uint32
	printTime := time.Now()
	for {
		b.Nonce = nonce
		if b.Valid() {
			return b.Nonce, true
		}
		if math.MaxUint32 == b.Nonce {
			return 0, false
		}
		nonce++
		now := time.Now()
		if now.Compare(printTime.Add(time.Minute)) > 0 {
			printTime = now
			fmt.Printf("still brute forcing... nonce=%d\n", nonce)
		}
	}
}
