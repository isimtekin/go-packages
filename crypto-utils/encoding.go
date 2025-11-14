package cryptoutils

import (
	"encoding/base64"
)

// Base64 encoding/decoding helpers

// EncodeBase64 encodes data to standard base64 string
func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeBase64 decodes a standard base64 string
func DecodeBase64(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

// EncodeBase64URL encodes data to URL-safe base64 string (with padding)
func EncodeBase64URL(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

// DecodeBase64URL decodes a URL-safe base64 string (with padding)
func DecodeBase64URL(encoded string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(encoded)
}

// EncodeBase64RawURL encodes data to URL-safe base64 string without padding
func EncodeBase64RawURL(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// DecodeBase64RawURL decodes a URL-safe base64 string without padding
func DecodeBase64RawURL(encoded string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(encoded)
}

// EncodeBase64Raw encodes data to standard base64 string without padding
func EncodeBase64Raw(data []byte) string {
	return base64.RawStdEncoding.EncodeToString(data)
}

// DecodeBase64Raw decodes a standard base64 string without padding
func DecodeBase64Raw(encoded string) ([]byte, error) {
	return base64.RawStdEncoding.DecodeString(encoded)
}
