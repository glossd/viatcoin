package chain

type privateKey struct{}

func (p privateKey) PublicKey() {
	// Elliptic Curve Multiplication
}

func (p privateKey) PublicAddress() {
	//  do
	// SHA-256 + RIPEMD-160 + Base58Chec
	// on PublicKey
}

// Generate private key with secp256k1 ECC
func NewPrivateKey() privateKey {
	// todo
}
