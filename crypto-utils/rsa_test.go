package cryptoutils

import (
	"bytes"
	"testing"
)

func TestGenerateRSAKeyPair(t *testing.T) {
	tests := []struct {
		name    string
		bits    int
		wantErr bool
	}{
		{"2048 bits", 2048, false},
		{"3072 bits", 3072, false},
		{"4096 bits", 4096, false},
		{"Too small", 1024, true},
		{"Invalid", 512, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := GenerateRSAKeyPair(tt.bits)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateRSAKeyPair() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && key.N.BitLen() < tt.bits {
				t.Errorf("Key size = %d bits, want >= %d", key.N.BitLen(), tt.bits)
			}
		})
	}
}

func TestRSAPEMEncoding(t *testing.T) {
	privateKey, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Test private key encoding/decoding
	privPEM := EncodeRSAPrivateKeyToPEM(privateKey)
	if len(privPEM) == 0 {
		t.Error("Private key PEM is empty")
	}

	decodedPriv, err := DecodeRSAPrivateKeyFromPEM(privPEM)
	if err != nil {
		t.Fatalf("Failed to decode private key: %v", err)
	}

	if !privateKey.Equal(decodedPriv) {
		t.Error("Decoded private key doesn't match original")
	}

	// Test public key encoding/decoding
	pubPEM, err := EncodeRSAPublicKeyToPEM(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to encode public key: %v", err)
	}

	if len(pubPEM) == 0 {
		t.Error("Public key PEM is empty")
	}

	decodedPub, err := DecodeRSAPublicKeyFromPEM(pubPEM)
	if err != nil {
		t.Fatalf("Failed to decode public key: %v", err)
	}

	if !privateKey.PublicKey.Equal(decodedPub) {
		t.Error("Decoded public key doesn't match original")
	}
}

func TestRSAPEMDecodingErrors(t *testing.T) {
	tests := []struct {
		name    string
		pemData []byte
	}{
		{"Invalid PEM", []byte("not a pem")},
		{"Empty PEM", []byte("")},
		{"Corrupted PEM", []byte("-----BEGIN RSA PRIVATE KEY-----\ngarbage\n-----END RSA PRIVATE KEY-----")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeRSAPrivateKeyFromPEM(tt.pemData)
			if err == nil {
				t.Error("Expected error for invalid PEM, got nil")
			}
		})
	}
}

func TestRSAOAEPEncryptDecrypt(t *testing.T) {
	privateKey, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	tests := []struct {
		name      string
		plaintext []byte
	}{
		{"Short message", []byte("Hello")},
		{"Medium message", []byte("This is a test message for RSA encryption")},
		{"Empty message", []byte("")},
		{"Binary data", []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := EncryptRSAOAEP(&privateKey.PublicKey, tt.plaintext)
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}

			decrypted, err := DecryptRSAOAEP(privateKey, ciphertext)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}

			if !bytes.Equal(decrypted, tt.plaintext) {
				t.Errorf("Decrypted = %v, want %v", decrypted, tt.plaintext)
			}
		})
	}
}

func TestRSAOAEPWithWrongKey(t *testing.T) {
	privateKey1, _ := GenerateRSAKeyPair(2048)
	privateKey2, _ := GenerateRSAKeyPair(2048)

	plaintext := []byte("Secret message")

	// Encrypt with key1
	ciphertext, err := EncryptRSAOAEP(&privateKey1.PublicKey, plaintext)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Try to decrypt with key2 (should fail)
	_, err = DecryptRSAOAEP(privateKey2, ciphertext)
	if err == nil {
		t.Error("Expected decryption to fail with wrong key")
	}
}

func TestRSAPSSSignVerify(t *testing.T) {
	privateKey, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	tests := []struct {
		name    string
		message []byte
	}{
		{"Short message", []byte("Sign me")},
		{"Long message", bytes.Repeat([]byte("x"), 1000)},
		{"Empty message", []byte("")},
		{"Binary data", []byte{0x00, 0xFF, 0x7F, 0x80}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature, err := SignRSAPSS(privateKey, tt.message)
			if err != nil {
				t.Fatalf("Signing failed: %v", err)
			}

			err = VerifyRSAPSS(&privateKey.PublicKey, tt.message, signature)
			if err != nil {
				t.Errorf("Verification failed: %v", err)
			}
		})
	}
}

func TestRSAPSSVerifyFailsWithTamperedMessage(t *testing.T) {
	privateKey, _ := GenerateRSAKeyPair(2048)
	message := []byte("Original message")

	signature, err := SignRSAPSS(privateKey, message)
	if err != nil {
		t.Fatalf("Signing failed: %v", err)
	}

	// Tamper with message
	tamperedMessage := []byte("Tampered message")
	err = VerifyRSAPSS(&privateKey.PublicKey, tamperedMessage, signature)
	if err == nil {
		t.Error("Expected verification to fail for tampered message")
	}
}

func TestRSAPSSVerifyFailsWithTamperedSignature(t *testing.T) {
	privateKey, _ := GenerateRSAKeyPair(2048)
	message := []byte("Message")

	signature, err := SignRSAPSS(privateKey, message)
	if err != nil {
		t.Fatalf("Signing failed: %v", err)
	}

	// Tamper with signature
	signature[0] ^= 0xFF

	err = VerifyRSAPSS(&privateKey.PublicKey, message, signature)
	if err == nil {
		t.Error("Expected verification to fail for tampered signature")
	}
}

func TestRSAPSSVerifyFailsWithWrongKey(t *testing.T) {
	privateKey1, _ := GenerateRSAKeyPair(2048)
	privateKey2, _ := GenerateRSAKeyPair(2048)
	message := []byte("Message")

	// Sign with key1
	signature, err := SignRSAPSS(privateKey1, message)
	if err != nil {
		t.Fatalf("Signing failed: %v", err)
	}

	// Verify with key2 (should fail)
	err = VerifyRSAPSS(&privateKey2.PublicKey, message, signature)
	if err == nil {
		t.Error("Expected verification to fail with wrong key")
	}
}
