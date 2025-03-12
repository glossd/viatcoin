package miner

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/glossd/fetch"
	"github.com/glossd/viatcoin/chain"
)

type StartConfig struct {
	Pk           *chain.PrivateKey // required
	Network      chain.Net         // defaults to Mainnet
	ApiUrl       string
}

func Start(cfg StartConfig) {
	if cfg.ApiUrl != "" {
		fetch.SetBaseURL(cfg.ApiUrl + "/api")
	}
	if cfg.Pk == nil {
		panic("private key isn't specified")
	}
	lb, err := fetch.Get[chain.Block]("/blocks/last")
	if err != nil {
		panic(err)	
	}
	txs, err := fetch.Get[[]chain.Transaction]("/mempool?limit=999")
	if err != nil {
		panic(err)	
	}

	difTargetBits, err := fetch.Get[uint32]("/difficulty/target/bits")
	if err != nil {
		panic(err)	
	}

	minerReward, err := fetch.Get[chain.Coin]("/reward")
	if err != nil {
		panic(err)	
	}

	pkAddress := cfg.Pk.PublicKey().Address(cfg.Network)
	coinbaseTx, err := chain.NewTransactionS(pkAddress, minerReward).Sign(cfg.Pk)
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
	_, err = fetch.Post[fetch.Empty]("/blocks", block)
	if err != nil {
		if strings.Contains(err.Error(), "invalid previous hash") {
			Start(cfg)
			return
		} else {
			panic("broadcasting valid block failed: " + err.Error())
		}
	} else {
		fmt.Printf("broadcasted block, earned %.2f Viatcoins\n Hash: %s\n Diff: %064s\n\n",
			minerReward.AsViatcoins(), block.HashString(), block.DifficultyTarget().Text(16))
	}
	Start(cfg)
}

func searchForValidBlock(last chain.Block, txs []chain.Transaction, difTarBits uint32) chain.Block {
	b := chain.NewBlock(last.Hash(), txs, difTarBits)
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
