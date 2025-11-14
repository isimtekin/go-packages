package cryptoutils

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
)

// Hash helpers

// HashSHA256 computes SHA-256 hash of data
func HashSHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// HashSHA256Hex computes SHA-256 hash and returns hex-encoded string
func HashSHA256Hex(data []byte) string {
	hash := HashSHA256(data)
	return hex.EncodeToString(hash)
}

// HashSHA512 computes SHA-512 hash of data
func HashSHA512(data []byte) []byte {
	hash := sha512.Sum512(data)
	return hash[:]
}

// HashSHA512Hex computes SHA-512 hash and returns hex-encoded string
func HashSHA512Hex(data []byte) string {
	hash := HashSHA512(data)
	return hex.EncodeToString(hash)
}

// HashSHA384 computes SHA-384 hash of data
func HashSHA384(data []byte) []byte {
	hash := sha512.Sum384(data)
	return hash[:]
}

// HashSHA384Hex computes SHA-384 hash and returns hex-encoded string
func HashSHA384Hex(data []byte) string {
	hash := HashSHA384(data)
	return hex.EncodeToString(hash)
}

// HMAC helpers

// HMACSHA256 computes HMAC-SHA256 of data with the given key
func HMACSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// HMACSHA256Hex computes HMAC-SHA256 and returns hex-encoded string
func HMACSHA256Hex(key, data []byte) string {
	mac := HMACSHA256(key, data)
	return hex.EncodeToString(mac)
}

// VerifyHMACSHA256 verifies an HMAC-SHA256 tag in constant time
func VerifyHMACSHA256(key, data, expectedMAC []byte) bool {
	mac := HMACSHA256(key, data)
	return hmac.Equal(mac, expectedMAC)
}

// HMACSHA512 computes HMAC-SHA512 of data with the given key
func HMACSHA512(key, data []byte) []byte {
	h := hmac.New(sha512.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// HMACSHA512Hex computes HMAC-SHA512 and returns hex-encoded string
func HMACSHA512Hex(key, data []byte) string {
	mac := HMACSHA512(key, data)
	return hex.EncodeToString(mac)
}

// VerifyHMACSHA512 verifies an HMAC-SHA512 tag in constant time
func VerifyHMACSHA512(key, data, expectedMAC []byte) bool {
	mac := HMACSHA512(key, data)
	return hmac.Equal(mac, expectedMAC)
}
