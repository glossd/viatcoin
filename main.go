package main

import (
	"github.com/glossd/viatcoin/chain"
	"github.com/glossd/viatcoin/miner"
)

func main() {
	go chain.RunAPI(8333)

	pk, err := chain.NewPrivateKey()
	if err != nil {
		panic(err)
	}
	miner.Start(miner.StartConfig{Pk: pk, ApiUrl: "localhost:8333"})
}
