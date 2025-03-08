package chain

import (
	"testing"

	"github.com/google/uuid"
)

func TestProveTransactionOwnership(t *testing.T) {
	privKey := mustPrivKey()
	tx := NewTransaction(uuid.New().String(), 0, mustPrivKey().PublicKey().Address(Mainnet))
	_, err := tx.Serialize()
	if err == nil {
		t.Error("should be allowed to serialize before signing")
	}
	signedTx, err := tx.Sign(privKey)
	if err != nil {
		t.Error(err)
	}
	err = signedTx.Verify()
	if err != nil {
		t.Error("tx isn't verified:", err)
	}
}

func TestCoinAsViatcoin(t *testing.T) {
	if Coin(1e8).AsViatcoins() != 1.0 {
		t.Error("expected 1")
	}
}
