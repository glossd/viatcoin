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

// todo adjust mining difficulty every 2,016 blocks (~2 weeks) to keep block times around 10 minutes.

var difficulty = 1.0

var originalMinerReward = 50 * Viatcoin

var blockchain []Block

func getMinerReward() Coin {
	return originalMinerReward / Coin(math.Pow(2, float64(len(blockchain)/210_000)))
}

func Broadcast(b Block) error {
	// todo check that transactions exists in the mempool
	if !b.checkValidity() {
		return fmt.Errorf("invalid block")
	}
	blockchain = append(blockchain, b)
	return nil
}
