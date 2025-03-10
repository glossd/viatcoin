package chain

import (
	"fmt"
	"math/big"
	"testing"
	"time"
)

var maxDifficultyTargetBits uint32 = 0x1D00FFFF

func TestTargetToBits(t *testing.T) {
	type testCase struct {
		I string
		E uint32
	}
	data := []testCase{
		{I: "00000000000404CB000000000000000000000000000000000000000000000000", E: 0x1b0404cb},
		{I: "00000000000000000005ae3af5b1628dc0000000000000000000000000000000", E: 0x1705ae3a}, // f5b1628dc is some noise
		{I: maxDifficultyTarget.Text(16), E: maxDifficultyTargetBits},
	}
	for _, c := range data {
		in, ok := new(big.Int).SetString(c.I, 16)
		if !ok {
			t.Error("wrong E ", c.E)
		}
		got := targetToBits(in)
		if got != c.E {
			t.Errorf("target didn't match expected %d, got=%d", c.E, got)
		}
	}
}

func TestBitsToTarget(t *testing.T) {
	type testCase struct {
		I uint32
		E string
	}
	data := []testCase{
		{I: 0x1b0404cb, E: "00000000000404CB000000000000000000000000000000000000000000000000"},
		{I: 0x1705dd01, E: "00000000000000000005dd010000000000000000000000000000000000000000"},
	}
	for _, c := range data {
		in := bitsToTarget(c.I)
		expected, ok := new(big.Int).SetString(c.E, 16)
		if !ok {
			t.Error("wrong E ", c.E)
		}
		if in.Cmp(expected) != 0 {
			t.Errorf("target didn't match expected %s, got=%s", c.E, in.Text(16))
		}
	}
}

func TestBitsToDifficulty(t *testing.T) {
	if BitsToDifficutly(maxDifficultyTargetBits).Cmp(new(big.Float).SetInt64(1)) != 0 {
		t.Error("max diff bits didn't match")
	}

	got := fmt.Sprintf("%0.11f", BitsToDifficutly(0x1b0404cb))
	if got != "16307.42093852398" {
		t.Errorf("expected 16307.42093852398, got=%s", got)
	}
}

func TestDifficultyToBits(t *testing.T) {
	if DiffucltyToBits(new(big.Float).SetInt64(1)) != maxDifficultyTargetBits {
		t.Error("max diff didn't match")
	}

	var dif int64 = 1000012
	difficutly := new(big.Float).SetInt64(dif)
	reversed := BitsToDifficutly(DiffucltyToBits(difficutly))
	reversedRounded, _ := reversed.Int64() // target doesn't have floating point
	if reversedRounded != dif {
		t.Errorf("reversal didn't work, got=%v", reversed)
	}
}

func TestAdjustDifficulty(t *testing.T) {
	t.Cleanup(func() {
		clear(blockchain)
	})
	const d = 1e-10
	diffic := new(big.Float).SetFloat64(d) // basically any block hash will do
	pk := mustPrivKey()
	ct, err := NewTransactionS(pk.PublicKey().Address(Mainnet), GetMinerReward()).Sign(pk)
	if err != nil {
		t.Fatal(err)
	}
	b := NewBlock(genesisBlock.PreviousHash, []Transaction{ct}, DiffucltyToBits(diffic))
	if !b.Valid() {
		t.Fatalf("block invalid: %v", b)
	}
	b.Timestamp = uint32(time.Now().Unix() + 1) // otherwise genesis will have the same timestamp
	err = doBroadcast(b, diffic, 2)
	if err != nil {
		t.Fatal(err)
	}
	if diffic.Cmp(new(big.Float).SetFloat64(d)) != 1 {
		t.Error("difficulty should've increased after one block: ", diffic.String())
	}
}
