package cryptoutils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"math/big"
)

// ECDSA key generation

// GenerateECDSAKeyPair generates an ECDSA key pair using P-256 curve
func GenerateECDSAKeyPair() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// GenerateECDSAKeyPairWithCurve generates an ECDSA key pair with specified curve
// Supported curves: P224, P256, P384, P521
func GenerateECDSAKeyPairWithCurve(curve elliptic.Curve) (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(curve, rand.Reader)
}

// EncodeECDSAPrivateKeyToPEM encodes an ECDSA private key to PEM format
func EncodeECDSAPrivateKeyToPEM(key *ecdsa.PrivateKey) ([]byte, error) {
	privBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, err
	}
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privBytes,
	})
	return privPEM, nil
}

// EncodeECDSAPublicKeyToPEM encodes an ECDSA public key to PEM format
func EncodeECDSAPublicKeyToPEM(key *ecdsa.PublicKey) ([]byte, error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, err
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})
	return pubPEM, nil
}

// DecodeECDSAPrivateKeyFromPEM decodes an ECDSA private key from PEM format
func DecodeECDSAPrivateKeyFromPEM(pemData []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	return x509.ParseECPrivateKey(block.Bytes)
}

// DecodeECDSAPublicKeyFromPEM decodes an ECDSA public key from PEM format
func DecodeECDSAPublicKeyFromPEM(pemData []byte) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("not an ECDSA public key")
	}
	return ecdsaPub, nil
}

// ECDSA signing/verification

// ECDSASignature represents an ECDSA signature
type ECDSASignature struct {
	R, S *big.Int
}

// SignECDSA signs a message using ECDSA with SHA-256
func SignECDSA(privateKey *ecdsa.PrivateKey, message []byte) (r, s *big.Int, err error) {
	hash := sha256.Sum256(message)
	return ecdsa.Sign(rand.Reader, privateKey, hash[:])
}

// VerifyECDSA verifies an ECDSA signature with SHA-256
func VerifyECDSA(publicKey *ecdsa.PublicKey, message []byte, r, s *big.Int) bool {
	hash := sha256.Sum256(message)
	return ecdsa.Verify(publicKey, hash[:], r, s)
}

// SignECDSAToBytes signs a message and returns the signature as bytes (R || S)
func SignECDSAToBytes(privateKey *ecdsa.PrivateKey, message []byte) ([]byte, error) {
	r, s, err := SignECDSA(privateKey, message)
	if err != nil {
		return nil, err
	}

	// Concatenate R and S as fixed-size bytes
	curveOrderByteSize := (privateKey.Curve.Params().BitSize + 7) / 8
	rBytes := r.Bytes()
	sBytes := s.Bytes()

	signature := make([]byte, 2*curveOrderByteSize)
	copy(signature[curveOrderByteSize-len(rBytes):], rBytes)
	copy(signature[2*curveOrderByteSize-len(sBytes):], sBytes)

	return signature, nil
}

// VerifyECDSAFromBytes verifies an ECDSA signature from bytes (R || S)
func VerifyECDSAFromBytes(publicKey *ecdsa.PublicKey, message, signature []byte) bool {
	curveOrderByteSize := (publicKey.Curve.Params().BitSize + 7) / 8
	if len(signature) != 2*curveOrderByteSize {
		return false
	}

	r := new(big.Int).SetBytes(signature[:curveOrderByteSize])
	s := new(big.Int).SetBytes(signature[curveOrderByteSize:])

	return VerifyECDSA(publicKey, message, r, s)
}
