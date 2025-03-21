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
	res, _ := new(big.Float).Quo(new(big.Float).SetUint64(uint64(c)), new(big.Float).SetInt64(1e8)).Float64()
	return res
}

// I did not like UTXOs, they really made the transactions complicated.
// Each transaction is address-based.

type Transaction struct {
	Version uint32
	ID      string
	// filled with Sign
	From string

	Transfers []Transfer

	// ScriptSig is divided
	Signature []byte
	PublicKey []byte
	// todo add miner fee, for now only block reward
}

type Transfer struct {
	To     string
	Amount Coin
}

func NewTransactionS(to string, amount Coin) Transaction {
	return Transaction{
		Version:   1,
		ID:        uuid.New().String(),
		Transfers: []Transfer{{To: to, Amount: amount}},
	}
}

func NewTransaction(to []Transfer) Transaction {
	return Transaction{
		Version:   1,
		ID:        uuid.New().String(),
		Transfers: to,
	}
}

func (t Transaction) Serialize() ([]byte, error) {
	if t.From == "" {
		return nil, fmt.Errorf("can't serialize before signing")
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	copyT := t
	copyT.Signature = nil
	copyT.PublicKey = nil
	err := enc.Encode(copyT)
	if err != nil {
		return nil, fmt.Errorf("serialization: %s", err)
	}
	return buf.Bytes(), nil
}

func (t Transaction) DoubleSha256() []byte {
	ser, err := t.Serialize()
	if err != nil {
		return nil
	}
	return doubleSHA256([]byte(hex.EncodeToString(ser)))
}

func (t Transaction) Sign(key *PrivateKey) (Transaction, error) {
	if key == nil {
		return t, fmt.Errorf("key can't be nil")
	}
	// from is populated before serialization, so that we know it wasn't tempered after verification.
	t.From = key.PublicKey().Address(network)
	ser, err := t.Serialize()
	if err != nil {
		return t, err
	}
	// did I need to doubleSha256 the transaction data?
	signature, err := key.Sign(ser)
	// I believe the signature is already DER encoded
	if err != nil {
		return t, fmt.Errorf("failed to sign: %s", err)
	}
	if t.From != key.PublicKey().Address(network) {
		return t, fmt.Errorf("private key address doesn't match the transaction from address")
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
	ser, err := t.Serialize()
	if err != nil {
		return err
	}
	ok := pubKey.Verify(ser, t.Signature)
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
