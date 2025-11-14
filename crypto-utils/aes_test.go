package cryptoutils

import (
	"bytes"
	"testing"
)

func TestAESGCM(t *testing.T) {
	tests := []struct {
		name      string
		keySize   int
		plaintext []byte
	}{
		{"AES-128", 16, []byte("Hello, World!")},
		{"AES-192", 24, []byte("Test message with 24-byte key")},
		{"AES-256", 32, []byte("Secure data encryption test")},
		{"Empty data", 32, []byte("")},
		{"Large data", 32, bytes.Repeat([]byte("x"), 10000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := GenerateRandomBytes(tt.keySize)
			if err != nil {
				t.Fatalf("Failed to generate key: %v", err)
			}

			// Encrypt
			ciphertext, err := EncryptAESGCM(key, tt.plaintext)
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}

			// Decrypt
			decrypted, err := DecryptAESGCM(key, ciphertext)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}

			if !bytes.Equal(decrypted, tt.plaintext) {
				t.Errorf("Decrypted data doesn't match original")
			}
		})
	}
}

func TestAESGCMInvalidKey(t *testing.T) {
	invalidKey := []byte("tooshort")
	plaintext := []byte("test")

	_, err := EncryptAESGCM(invalidKey, plaintext)
	if err != ErrInvalidKeySize {
		t.Errorf("Expected ErrInvalidKeySize, got: %v", err)
	}
}

func TestAESCBC(t *testing.T) {
	tests := []struct {
		name      string
		keySize   int
		plaintext []byte
	}{
		{"AES-128", 16, []byte("Hello, World!")},
		{"AES-192", 24, []byte("Test message with 24-byte key")},
		{"AES-256", 32, []byte("Secure data encryption test")},
		{"Short data", 32, []byte("Hi")},
		{"Large data", 32, bytes.Repeat([]byte("x"), 5000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := GenerateRandomBytes(tt.keySize)
			if err != nil {
				t.Fatalf("Failed to generate key: %v", err)
			}

			// Encrypt
			ciphertext, err := EncryptAESCBC(key, tt.plaintext)
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}

			// Decrypt
			decrypted, err := DecryptAESCBC(key, ciphertext)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}

			if !bytes.Equal(decrypted, tt.plaintext) {
				t.Errorf("Decrypted data doesn't match original")
			}
		})
	}
}

func TestAESCBCInvalidKey(t *testing.T) {
	invalidKey := []byte("tooshort")
	plaintext := []byte("test")

	_, err := EncryptAESCBC(invalidKey, plaintext)
	if err != ErrInvalidKeySize {
		t.Errorf("Expected ErrInvalidKeySize, got: %v", err)
	}
}

func TestPKCS7Padding(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		blockSize int
	}{
		{"Empty", []byte(""), 16},
		{"Exact block", []byte("1234567890123456"), 16},
		{"Needs padding", []byte("Hello"), 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			padded := pkcs7Pad(tt.data, tt.blockSize)
			unpadded, err := pkcs7Unpad(padded, tt.blockSize)
			if err != nil {
				t.Fatalf("Unpad failed: %v", err)
			}
			if !bytes.Equal(unpadded, tt.data) {
				t.Errorf("Unpadded data doesn't match original")
			}
		})
	}
}
