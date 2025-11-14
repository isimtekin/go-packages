package cryptoutils

import "errors"

// Common errors
var (
	// Encryption/Decryption errors
	ErrInvalidKeySize   = errors.New("invalid key size")
	ErrInvalidBlockSize = errors.New("invalid block size")
	ErrInvalidPadding   = errors.New("invalid padding")
	ErrDecryptionFailed = errors.New("decryption failed")
	ErrEncryptionFailed = errors.New("encryption failed")

	// Key generation errors
	ErrKeyGenerationFailed = errors.New("key generation failed")
	ErrInvalidKeyFormat    = errors.New("invalid key format")
	ErrInvalidCurve        = errors.New("invalid curve")

	// Signature errors
	ErrSignatureFailed       = errors.New("signature generation failed")
	ErrSignatureVerification = errors.New("signature verification failed")
	ErrInvalidSignature      = errors.New("invalid signature")

	// Random generation errors
	ErrInsufficientRandomness = errors.New("insufficient randomness")
	ErrInvalidRange           = errors.New("invalid range")

	// Encoding errors
	ErrInvalidEncoding = errors.New("invalid encoding")
	ErrDecodeFailed    = errors.New("decode failed")
)
