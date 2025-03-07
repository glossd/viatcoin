package chain

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"time"
)

var maxDifficultyTarget = new(big.Int).SetBytes([]byte{
	0x00, 0x00, 0x00, 0xFF, 0xFF, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
})

var firstTransaction = func() Transaction {
	// todo serialize signed transaction and deserialize it here
	pkBytes, err := hex.DecodeString("56b29cd95fb3ecc7e729e564d3af72e73d67e2d97d038d4937c6a487de282a0d")
	if err != nil {
		panic(err)
	}
	pk := PrivateKeyFromBytes(pkBytes)
	address := pk.PublicKey().Address(network)
	tx, err := NewTransaction("", 50*Viatcoin, address, address).Sign(pk)
	if err != nil {
		panic(err)
	}
	return tx
}()
var genesisBlock = Block{
	Version:              1,
	PreviousHash:         []byte{},
	MerkleRoot:           calcMerkelRoot([]Transaction{firstTransaction}),
	Timestamp:            1231006505,
	Nonce:                2083236893,
	DifficultyTargetBits: targetToBits(maxDifficultyTarget),
	Transactions:         []Transaction{firstTransaction},
}

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

func (b Block) Hash() []byte {
	return doubleSHA256(b.blockHeader())
}

func (b Block) DifficultyTarget() *big.Int {
	return bitsToTarget(b.DifficultyTargetBits)
}

func bitsToTarget(compact uint32) *big.Int {
	exponent := compact >> 24         // first byte
	coefficient := compact & 0xFFFFFF // 3 last bytes
	target := new(big.Int).SetUint64(uint64(coefficient))
	return target.Lsh(target, uint(8*(exponent-3)))
}

func BitsToDifficutly(compact uint32) float64 {
	div := new(big.Float).Quo(new(big.Float).SetInt(maxDifficultyTarget), new(big.Float).SetInt(bitsToTarget(compact)))
	res, _ := div.Float64()
	return res
}

// reverse of bitsToTarget
func targetToBits(target *big.Int) uint32 {
	size := uint32((target.BitLen() + 7) / 8) // Number of bytes required
	var compact uint32

	if size <= 3 {
		compact = uint32(target.Uint64() << (8 * (3 - size)))
	} else {
		tmp := new(big.Int).Rsh(target, uint(8*(size-3))) // Shift right to fit in 3 bytes
		compact = uint32(tmp.Uint64())
	}

	// Add exponent (size) as the first byte
	if compact&0x00800000 != 0 {
		compact >>= 8
		size++
	}

	compact |= size << 24
	return compact
}

func DiffucltyToBits(dif float64) uint32 {
	diffTargetFloat := new(big.Float).Quo(new(big.Float).SetInt(maxDifficultyTarget), big.NewFloat(dif))
	diffTarget := new(big.Int)
	diffTargetFloat.Int(diffTarget)
	return targetToBits(diffTarget)
}

func (b Block) Difficulty() float64 {
	div := new(big.Int).Div(maxDifficultyTarget, b.DifficultyTarget())
	res, _ := div.Float64()
	return res
}

func (b Block) Valid() bool {
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

func NewBlock(previousHash []byte, selected []Transaction, netDifficutlyBits uint32) Block {
	return Block{
		PreviousHash:         previousHash,
		Timestamp:            uint32(time.Now().Unix()),
		Transactions:         selected,
		MerkleRoot:           calcMerkelRoot(selected),
		DifficultyTargetBits: netDifficutlyBits,
	}
}
