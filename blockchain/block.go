package blockchain

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"time"
)

type Block struct {
	Version      uint32
	PreviousHash []byte

	// a hash summarizing all transactions in the block
	MerkleRoot []byte

	Timestamp uint32
	// Brute-forced
	Nonce uint32
	// Taken from the network.
	// DifficultyTarget is calculated from these bits. And Difficulty can be calculated from DifficultyTarget.
	DifficultyTargetBits uint32
	// Taken from the mempool
	Transactions []Transaction
}

// All fields are combined into an 80-byte block header.
func (b Block) blockHeader() []byte {
	buf := bytes.NewBuffer(make([]byte, 80))
	buf.Write(uib(b.Version))
	buf.Write(b.PreviousHash)
	buf.Write(b.MerkleRoot)
	buf.Write(uib(b.Timestamp))
	buf.Write(uib(b.DifficultyTargetBits))
	buf.Write(uib(b.Nonce))
	return buf.Bytes()
}

func (b Block) DifficultyTarget() *big.Int {
	exponent := b.DifficultyTargetBits >> 24         // first byte
	coefficient := b.DifficultyTargetBits & 0xFFFFFF // 3 last bytes
	target := new(big.Int).SetUint64(uint64(coefficient))
	return target.Lsh(target, uint(8*(exponent-3)))
}

func (b Block) checkValidity() bool {
	hash := doubleSHA256(b.blockHeader())
	return bi(hash).Cmp(b.DifficultyTarget()) <= 0
}

func uib(v uint32) []byte {
	a := make([]byte, 4)
	binary.LittleEndian.PutUint32(a, v)
	return a
}

func bi(v []byte) *big.Int {
	return new(big.Int).SetBytes(v)
}

type pair struct {
	first  []byte
	second []byte
}

func (p pair) DoubleSha256() []byte {
	return doubleSHA256(append(p.first, p.second...))
}

type shaable interface {
	DoubleSha256() []byte
}

func calcMerkelRoot[T shaable](list []T) []byte {
	if len(list) == 1 {
		return list[0].DoubleSha256()
	}
	if len(list) == 2 {
		return doubleSHA256(append(list[0].DoubleSha256(), list[1].DoubleSha256()...))
	}
	var newList []pair
	var firstInPair []byte
	for i, t := range list {
		if i%2 == 0 {
			firstInPair = t.DoubleSha256()
		} else {
			newList = append(newList, pair{first: firstInPair, second: t.DoubleSha256()})
		}
	}
	if len(list)%2 == 1 {
		newList = append(newList, pair{first: firstInPair, second: firstInPair})
	}
	return calcMerkelRoot(newList)
}

func NewBlock(previousHash []byte, selected []Transaction, nonce uint32) Block {
	return Block{
		PreviousHash: previousHash,
		Timestamp:    uint32(time.Now().Unix()),
		Transactions: selected,
		Nonce:        nonce,
		MerkleRoot:   calcMerkelRoot(selected),
	}
}
