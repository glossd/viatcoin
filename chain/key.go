package chain

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"

	"github.com/decred/base58"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"golang.org/x/crypto/ripemd160"
)

type PrivateKey struct {
	key *secp256k1.PrivateKey
}

type PublicKey struct {
	key *secp256k1.PublicKey
}

func (p *PrivateKey) PublicKey() *PublicKey {
	if p == nil {
		return nil
	}
	var pub PublicKey
	pub.key = p.key.PubKey()
	return &pub
}

func (p *PrivateKey) Bytes() []byte {
	return p.key.Serialize()
}

func (p *PrivateKey) Sign(in []byte) ([]byte, error) {
	return p.key.ToECDSA().Sign(rand.Reader, in, nil)
}

func (p *PublicKey) Bytes() []byte {
	return p.key.SerializeCompressed()
}

func PrivateKeyFromBytes(serialized []byte) *PrivateKey {
	key := secp256k1.PrivKeyFromBytes(serialized)
	return &PrivateKey{key: key}
}

func PublicKeyFromBytes(serialized []byte) (*PublicKey, error) {
	res, err := secp256k1.ParsePubKey(serialized)
	if err != nil {
		return nil, err
	}
	return &PublicKey{key: res}, nil
}

func (p *PublicKey) Verify(txData []byte, signature []byte) bool {
	return ecdsa.VerifyASN1(p.key.ToECDSA(), txData, signature)
}

func (p *PublicKey) publicKeyHash() []byte {
	return doRipemd160(doSHA256(p.Bytes()))
}

func (p *PublicKey) Address(n Net) string {
	pkh := p.publicKeyHash()
	checksum := doubleSHA256(append([]byte{byte(n)}, pkh...))[:4]
	return base58.Encode(append(pkh, checksum...))
}

// as P2PKH
func (p *PublicKey) ScriptPubKey() string {
	// OP_DUP OP_HASH160 <pubKeyHash> OP_EQUALVERIFY OP_CHECKSIG
	return "76a914" + hex.EncodeToString(p.publicKeyHash()) + "88ac"
}

func doRipemd160(b []byte) []byte {
	hasher := ripemd160.New()
	hasher.Write(b)
	return hasher.Sum(nil)
}

// Generate private key with secp256k1
func NewPrivateKey() (*PrivateKey, error) {
	key, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}
	return &PrivateKey{key: key}, nil
}

func mustPrivKey() *PrivateKey {
	k, err := NewPrivateKey()
	if err != nil {
		panic(err)
	}
	return k
}
