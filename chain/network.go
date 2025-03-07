package chain

import (
	"fmt"
	"math"
)

type Net byte

const (
	Mainnet Net = 0x00
	Testnet Net = 0x6F
)

var network = Mainnet

func SetNetwork(net Net) {
	network = net
}

// todo adjust mining difficulty every 2,016 blocks (~2 weeks) to keep block times around 10 minutes.
var difficulty = 1.0

var originalMinerReward = 50 * Viatcoin

var blockchain = []Block{genesisBlock}

func GetLastBlock() Block {
	return blockchain[len(blockchain)-1]
}

func GetMinerReward() Coin {
	return originalMinerReward / Coin(math.Pow(2, float64(len(blockchain)/210_000)))
}

func GetDiffuctlyTargetBits() uint32 {
	return DiffucltyToBits(difficulty)
}

func Broadcast(b Block) error {
	if !b.Valid() {
		return fmt.Errorf("invalid block")
	}
	if len(b.Transactions) == 0 {
		return fmt.Errorf("block must have at least one coinbase transaction")
	}

	coinbase := b.Transactions[0]
	if err := coinbase.Verify(); err != nil {
		return fmt.Errorf("coinbase transaction is invalid: %s", err)
	}
	if coinbase.PreviousHash == "" {
		if coinbase.Balance != GetMinerReward() {
			return fmt.Errorf("new coinbase balance doesn't match miner reward")
		}
	} else {
		unspent, ok := GetUnspent(coinbase.PreviousHash)
		if !ok {
			return fmt.Errorf("previous hash of coinbase is not found")
		}
		if coinbase.Balance != unspent.Balance+GetMinerReward() {
			return fmt.Errorf("coinbase miner reward doesn't add up with previous transaction")
		}
	}

	for _, tx := range b.Transactions[1:] {
		verifyTx(tx)
	}

	MarkIngested(b.Transactions)
	blockchain = append(blockchain, b)
	return nil
}
