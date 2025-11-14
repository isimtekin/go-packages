package cryptoutils

import (
	"crypto/sha256"
	"crypto/sha512"
	"golang.org/x/crypto/pbkdf2"
)

// Key derivation functions

// DerivePBKDF2SHA256 derives a key using PBKDF2 with SHA-256
// iterations: recommended minimum 100,000 (higher for more security)
// keyLen: desired key length in bytes (e.g., 32 for AES-256)
func DerivePBKDF2SHA256(password, salt []byte, iterations, keyLen int) []byte {
	return pbkdf2.Key(password, salt, iterations, keyLen, sha256.New)
}

// DerivePBKDF2SHA512 derives a key using PBKDF2 with SHA-512
// iterations: recommended minimum 100,000 (higher for more security)
// keyLen: desired key length in bytes (e.g., 32 for AES-256)
func DerivePBKDF2SHA512(password, salt []byte, iterations, keyLen int) []byte {
	return pbkdf2.Key(password, salt, iterations, keyLen, sha512.New)
}

// DeriveKeyFromPassword is a convenience function that uses PBKDF2-SHA256
// with recommended defaults (210,000 iterations, 32-byte key for AES-256)
func DeriveKeyFromPassword(password, salt []byte) []byte {
	return DerivePBKDF2SHA256(password, salt, 210000, 32)
}
