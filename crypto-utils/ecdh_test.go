package cryptoutils

import (
	"bytes"
	"testing"
)

func TestECDHKeyExchangeX25519(t *testing.T) {
	// Alice generates key pair
	alicePriv, err := GenerateECDHKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate Alice's key: %v", err)
	}
	alicePub := alicePriv.PublicKey()

	// Bob generates key pair
	bobPriv, err := GenerateECDHKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate Bob's key: %v", err)
	}
	bobPub := bobPriv.PublicKey()

	// Derive shared secrets
	aliceShared, err := DeriveSharedSecret(alicePriv, bobPub)
	if err != nil {
		t.Fatalf("Alice failed to derive shared secret: %v", err)
	}

	bobShared, err := DeriveSharedSecret(bobPriv, alicePub)
	if err != nil {
		t.Fatalf("Bob failed to derive shared secret: %v", err)
	}

	// Shared secrets should match
	if !bytes.Equal(aliceShared, bobShared) {
		t.Error("Shared secrets don't match")
	}

	// Shared secret should be 32 bytes for X25519
	if len(aliceShared) != 32 {
		t.Errorf("Shared secret length = %d, want 32", len(aliceShared))
	}
}

func TestECDHKeyExchangeP256(t *testing.T) {
	alicePriv, err := GenerateECDHKeyPairP256()
	if err != nil {
		t.Fatalf("Failed to generate Alice's key: %v", err)
	}
	alicePub := alicePriv.PublicKey()

	bobPriv, err := GenerateECDHKeyPairP256()
	if err != nil {
		t.Fatalf("Failed to generate Bob's key: %v", err)
	}
	bobPub := bobPriv.PublicKey()

	aliceShared, err := DeriveSharedSecret(alicePriv, bobPub)
	if err != nil {
		t.Fatalf("Alice failed to derive shared secret: %v", err)
	}

	bobShared, err := DeriveSharedSecret(bobPriv, alicePub)
	if err != nil {
		t.Fatalf("Bob failed to derive shared secret: %v", err)
	}

	if !bytes.Equal(aliceShared, bobShared) {
		t.Error("Shared secrets don't match")
	}
}

func TestECDHKeyExchangeP384(t *testing.T) {
	alicePriv, err := GenerateECDHKeyPairP384()
	if err != nil {
		t.Fatalf("Failed to generate Alice's key: %v", err)
	}
	alicePub := alicePriv.PublicKey()

	bobPriv, err := GenerateECDHKeyPairP384()
	if err != nil {
		t.Fatalf("Failed to generate Bob's key: %v", err)
	}
	bobPub := bobPriv.PublicKey()

	aliceShared, err := DeriveSharedSecret(alicePriv, bobPub)
	if err != nil {
		t.Fatalf("Alice failed to derive shared secret: %v", err)
	}

	bobShared, err := DeriveSharedSecret(bobPriv, alicePub)
	if err != nil {
		t.Fatalf("Bob failed to derive shared secret: %v", err)
	}

	if !bytes.Equal(aliceShared, bobShared) {
		t.Error("Shared secrets don't match")
	}
}

func TestECDHKeyExchangeP521(t *testing.T) {
	alicePriv, err := GenerateECDHKeyPairP521()
	if err != nil {
		t.Fatalf("Failed to generate Alice's key: %v", err)
	}
	alicePub := alicePriv.PublicKey()

	bobPriv, err := GenerateECDHKeyPairP521()
	if err != nil {
		t.Fatalf("Failed to generate Bob's key: %v", err)
	}
	bobPub := bobPriv.PublicKey()

	aliceShared, err := DeriveSharedSecret(alicePriv, bobPub)
	if err != nil {
		t.Fatalf("Alice failed to derive shared secret: %v", err)
	}

	bobShared, err := DeriveSharedSecret(bobPriv, alicePub)
	if err != nil {
		t.Fatalf("Bob failed to derive shared secret: %v", err)
	}

	if !bytes.Equal(aliceShared, bobShared) {
		t.Error("Shared secrets don't match")
	}
}

func TestECDHCurveMismatch(t *testing.T) {
	// X25519 key
	x25519Priv, _ := GenerateECDHKeyPair()

	// P-256 public key
	p256Priv, _ := GenerateECDHKeyPairP256()
	p256Pub := p256Priv.PublicKey()

	// Should fail due to curve mismatch
	_, err := DeriveSharedSecret(x25519Priv, p256Pub)
	if err == nil {
		t.Error("Expected error for curve mismatch")
	}
}

func TestECDHKeyEncodingX25519(t *testing.T) {
	priv, err := GenerateECDHKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	pub := priv.PublicKey()

	// Encode keys
	privBytes := EncodeECDHPrivateKey(priv)
	pubBytes := EncodeECDHPublicKey(pub)

	// Decode keys
	decodedPriv, err := DecodeECDHPrivateKeyX25519(privBytes)
	if err != nil {
		t.Fatalf("Failed to decode private key: %v", err)
	}

	decodedPub, err := DecodeECDHPublicKeyX25519(pubBytes)
	if err != nil {
		t.Fatalf("Failed to decode public key: %v", err)
	}

	// Verify keys work
	shared1, _ := DeriveSharedSecret(priv, decodedPub)
	shared2, _ := DeriveSharedSecret(decodedPriv, pub)

	if !bytes.Equal(shared1, shared2) {
		t.Error("Decoded keys don't work correctly")
	}
}

func TestECDHKeyEncodingP256(t *testing.T) {
	priv, err := GenerateECDHKeyPairP256()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	pub := priv.PublicKey()

	// Encode keys
	privBytes := EncodeECDHPrivateKey(priv)
	pubBytes := EncodeECDHPublicKey(pub)

	// Decode keys
	decodedPriv, err := DecodeECDHPrivateKeyP256(privBytes)
	if err != nil {
		t.Fatalf("Failed to decode private key: %v", err)
	}

	decodedPub, err := DecodeECDHPublicKeyP256(pubBytes)
	if err != nil {
		t.Fatalf("Failed to decode public key: %v", err)
	}

	// Verify keys work
	shared1, _ := DeriveSharedSecret(priv, decodedPub)
	shared2, _ := DeriveSharedSecret(decodedPriv, pub)

	if !bytes.Equal(shared1, shared2) {
		t.Error("Decoded keys don't work correctly")
	}
}

func TestECDHKeyEncodingP384(t *testing.T) {
	priv, err := GenerateECDHKeyPairP384()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	pub := priv.PublicKey()

	privBytes := EncodeECDHPrivateKey(priv)
	pubBytes := EncodeECDHPublicKey(pub)

	decodedPriv, err := DecodeECDHPrivateKeyP384(privBytes)
	if err != nil {
		t.Fatalf("Failed to decode private key: %v", err)
	}

	decodedPub, err := DecodeECDHPublicKeyP384(pubBytes)
	if err != nil {
		t.Fatalf("Failed to decode public key: %v", err)
	}

	shared1, _ := DeriveSharedSecret(priv, decodedPub)
	shared2, _ := DeriveSharedSecret(decodedPriv, pub)

	if !bytes.Equal(shared1, shared2) {
		t.Error("Decoded keys don't work correctly")
	}
}

func TestECDHKeyEncodingP521(t *testing.T) {
	priv, err := GenerateECDHKeyPairP521()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	pub := priv.PublicKey()

	privBytes := EncodeECDHPrivateKey(priv)
	pubBytes := EncodeECDHPublicKey(pub)

	decodedPriv, err := DecodeECDHPrivateKeyP521(privBytes)
	if err != nil {
		t.Fatalf("Failed to decode private key: %v", err)
	}

	decodedPub, err := DecodeECDHPublicKeyP521(pubBytes)
	if err != nil {
		t.Fatalf("Failed to decode public key: %v", err)
	}

	shared1, _ := DeriveSharedSecret(priv, decodedPub)
	shared2, _ := DeriveSharedSecret(decodedPriv, pub)

	if !bytes.Equal(shared1, shared2) {
		t.Error("Decoded keys don't work correctly")
	}
}

func TestECDHInvalidKeyDecoding(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"Empty", []byte{}},
		{"Too short", []byte{0x01, 0x02}},
		{"Invalid data", make([]byte, 32)}, // All zeros
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeECDHPrivateKeyX25519(tt.data)
			if err == nil && len(tt.data) != 32 {
				t.Error("Expected error for invalid key data")
			}

			_, err = DecodeECDHPublicKeyX25519(tt.data)
			if err == nil && len(tt.data) != 32 {
				t.Error("Expected error for invalid key data")
			}
		})
	}
}
