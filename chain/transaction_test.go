package chain

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func TestProveTransactionOwnership(t *testing.T) {
	privKey := mustPrivKey()
	tx := NewTransaction(uuid.New().String(), 0, privKey.PublicKey().Address(Mainnet), mustPrivKey().PublicKey().Address(Mainnet))
	signedTx, err := tx.Sign(privKey)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(tx.Serialize(), signedTx.Serialize()) {
		t.Error("serialized bytes changed after signing")
	}
	err = signedTx.Verify()
	if err != nil {
		t.Error("tx isn't verified:", err)
	}
}

func TestCoinAsViatcoin(t *testing.T) {
	if 1.0 != Coin(1e8).AsViatcoins() {
		t.Error("expected 1")
	}
}
