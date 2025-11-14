package cryptoutils

import (
	"crypto/ecdh"
	"crypto/rand"
	"errors"
)

// ECDH key exchange

// GenerateECDHKeyPair generates an ECDH key pair using X25519 curve
func GenerateECDHKeyPair() (*ecdh.PrivateKey, error) {
	return ecdh.X25519().GenerateKey(rand.Reader)
}

// GenerateECDHKeyPairP256 generates an ECDH key pair using P-256 curve
func GenerateECDHKeyPairP256() (*ecdh.PrivateKey, error) {
	return ecdh.P256().GenerateKey(rand.Reader)
}

// GenerateECDHKeyPairP384 generates an ECDH key pair using P-384 curve
func GenerateECDHKeyPairP384() (*ecdh.PrivateKey, error) {
	return ecdh.P384().GenerateKey(rand.Reader)
}

// GenerateECDHKeyPairP521 generates an ECDH key pair using P-521 curve
func GenerateECDHKeyPairP521() (*ecdh.PrivateKey, error) {
	return ecdh.P521().GenerateKey(rand.Reader)
}

// DeriveSharedSecret derives a shared secret using ECDH
func DeriveSharedSecret(privateKey *ecdh.PrivateKey, publicKey *ecdh.PublicKey) ([]byte, error) {
	if privateKey.Curve() != publicKey.Curve() {
		return nil, errors.New("curve mismatch between private and public keys")
	}
	return privateKey.ECDH(publicKey)
}

// EncodeECDHPrivateKey encodes an ECDH private key to bytes
func EncodeECDHPrivateKey(key *ecdh.PrivateKey) []byte {
	return key.Bytes()
}

// EncodeECDHPublicKey encodes an ECDH public key to bytes
func EncodeECDHPublicKey(key *ecdh.PublicKey) []byte {
	return key.Bytes()
}

// DecodeECDHPrivateKeyX25519 decodes an X25519 private key from bytes
func DecodeECDHPrivateKeyX25519(data []byte) (*ecdh.PrivateKey, error) {
	return ecdh.X25519().NewPrivateKey(data)
}

// DecodeECDHPublicKeyX25519 decodes an X25519 public key from bytes
func DecodeECDHPublicKeyX25519(data []byte) (*ecdh.PublicKey, error) {
	return ecdh.X25519().NewPublicKey(data)
}

// DecodeECDHPrivateKeyP256 decodes a P-256 private key from bytes
func DecodeECDHPrivateKeyP256(data []byte) (*ecdh.PrivateKey, error) {
	return ecdh.P256().NewPrivateKey(data)
}

// DecodeECDHPublicKeyP256 decodes a P-256 public key from bytes
func DecodeECDHPublicKeyP256(data []byte) (*ecdh.PublicKey, error) {
	return ecdh.P256().NewPublicKey(data)
}

// DecodeECDHPrivateKeyP384 decodes a P-384 private key from bytes
func DecodeECDHPrivateKeyP384(data []byte) (*ecdh.PrivateKey, error) {
	return ecdh.P384().NewPrivateKey(data)
}

// DecodeECDHPublicKeyP384 decodes a P-384 public key from bytes
func DecodeECDHPublicKeyP384(data []byte) (*ecdh.PublicKey, error) {
	return ecdh.P384().NewPublicKey(data)
}

// DecodeECDHPrivateKeyP521 decodes a P-521 private key from bytes
func DecodeECDHPrivateKeyP521(data []byte) (*ecdh.PrivateKey, error) {
	return ecdh.P521().NewPrivateKey(data)
}

// DecodeECDHPublicKeyP521 decodes a P-521 public key from bytes
func DecodeECDHPublicKeyP521(data []byte) (*ecdh.PublicKey, error) {
	return ecdh.P521().NewPublicKey(data)
}
