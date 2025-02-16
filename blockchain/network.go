package blockchain

import (
	"fmt"
	"math"
)

// todo adjust mining difficulty every 2,016 blocks (~2 weeks) to keep block times around 10 minutes.

var difficulty = 1.0

var originalMinerReward = 50 * Viatcoin

var blockchain []Block

func getMinerReward() Coin {
	return originalMinerReward / Coin(math.Pow(2, float64(len(blockchain)/210_000)))
}

func Broadcast(b Block) error {
	if !b.checkValidity() {
		return fmt.Errorf("invalid block")
	}
	// todo how does bitcoin blockchain checks that the block is correct
	blockchain = append(blockchain, b)
	return nil
}
