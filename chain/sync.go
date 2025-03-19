package chain

import (
	"bytes"
	"fmt"
	"github.com/glossd/fetch"
	"math/big"
)

func Sync(apiUrls []string) error {
	// Longest Chain Rule
	// Multiple valid chains may exist at the same time, but one eventually will outgrow another.
	var longestChain []Block
	var longestChainTotalWork = new(big.Int)
	var longestChainUrl = ""
	for _, url := range apiUrls {
		// todo start working with the first chain
		blocks, err := downloadBlocks(url)
		if err != nil {
			return err
		}
		newTotalWork := TotalWork(blocks)
		if newTotalWork.Cmp(longestChainTotalWork) > 0 {
			clear(longestChain) // help gc
			longestChain = blocks
			longestChainTotalWork = newTotalWork
			longestChainUrl = url
		} else {
			clear(blocks) // help gc
		}
	}

	setBlockchain(longestChain)

	err := downloadMempool(longestChainUrl)
	if err != nil {
		return err
	}

	// todo constantly check how's the height of others and switch if any is longer
	return Stream(longestChainUrl, func(block Block) {
		persist(block)
	})
}

func TotalWork(blocks []Block) *big.Int {
	acc := new(big.Int)
	for _, block := range blocks {
		acc.Add(acc, block.Work())
	}
	return acc
}

func downloadBlocks(apiUrl string) ([]Block, error) {
	blocks, err := fetch.Get[[]Block](apiUrl + "/api/block?sort=asc&limit=-1")
	if err != nil {
		return nil, err
	}
	// verifying the peer isn't corrupted
	for i, block := range blocks {
		if !block.Valid() {
			return nil, fmt.Errorf("invalid block found: index=%d, hash=%s", i, block.HashString())
		}

		// verify chaining
		if i > 0 {
			if !bytes.Equal(block.PreviousHash, blocks[i-1].Hash()) {
				return nil, fmt.Errorf("corrupted blockchain: PreviousHash is wrong, index=%d", i)
			}
		}

		// verify integrity of transactions
		for _, tx := range block.Transactions {
			err := tx.Verify()
			if err != nil {
				return nil, fmt.Errorf("invalid transaction found: %s, tx_id=%s, block_index=%d, block_hash=%s",
					err, tx.ID, i, block.HashString())
			}
		}
		if !bytes.Equal(block.MerkleRoot, calcMerkelRoot(block.Transactions)) {
			return nil, fmt.Errorf("invlaid block, markle root: index=%d, hash=%s", i, block.HashString())
		}
	}
	return blocks, nil
}

func setBlockchain(blocks []Block) {
	for _, block := range blocks {
		// whereas Bitcoin recontructs UTXO set, Viatcoin reconstruct the wallets.
		persist(block)
	}
}

func downloadMempool(apiUrl string) error {
	txs, err := fetch.Get[[]Transaction](apiUrl+"/api/mempool?limit=-1", fetch.Config{})
	if err != nil {
		return err
	}
	for _, tx := range txs {
		err := Push(tx)
		if err != nil {
			return fmt.Errorf("failed to add to mempool: %s, tx_id=%s", err, tx.ID)
		}
	}
	return nil
}
