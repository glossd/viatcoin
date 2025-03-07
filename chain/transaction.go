package chain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/google/uuid"
)

type Coin uint64

const (
	Gloshi   Coin = 1
	Viatcoin      = 1e8 * Gloshi
)

func (c Coin) AsViatcoins() float64 {
	res, _ := new(big.Int).Div(new(big.Int).SetUint64(uint64(c)), new(big.Int).SetInt64(1e8)).Float64()
	return res
}

// I did not like UTXOs, they really made the transactions complicated.
// Each transaction acts as one Input and one Output.

type Transaction struct {
	Version uint32

	Hash         string
	PreviousHash string
	// filled with Sign
	From string

	To string
	// Updated Balance. OldBalance - Balance = Amount of coins to send to To.
	Balance Coin

	// ScriptSig is divided
	Signature []byte
	PublicKey []byte
	// todo add miner fee, for now only block reward
}

func NewTransaction(previousHash string, newBalance Coin, from, to string) Transaction {
	return Transaction{
		Version:      1,
		Hash:         uuid.New().String(),
		PreviousHash: previousHash,
		From:         from, // from could've been populated during singuture, but From needs to be present during serialization so that we know it wasn't tempered after.
		To:           to,
		Balance:      newBalance,
	}
}

func (t Transaction) Serialize() []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	copyT := t
	copyT.Signature = nil
	copyT.PublicKey = nil
	err := enc.Encode(copyT)
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

func (t Transaction) Sign(key *PrivateKey) (Transaction, error) {
	if key == nil {
		return t, fmt.Errorf("key can't be nil")
	}
	// did I need to doubleSha256 the transaction data?
	signature, err := key.Sign(t.Serialize())
	// I believe the signature is already DER encoded
	if err != nil {
		return t, fmt.Errorf("failed to sign: %s", err)
	}
	if t.From != key.PublicKey().Address(network) {
		return t, fmt.Errorf("private key address doesn't match the transaction from address.")
	}
	t.Signature = signature
	t.PublicKey = key.PublicKey().Bytes()

	return t, nil
}

func (t Transaction) Verify() error {
	pubKey, err := PublicKeyFromBytes(t.PublicKey)
	if err != nil {
		return fmt.Errorf("couldn't deserialize public key: %s", err)
	}
	pubKeyAddr := pubKey.Address(network)
	if pubKeyAddr != t.From {
		return fmt.Errorf("address of the public key '%s' didn't match transaction's address: %s", pubKeyAddr, t.From)
	}
	ok := pubKey.Verify(t.Serialize(), t.Signature)
	if !ok {
		return fmt.Errorf("failed to verify")
	}
	return nil
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
