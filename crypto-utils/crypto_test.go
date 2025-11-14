package cryptoutils

import (
	"bytes"
	"testing"
)

// Integration tests combining multiple crypto operations

func TestRSAKeyPairGeneration(t *testing.T) {
	privateKey, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key pair: %v", err)
	}

	if privateKey.N.BitLen() < 2048 {
		t.Errorf("Generated key size is less than 2048 bits")
	}

	// Test PEM encoding/decoding
	privPEM := EncodeRSAPrivateKeyToPEM(privateKey)
	pubPEM, err := EncodeRSAPublicKeyToPEM(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to encode public key: %v", err)
	}

	decodedPriv, err := DecodeRSAPrivateKeyFromPEM(privPEM)
	if err != nil {
		t.Fatalf("Failed to decode private key: %v", err)
	}

	decodedPub, err := DecodeRSAPublicKeyFromPEM(pubPEM)
	if err != nil {
		t.Fatalf("Failed to decode public key: %v", err)
	}

	if !privateKey.Equal(decodedPriv) {
		t.Error("Decoded private key doesn't match original")
	}

	if !privateKey.PublicKey.Equal(decodedPub) {
		t.Error("Decoded public key doesn't match original")
	}
}

func TestRSAOAEP(t *testing.T) {
	privateKey, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	plaintext := []byte("Secret message for RSA encryption")

	// Encrypt
	ciphertext, err := EncryptRSAOAEP(&privateKey.PublicKey, plaintext)
	if err != nil {
		t.Fatalf("RSA encryption failed: %v", err)
	}

	// Decrypt
	decrypted, err := DecryptRSAOAEP(privateKey, ciphertext)
	if err != nil {
		t.Fatalf("RSA decryption failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypted message doesn't match original")
	}
}

func TestRSAPSS(t *testing.T) {
	privateKey, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	message := []byte("Message to sign")

	// Sign
	signature, err := SignRSAPSS(privateKey, message)
	if err != nil {
		t.Fatalf("RSA signing failed: %v", err)
	}

	// Verify
	err = VerifyRSAPSS(&privateKey.PublicKey, message, signature)
	if err != nil {
		t.Errorf("Signature verification failed: %v", err)
	}

	// Tamper with message
	tamperedMessage := []byte("Tampered message")
	err = VerifyRSAPSS(&privateKey.PublicKey, tamperedMessage, signature)
	if err == nil {
		t.Error("Signature verification should have failed for tampered message")
	}
}

func TestECDSASignVerify(t *testing.T) {
	privateKey, err := GenerateECDSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate ECDSA key: %v", err)
	}

	message := []byte("Message to sign with ECDSA")

	// Sign
	r, s, err := SignECDSA(privateKey, message)
	if err != nil {
		t.Fatalf("ECDSA signing failed: %v", err)
	}

	// Verify
	valid := VerifyECDSA(&privateKey.PublicKey, message, r, s)
	if !valid {
		t.Error("ECDSA signature verification failed")
	}

	// Test byte encoding
	sigBytes, err := SignECDSAToBytes(privateKey, message)
	if err != nil {
		t.Fatalf("Failed to sign to bytes: %v", err)
	}

	valid = VerifyECDSAFromBytes(&privateKey.PublicKey, message, sigBytes)
	if !valid {
		t.Error("ECDSA byte signature verification failed")
	}
}

func TestECDSAPEMEncoding(t *testing.T) {
	privateKey, err := GenerateECDSAKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate ECDSA key: %v", err)
	}

	// Encode to PEM
	privPEM, err := EncodeECDSAPrivateKeyToPEM(privateKey)
	if err != nil {
		t.Fatalf("Failed to encode private key: %v", err)
	}

	pubPEM, err := EncodeECDSAPublicKeyToPEM(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to encode public key: %v", err)
	}

	// Decode from PEM
	decodedPriv, err := DecodeECDSAPrivateKeyFromPEM(privPEM)
	if err != nil {
		t.Fatalf("Failed to decode private key: %v", err)
	}

	decodedPub, err := DecodeECDSAPublicKeyFromPEM(pubPEM)
	if err != nil {
		t.Fatalf("Failed to decode public key: %v", err)
	}

	if !privateKey.Equal(decodedPriv) {
		t.Error("Decoded private key doesn't match original")
	}

	if !privateKey.PublicKey.Equal(decodedPub) {
		t.Error("Decoded public key doesn't match original")
	}
}

func TestECDHKeyExchange(t *testing.T) {
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

	// Alice derives shared secret
	aliceShared, err := DeriveSharedSecret(alicePriv, bobPub)
	if err != nil {
		t.Fatalf("Alice failed to derive shared secret: %v", err)
	}

	// Bob derives shared secret
	bobShared, err := DeriveSharedSecret(bobPriv, alicePub)
	if err != nil {
		t.Fatalf("Bob failed to derive shared secret: %v", err)
	}

	// Shared secrets should match
	if !bytes.Equal(aliceShared, bobShared) {
		t.Error("Shared secrets don't match")
	}
}

func TestPasswordGeneration(t *testing.T) {
	tests := []struct {
		name string
		opts PasswordOptions
	}{
		{
			name: "Default",
			opts: DefaultPasswordOptions(),
		},
		{
			name: "Alphanumeric only",
			opts: PasswordOptions{
				Length:         12,
				IncludeLower:   true,
				IncludeUpper:   true,
				IncludeDigits:  true,
				IncludeSpecial: false,
			},
		},
		{
			name: "Digits only",
			opts: PasswordOptions{
				Length:        8,
				IncludeDigits: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := GeneratePassword(tt.opts)
			if err != nil {
				t.Fatalf("Password generation failed: %v", err)
			}

			if len(password) != tt.opts.Length {
				t.Errorf("Password length = %d, want %d", len(password), tt.opts.Length)
			}
		})
	}
}

func TestHashFunctions(t *testing.T) {
	data := []byte("Test data for hashing")

	// SHA-256
	hash256 := HashSHA256(data)
	if len(hash256) != 32 {
		t.Errorf("SHA-256 hash length = %d, want 32", len(hash256))
	}

	hash256Hex := HashSHA256Hex(data)
	if len(hash256Hex) != 64 {
		t.Errorf("SHA-256 hex length = %d, want 64", len(hash256Hex))
	}

	// SHA-512
	hash512 := HashSHA512(data)
	if len(hash512) != 64 {
		t.Errorf("SHA-512 hash length = %d, want 64", len(hash512))
	}

	// SHA-384
	hash384 := HashSHA384(data)
	if len(hash384) != 48 {
		t.Errorf("SHA-384 hash length = %d, want 48", len(hash384))
	}
}

func TestHMAC(t *testing.T) {
	key := []byte("secret-key")
	data := []byte("data to authenticate")

	// HMAC-SHA256
	mac256 := HMACSHA256(key, data)
	if len(mac256) != 32 {
		t.Errorf("HMAC-SHA256 length = %d, want 32", len(mac256))
	}

	if !VerifyHMACSHA256(key, data, mac256) {
		t.Error("HMAC-SHA256 verification failed")
	}

	// Tampered data should fail
	tamperedData := []byte("tampered data")
	if VerifyHMACSHA256(key, tamperedData, mac256) {
		t.Error("HMAC verification should have failed for tampered data")
	}

	// HMAC-SHA512
	mac512 := HMACSHA512(key, data)
	if len(mac512) != 64 {
		t.Errorf("HMAC-SHA512 length = %d, want 64", len(mac512))
	}

	if !VerifyHMACSHA512(key, data, mac512) {
		t.Error("HMAC-SHA512 verification failed")
	}
}

func TestPBKDF2(t *testing.T) {
	password := []byte("my-secure-password")
	salt, err := GenerateRandomBytes(16)
	if err != nil {
		t.Fatalf("Failed to generate salt: %v", err)
	}

	// Test SHA-256
	key256 := DerivePBKDF2SHA256(password, salt, 100000, 32)
	if len(key256) != 32 {
		t.Errorf("Derived key length = %d, want 32", len(key256))
	}

	// Test SHA-512
	key512 := DerivePBKDF2SHA512(password, salt, 100000, 32)
	if len(key512) != 32 {
		t.Errorf("Derived key length = %d, want 32", len(key512))
	}

	// Same inputs should produce same output
	key256_2 := DerivePBKDF2SHA256(password, salt, 100000, 32)
	if !bytes.Equal(key256, key256_2) {
		t.Error("PBKDF2 produced different keys for same inputs")
	}

	// Different salt should produce different key
	salt2, _ := GenerateRandomBytes(16)
	key256_3 := DerivePBKDF2SHA256(password, salt2, 100000, 32)
	if bytes.Equal(key256, key256_3) {
		t.Error("PBKDF2 produced same key for different salts")
	}
}

func TestRandomGeneration(t *testing.T) {
	// Test random bytes
	bytes1, err := GenerateRandomBytes(32)
	if err != nil {
		t.Fatalf("Failed to generate random bytes: %v", err)
	}
	if len(bytes1) != 32 {
		t.Errorf("Random bytes length = %d, want 32", len(bytes1))
	}

	bytes2, _ := GenerateRandomBytes(32)
	if bytes.Equal(bytes1, bytes2) {
		t.Error("Generated identical random bytes")
	}

	// Test random string
	str, err := GenerateRandomString(22)
	if err != nil {
		t.Fatalf("Failed to generate random string: %v", err)
	}
	if len(str) < 20 || len(str) > 22 {
		t.Errorf("Random string length = %d, want ~22", len(str))
	}

	// Test short ID
	id, err := GenerateShortID(16)
	if err != nil {
		t.Fatalf("Failed to generate short ID: %v", err)
	}
	if len(id) != 16 {
		t.Errorf("Short ID length = %d, want 16", len(id))
	}

	// Test secure token
	token, err := GenerateSecureToken()
	if err != nil {
		t.Fatalf("Failed to generate secure token: %v", err)
	}
	if len(token) != 43 {
		t.Errorf("Secure token length = %d, want 43", len(token))
	}

	// Test random int
	n, err := GenerateRandomInt(100)
	if err != nil {
		t.Fatalf("Failed to generate random int: %v", err)
	}
	if n < 0 || n >= 100 {
		t.Errorf("Random int = %d, want [0, 100)", n)
	}

	// Test random int range
	rangeN, err := GenerateRandomIntRange(50, 100)
	if err != nil {
		t.Fatalf("Failed to generate random int range: %v", err)
	}
	if rangeN < 50 || rangeN >= 100 {
		t.Errorf("Random int range = %d, want [50, 100)", rangeN)
	}
}
