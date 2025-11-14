package cryptoutils

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"math/big"
)

// Random generation utilities

// GenerateRandomBytes generates n cryptographically secure random bytes
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GenerateRandomString generates a random string of length n using base64 URL encoding
func GenerateRandomString(n int) (string, error) {
	// Calculate bytes needed for desired string length
	// Base64 encoding produces 4 characters for every 3 bytes
	bytesNeeded := (n * 3) / 4
	if (n*3)%4 != 0 {
		bytesNeeded++
	}

	b, err := GenerateRandomBytes(bytesNeeded)
	if err != nil {
		return "", err
	}

	// Use URL-safe base64 without padding
	str := base64.RawURLEncoding.EncodeToString(b)
	if len(str) > n {
		str = str[:n]
	}
	return str, nil
}

// GenerateShortID generates a short cryptographically secure ID
// Length: 8, 16, 22 (recommended), 32 characters
func GenerateShortID(length int) (string, error) {
	if length <= 0 {
		length = 22 // Default to 22 chars (similar to nanoid)
	}
	return GenerateRandomString(length)
}

// GenerateSecureToken generates a secure random token suitable for authentication
// Returns a 32-byte (256-bit) token encoded as base64 URL-safe string (43 chars)
func GenerateSecureToken() (string, error) {
	b, err := GenerateRandomBytes(32)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GenerateRandomInt generates a cryptographically secure random integer in range [0, max)
func GenerateRandomInt(max int64) (int64, error) {
	if max <= 0 {
		return 0, ErrInvalidRange
	}
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0, err
	}
	return n.Int64(), nil
}

// GenerateRandomIntRange generates a cryptographically secure random integer in range [min, max)
func GenerateRandomIntRange(min, max int64) (int64, error) {
	if min >= max {
		return 0, ErrInvalidRange
	}
	rangeSize := max - min
	n, err := GenerateRandomInt(rangeSize)
	if err != nil {
		return 0, err
	}
	return min + n, nil
}
