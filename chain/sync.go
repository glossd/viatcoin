package chain

import (
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

		for _, tx := range block.Transactions {
			err := tx.Verify()
			if err != nil {
				return fmt.Errorf("invalid transaction found: %s, tx_id=%s, block_index=%d, block_hash=%s",
					err, tx.ID, i, block.HashString())
			}
			// todo what if the transaction was tempered with? corrupted peer can potentially write any amount for transfer
		}
		// whereas Bitcoin recontructs UTXO set, Viatcoin reconstruct the wallets.
		MarkIngested(block.Transactions)
		blockchain.Store(block.HashString(), block)
	}

	// todo synchronization must be constant.
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
