package chain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
)

type Coin uint64

const (
	Gloshi   Coin = 1
	Viatcoin      = 1e8 * Gloshi
)

// I did not like UTXOs, they really made the transactions complicated.

type Transaction struct {
	Hash   string
	From   string
	To     string
	Amount Coin
	// todo add miner fee, for now only block reward
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

func doSHA256(in []byte) []byte {
	h := sha256.New()
	h.Write(in)
	bs := h.Sum(nil)
	return bs
}

func doubleSHA256(in []byte) []byte {
	return doSHA256(doSHA256(in))
}

type Transfer struct {
	Address string
	Amount  Coin
}

func coinbaseTransaction(minerAddress string) Transaction {
	return Transaction{
		From:   "", // creates new coins
		To:     minerAddress,
		Amount: getMinerReward(),
	}
}
