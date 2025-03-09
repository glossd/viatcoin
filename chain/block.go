package chain

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"time"
)

var maxDifficultyTarget, _ = new(big.Int).SetString("00000000FFFF0000000000000000000000000000000000000000000000000000", 16)

var firstTransaction = func() Transaction {
	// todo serialize signed transaction and deserialize it here
	pkBytes, err := hex.DecodeString("56b29cd95fb3ecc7e729e564d3af72e73d67e2d97d038d4937c6a487de282a0d")
	if err != nil {
		panic(err)
	}
	pk := PrivateKeyFromBytes(pkBytes)
	address := pk.PublicKey().Address(network)
	tx, err := NewTransactionS(address, 50*Viatcoin).Sign(pk)
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
	buf.Write(uiLE(b.Version))
	buf.Write(b.PreviousHash)
	buf.Write(b.MerkleRoot)
	buf.Write(uiLE(b.Timestamp))
	buf.Write(uiLE(b.DifficultyTargetBits))
	buf.Write(uiLE(b.Nonce))
	return buf.Bytes()
}

func (b Block) Hash() []byte {
	return doubleSHA256(b.blockHeader())
}

func (b Block) HashString() string {
	return hex.EncodeToString(doubleSHA256(b.blockHeader()))
}

func (b Block) DifficultyTarget() *big.Int {
	return bitsToTarget(b.DifficultyTargetBits)
}

func bitsToTarget(compact uint32) *big.Int {
	bitsBytes := uiBE(compact)
	exponent := bitsBytes[0]
	coefficient := bitsBytes[1:]
	coefficientInt := new(big.Int).SetBytes(coefficient)
	shift := exponent-3 // this number to shift coefficient left.
	powered := new(big.Int).Exp(new(big.Int).SetInt64(256), new(big.Int).SetInt64(int64(shift)), nil)
	return new(big.Int).Mul(coefficientInt, powered)
}

func BitsToDifficutly(compact uint32) float64 {
	div := new(big.Float).Quo(new(big.Float).SetInt(maxDifficultyTarget), new(big.Float).SetInt(bitsToTarget(compact)))
	res, _ := div.Float64()
	return res
	// target := bitsToTarget(compact)
	// dif, _ := new(big.Int).Div(maxDifficultyTarget, target).Float64()
	// return dif
}

// reverse of bitsToTarget
func targetToBits(target *big.Int) uint32 {
	bs := target.Bytes()
	coefficient := bs[0:3]
	shift := len(bs[3:])
	if bs[0] >= 80 {
		coefficient = []byte{0, bs[0], bs[1]}
		shift = len(bs[2:])
	}
	exponent := shift + 3
	bitsBytes := append([]byte{byte(exponent)}, coefficient...)
	return BEui(bitsBytes)
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

// reversed order... 
func uiLE(v uint32) []byte {
	a := make([]byte, 4)
	binary.LittleEndian.PutUint32(a, v)
	return a
}

func uiBE(v uint32) []byte {
	a := make([]byte, 4)
	binary.BigEndian.PutUint32(a, v)
	return a
}

func BEui(b []byte) uint32 {
	return binary.BigEndian.Uint32(b)
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
