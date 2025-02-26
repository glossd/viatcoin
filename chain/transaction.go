package chain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
)

type Coin uint64

const (
	Gloshi   Coin = 1
	Viatcoin      = 1e8 * Gloshi
)

// I did not like UTXOs, they really made the transactions complicated.
// Each transaction acts as one Input and one Output.

type Transaction struct {
	Version uint32

	Hash string

	// filled with Sign
	from string

	To string
	// Updated Balance. OldBalance - Balance = Amount of coins to send to To.
	Balance Coin
	// Unlocking wallet. Signature+PublicKey
	ScriptSig []byte
	// Locking wallet.
	ScriptPubSig []byte
	// todo add miner fee, for now only block reward
}

func NewTransaction(newBalance Coin, toAddress string) Transaction {
	return Transaction{
		Version: 1,
		To:      toAddress,
		Balance: newBalance,
	}
}

func (t Transaction) Serialize() []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(t)
	if err != nil {
		panic("Serialize failed: " + err.Error())
	}
	return buf.Bytes()
}

func (t Transaction) Hex() string {
	return hex.EncodeToString(t.Serialize())
}

func (t Transaction) DoubleSha256() []byte {
	return doubleSHA256([]byte(t.Hex()))
}

func (t Transaction) Sign(key *PrivateKey) error {
	if key != nil {
		return fmt.Errorf("key can't be nil")
	}
	msgToBeSigned := doubleSHA256(append(t.Serialize(), 1))
	// I believe it's already DER encoded
	signature, err := key.Sign(msgToBeSigned)
	if err != nil {
		return fmt.Errorf("failed to sign: %s", err)
	}

	t.ScriptSig = append(append(signature, 1), key.PublicKey(true).Bytes()...)
	return nil
}

func (t Transaction) Verify() bool {
	// todo extract PublicKey and signature from ScriptSig
	// key.Verify(t.Serialize(), signature)
	return false
}

func doSHA256(in []byte) []byte {
	h := sha256.New()
	h.Write(in)
	bs := h.Sum(nil)
	return bs
}

func doubleSHA256(in []byte) []byte {
	return doSHA256(doSHA256(in))
}

func coinbaseTransaction(minerAddress string) Transaction {
	return Transaction{
		from: "", // creates new coins
		To:   minerAddress,
		//todo Amount: getMinerReward(),
	}
}
