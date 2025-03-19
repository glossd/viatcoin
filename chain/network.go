package chain

import (
	"fmt"
	"math"
	"math/big"
	"reflect"

	"github.com/glossd/viatcoin/chain/util"
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

// adjust mining difficulty every 2,016 blocks (~2 weeks) to keep block times around 10 minutes.
var difficulty = new(big.Float).SetFloat64(0.001)

const NumBlocksAdjust = 2016

var originalMinerReward = 50 * Viatcoin

var blockchain = util.SortedMap[string, Block]{}

func init() {
	blockchain.Store(genesisBlock.HashString(), genesisBlock)
}

func GetLastBlock() Block {
	return blockchain.Last()
}

func GetMinerReward() Coin {
	return originalMinerReward / Coin(math.Pow(2, float64(blockchain.Len()/210_000)))
}

func GetDiffuctlyTargetBits() uint32 {
	return DiffucltyToBits(difficulty)
}

func Broadcast(b Block) error {
	return doBroadcast(b, difficulty, NumBlocksAdjust)
}

func doBroadcast(b Block, diff *big.Float, numOfBlocksBeforeAdjust int) error {
	if !b.Valid() {
		return fmt.Errorf("invalid block")
	}
	if !reflect.DeepEqual(b.PreviousHash, blockchain.Last().Hash()) {
		return fmt.Errorf("invalid previous hash")
	}
	// todo check the timestamps
	if len(b.Transactions) == 0 {
		return fmt.Errorf("block must have at least one coinbase transaction")
	}

	coinbase := b.Transactions[0]
	if err := coinbase.Verify(); err != nil {
		return fmt.Errorf("coinbase transaction is invalid: %s", err)
	}
	if len(coinbase.Transfers) != 1 || coinbase.Transfers[0].Amount != GetMinerReward() {
		return fmt.Errorf("coinbase transfer amount doesn't match miner reward")
	}

	for _, tx := range b.Transactions[1:] {
		verifyTx(tx)
	}

	persist(b)

	if blockchain.Len()%numOfBlocksBeforeAdjust == 0 { // genesis block is hard-coded not broadcasted
		first := blockchain.LoadIndex(blockchain.Len() - numOfBlocksBeforeAdjust)
		last := blockchain.Last()
		actualTime := new(big.Float).SetInt64(int64(last.Timestamp - first.Timestamp))
		expectedTime := new(big.Float).SetInt64(10 * 60 * int64(numOfBlocksBeforeAdjust))
		diff.Mul(diff, new(big.Float).Quo(expectedTime, actualTime))
	}
	return nil
}

func persist(b Block) {
	MarkIngested(b.Transactions)
	blockchain.Store(b.HashString(), b)
}
