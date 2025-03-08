package chain

import (
	"fmt"
	"math/big"
	"testing"
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
	if BitsToDifficutly(maxDifficultyTargetBits) != 1.0 {
		t.Error("max diff bits didn't match")
	}

	got := fmt.Sprintf("%0.11f", BitsToDifficutly(0x1b0404cb))
	if got != "16307.42093852398" {
		t.Errorf("expected 16307.42093852398, got=%s", got)
	}
}

func TestDifficultyToBits(t *testing.T) {
	if DiffucltyToBits(1) != maxDifficultyTargetBits {
		t.Error("max diff didn't match")
	}

	difficutly := 1000012.0
	reversed := BitsToDifficutly(DiffucltyToBits(difficutly))
	if int(difficutly) != int(reversed) { // target doesn't have floating point
		t.Errorf("reversal didn't work, got=%v", int(reversed))
	}
}
