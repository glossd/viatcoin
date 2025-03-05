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

func Broadcast(b Block) error {
	if !b.Valid() {
		return fmt.Errorf("invalid block")
	}
	// todo validate the first coinbase transaction,
	// validate balances of the transactions
	if err := ExistsUnspent(b.Transactions); err != nil {
		return fmt.Errorf("mempool: %s", err)
	}
	Delete(b.Transactions)
	blockchain = append(blockchain, b)
	return nil
}
