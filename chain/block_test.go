package chain

import "testing"

func TestTargetToBits(t *testing.T) {
	var bits uint32 = 1e6
	if bits != targetToBits(bitsToTarget(bits)) {
		t.Error("didn't reverse")
	}
}

func TestDifficultyToBits(t *testing.T) {
	if DiffucltyToBits(1) != targetToBits(maxDifficultyTarget) {
		t.Error("didn't match")
	}

	difficutly := 1000012.0
	reversed := BitsToDifficutly(DiffucltyToBits(difficulty))
	if difficutly != reversed {
		t.Errorf("reversal didn't work, got=%v", reversed)
	}
}
