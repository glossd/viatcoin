package main

import (
	"github.com/glossd/viatcoin/chain"
	"github.com/glossd/viatcoin/chain/miner"
)

func main() {
	pk, err := chain.NewPrivateKey()
	if err != nil {
		panic(err)
	}
	miner.Start(miner.StartConfig{Pk: pk})
}
