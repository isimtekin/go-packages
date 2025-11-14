package cryptoutils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// RSA key generation

// GenerateRSAKeyPair generates an RSA key pair with the specified bit size.
// Recommended sizes: 2048, 3072, 4096
func GenerateRSAKeyPair(bits int) (*rsa.PrivateKey, error) {
	if bits < 2048 {
		return nil, errors.New("key size must be at least 2048 bits")
	}
	return rsa.GenerateKey(rand.Reader, bits)
}

// EncodeRSAPrivateKeyToPEM encodes an RSA private key to PEM format
func EncodeRSAPrivateKeyToPEM(key *rsa.PrivateKey) []byte {
	privBytes := x509.MarshalPKCS1PrivateKey(key)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})
	return privPEM
}

// EncodeRSAPublicKeyToPEM encodes an RSA public key to PEM format
func EncodeRSAPublicKeyToPEM(key *rsa.PublicKey) ([]byte, error) {
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

// DecodeRSAPrivateKeyFromPEM decodes an RSA private key from PEM format
func DecodeRSAPrivateKeyFromPEM(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// DecodeRSAPublicKeyFromPEM decodes an RSA public key from PEM format
func DecodeRSAPublicKeyFromPEM(pemData []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}
	return rsaPub, nil
}

// RSA-OAEP encryption/decryption

// EncryptRSAOAEP encrypts plaintext using RSA-OAEP with SHA-256
func EncryptRSAOAEP(publicKey *rsa.PublicKey, plaintext []byte) ([]byte, error) {
	hash := sha256.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, publicKey, plaintext, nil)
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

// DecryptRSAOAEP decrypts ciphertext using RSA-OAEP with SHA-256
func DecryptRSAOAEP(privateKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	hash := sha256.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, privateKey, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

// RSA-PSS signing/verification

// SignRSAPSS signs a message using RSA-PSS with SHA-256
func SignRSAPSS(privateKey *rsa.PrivateKey, message []byte) ([]byte, error) {
	hash := sha256.Sum256(message)
	signature, err := rsa.SignPSS(rand.Reader, privateKey, crypto.SHA256, hash[:], nil)
	if err != nil {
		return nil, err
	}
	return signature, nil
}

// VerifyRSAPSS verifies an RSA-PSS signature with SHA-256
func VerifyRSAPSS(publicKey *rsa.PublicKey, message, signature []byte) error {
	hash := sha256.Sum256(message)
	err := rsa.VerifyPSS(publicKey, crypto.SHA256, hash[:], signature, nil)
	return err
}
