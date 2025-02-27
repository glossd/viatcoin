package chain

import "testing"

func TestSignAndVerify(t *testing.T) {
	prv, err := NewPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	myBytes := []byte{1, 2, 3, 4}
	signature, err := prv.Sign(myBytes)
	if err != nil {
		t.Fatal(err)
	}
	if !(prv.PublicKey().Verify(myBytes, signature)) {
		t.Error("signature should've been verified")
	}
}
