package cryptoutils

import (
	"crypto/elliptic"
	"testing"
)

func TestGenerateECDSAKeyPairWithDifferentCurves(t *testing.T) {
	curves := []struct {
		name  string
		curve elliptic.Curve
	}{
		{"P224", elliptic.P224()},
		{"P256", elliptic.P256()},
		{"P384", elliptic.P384()},
		{"P521", elliptic.P521()},
	}

	for _, tc := range curves {
		t.Run(tc.name, func(t *testing.T) {
			key, err := GenerateECDSAKeyPairWithCurve(tc.curve)
			if err != nil {
				t.Fatalf("Failed to generate key: %v", err)
			}

			if key.Curve != tc.curve {
				t.Errorf("Key curve doesn't match requested curve")
			}
		})
	}
}

func TestECDSASignVerifyWithDifferentCurves(t *testing.T) {
	curves := []elliptic.Curve{
		elliptic.P224(),
		elliptic.P256(),
		elliptic.P384(),
		elliptic.P521(),
	}

	message := []byte("Test message for different curves")

	for _, curve := range curves {
		t.Run(curve.Params().Name, func(t *testing.T) {
			key, err := GenerateECDSAKeyPairWithCurve(curve)
			if err != nil {
				t.Fatalf("Failed to generate key: %v", err)
			}

			r, s, err := SignECDSA(key, message)
			if err != nil {
				t.Fatalf("Signing failed: %v", err)
			}

			if !VerifyECDSA(&key.PublicKey, message, r, s) {
				t.Error("Signature verification failed")
			}
		})
	}
}

func TestECDSASignVerifyToBytes(t *testing.T) {
	key, err := GenerateECDSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	tests := []struct {
		name    string
		message []byte
	}{
		{"Empty", []byte("")},
		{"Short", []byte("Hi")},
		{"Medium", []byte("This is a test message")},
		{"Long", make([]byte, 10000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature, err := SignECDSAToBytes(key, tt.message)
			if err != nil {
				t.Fatalf("Signing failed: %v", err)
			}

			// Signature should be fixed size (2 * curve order bytes)
			expectedSize := 2 * ((key.Curve.Params().BitSize + 7) / 8)
			if len(signature) != expectedSize {
				t.Errorf("Signature size = %d, want %d", len(signature), expectedSize)
			}

			if !VerifyECDSAFromBytes(&key.PublicKey, tt.message, signature) {
				t.Error("Signature verification failed")
			}
		})
	}
}

func TestECDSAVerifyFromBytesInvalidSignature(t *testing.T) {
	key, _ := GenerateECDSAKeyPair()
	message := []byte("Message")

	tests := []struct {
		name      string
		signature []byte
	}{
		{"Too short", []byte{0x01, 0x02}},
		{"Too long", make([]byte, 100)},
		{"Empty", []byte{}},
		{"Wrong size", make([]byte, 63)}, // P-256 needs 64 bytes
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if VerifyECDSAFromBytes(&key.PublicKey, message, tt.signature) {
				t.Error("Expected verification to fail for invalid signature")
			}
		})
	}
}

func TestECDSAVerifyFailsWithTamperedMessage(t *testing.T) {
	key, _ := GenerateECDSAKeyPair()
	message := []byte("Original message")

	r, s, err := SignECDSA(key, message)
	if err != nil {
		t.Fatalf("Signing failed: %v", err)
	}

	tamperedMessage := []byte("Tampered message")
	if VerifyECDSA(&key.PublicKey, tamperedMessage, r, s) {
		t.Error("Expected verification to fail for tampered message")
	}
}

func TestECDSAVerifyFailsWithWrongKey(t *testing.T) {
	key1, _ := GenerateECDSAKeyPair()
	key2, _ := GenerateECDSAKeyPair()
	message := []byte("Message")

	r, s, err := SignECDSA(key1, message)
	if err != nil {
		t.Fatalf("Signing failed: %v", err)
	}

	if VerifyECDSA(&key2.PublicKey, message, r, s) {
		t.Error("Expected verification to fail with wrong key")
	}
}

func TestECDSAPEMEncodingDecoding(t *testing.T) {
	key, err := GenerateECDSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Test private key
	privPEM, err := EncodeECDSAPrivateKeyToPEM(key)
	if err != nil {
		t.Fatalf("Failed to encode private key: %v", err)
	}

	decodedPriv, err := DecodeECDSAPrivateKeyFromPEM(privPEM)
	if err != nil {
		t.Fatalf("Failed to decode private key: %v", err)
	}

	if !key.Equal(decodedPriv) {
		t.Error("Decoded private key doesn't match original")
	}

	// Test public key
	pubPEM, err := EncodeECDSAPublicKeyToPEM(&key.PublicKey)
	if err != nil {
		t.Fatalf("Failed to encode public key: %v", err)
	}

	decodedPub, err := DecodeECDSAPublicKeyFromPEM(pubPEM)
	if err != nil {
		t.Fatalf("Failed to decode public key: %v", err)
	}

	if !key.PublicKey.Equal(decodedPub) {
		t.Error("Decoded public key doesn't match original")
	}
}

func TestECDSAPEMDecodingErrors(t *testing.T) {
	tests := []struct {
		name    string
		pemData []byte
	}{
		{"Invalid PEM", []byte("not a pem")},
		{"Empty PEM", []byte("")},
		{"Wrong key type", []byte("-----BEGIN RSA PRIVATE KEY-----\nMIICXAIBAAKBgQC\n-----END RSA PRIVATE KEY-----")},
	}

	for _, tt := range tests {
		t.Run("Private_"+tt.name, func(t *testing.T) {
			_, err := DecodeECDSAPrivateKeyFromPEM(tt.pemData)
			if err == nil {
				t.Error("Expected error for invalid PEM")
			}
		})

		t.Run("Public_"+tt.name, func(t *testing.T) {
			_, err := DecodeECDSAPublicKeyFromPEM(tt.pemData)
			if err == nil {
				t.Error("Expected error for invalid PEM")
			}
		})
	}
}

func TestECDSAPublicKeyPEMWithWrongKeyType(t *testing.T) {
	// Generate RSA key and encode as PEM
	rsaKey, _ := GenerateRSAKeyPair(2048)
	rsaPubPEM, _ := EncodeRSAPublicKeyToPEM(&rsaKey.PublicKey)

	// Try to decode as ECDSA (should fail)
	_, err := DecodeECDSAPublicKeyFromPEM(rsaPubPEM)
	if err == nil {
		t.Error("Expected error when decoding RSA key as ECDSA")
	}
}
