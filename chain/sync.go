package chain

import (
	"bytes"
	"fmt"
	"github.com/glossd/fetch"
)

func Sync(apiUrl string) error {
	err := downloadBlocks(apiUrl)
	if err != nil {
		return err
	}
	err = downloadMempool(apiUrl)
	if err != nil {
		return err
	}
	// todo synchronization must be constant
	// todo Longest Chain Rule
	return nil
}

func downloadBlocks(apiUrl string) error {
	blocks, err := fetch.Get[[]Block](apiUrl + "/api/block?sort=asc&limit=-1")
	if err != nil {
		return err
	}
	// verifying the peer isn't corrupted
	for i, block := range blocks {
		if !block.Valid() {
			return fmt.Errorf("invalid block found: index=%d, hash=%s", i, block.HashString())
		}

		// verify chaining
		if i > 0 {
			if !bytes.Equal(block.PreviousHash, blocks[i-1].Hash()) {
				return fmt.Errorf("corrupted blockchain: PreviousHash is wrong, index=%d", i)
			}
		}

		// verify integrity of transactions
		for _, tx := range block.Transactions {
			err := tx.Verify()
			if err != nil {
				return fmt.Errorf("invalid transaction found: %s, tx_id=%s, block_index=%d, block_hash=%s",
					err, tx.ID, i, block.HashString())
			}
		}
		if !bytes.Equal(block.MerkleRoot, calcMerkelRoot(block.Transactions)) {
			return fmt.Errorf("invlaid block, markle root: index=%d, hash=%s", i, block.HashString())
		}

		// whereas Bitcoin recontructs UTXO set, Viatcoin reconstruct the wallets.
		MarkIngested(block.Transactions)
		blockchain.Store(block.HashString(), block)
	}
	return nil
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
