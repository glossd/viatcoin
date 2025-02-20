package chain

import "testing"

func TestTargetToBits(t *testing.T) {
	var bits uint32 = 1e6
	if bits != targetToBits(bitsToTarget(bits)) {
		t.Error("didn't reverse")
	}
}
