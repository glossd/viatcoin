package chain

import (
	"github.com/google/uuid"
	"reflect"
	"testing"
)

func TestProveTransactionOwnership(t *testing.T) {
	privKey := mustPrivKey()
	tx := NewTransaction(uuid.New().String(), 0, mustPrivKey().PublicKey().Address(Mainnet))
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
