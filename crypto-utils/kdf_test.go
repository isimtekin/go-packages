package cryptoutils

import (
	"bytes"
	"testing"
)

func TestDerivePBKDF2SHA256(t *testing.T) {
	password := []byte("my-password")
	salt := []byte("random-salt-1234")
	iterations := 10000
	keyLen := 32

	key := DerivePBKDF2SHA256(password, salt, iterations, keyLen)

	if len(key) != keyLen {
		t.Errorf("Key length = %d, want %d", len(key), keyLen)
	}

	// Same inputs should produce same output
	key2 := DerivePBKDF2SHA256(password, salt, iterations, keyLen)
	if !bytes.Equal(key, key2) {
		t.Error("PBKDF2 is not deterministic")
	}
}

func TestDerivePBKDF2SHA512(t *testing.T) {
	password := []byte("my-password")
	salt := []byte("random-salt-1234")
	iterations := 10000
	keyLen := 64

	key := DerivePBKDF2SHA512(password, salt, iterations, keyLen)

	if len(key) != keyLen {
		t.Errorf("Key length = %d, want %d", len(key), keyLen)
	}

	// Same inputs should produce same output
	key2 := DerivePBKDF2SHA512(password, salt, iterations, keyLen)
	if !bytes.Equal(key, key2) {
		t.Error("PBKDF2 is not deterministic")
	}
}

func TestPBKDF2DifferentPasswords(t *testing.T) {
	password1 := []byte("password1")
	password2 := []byte("password2")
	salt := []byte("salt")
	iterations := 1000
	keyLen := 32

	key1 := DerivePBKDF2SHA256(password1, salt, iterations, keyLen)
	key2 := DerivePBKDF2SHA256(password2, salt, iterations, keyLen)

	if bytes.Equal(key1, key2) {
		t.Error("Different passwords produced same key")
	}
}

func TestPBKDF2DifferentSalts(t *testing.T) {
	password := []byte("password")
	salt1 := []byte("salt1")
	salt2 := []byte("salt2")
	iterations := 1000
	keyLen := 32

	key1 := DerivePBKDF2SHA256(password, salt1, iterations, keyLen)
	key2 := DerivePBKDF2SHA256(password, salt2, iterations, keyLen)

	if bytes.Equal(key1, key2) {
		t.Error("Different salts produced same key")
	}
}

func TestPBKDF2DifferentIterations(t *testing.T) {
	password := []byte("password")
	salt := []byte("salt")
	keyLen := 32

	key1 := DerivePBKDF2SHA256(password, salt, 1000, keyLen)
	key2 := DerivePBKDF2SHA256(password, salt, 2000, keyLen)

	if bytes.Equal(key1, key2) {
		t.Error("Different iteration counts produced same key")
	}
}

func TestPBKDF2DifferentKeyLengths(t *testing.T) {
	password := []byte("password")
	salt := []byte("salt")
	iterations := 1000

	key16 := DerivePBKDF2SHA256(password, salt, iterations, 16)
	key32 := DerivePBKDF2SHA256(password, salt, iterations, 32)
	key64 := DerivePBKDF2SHA256(password, salt, iterations, 64)

	if len(key16) != 16 {
		t.Errorf("16-byte key length = %d, want 16", len(key16))
	}
	if len(key32) != 32 {
		t.Errorf("32-byte key length = %d, want 32", len(key32))
	}
	if len(key64) != 64 {
		t.Errorf("64-byte key length = %d, want 64", len(key64))
	}

	// First 16 bytes of 32-byte key should match 16-byte key
	if !bytes.Equal(key16, key32[:16]) {
		t.Error("16-byte key doesn't match prefix of 32-byte key")
	}
}

func TestPBKDF2SHA256vsSHA512(t *testing.T) {
	password := []byte("password")
	salt := []byte("salt")
	iterations := 1000
	keyLen := 32

	key256 := DerivePBKDF2SHA256(password, salt, iterations, keyLen)
	key512 := DerivePBKDF2SHA512(password, salt, iterations, keyLen)

	// Different hash functions should produce different keys
	if bytes.Equal(key256, key512) {
		t.Error("SHA-256 and SHA-512 produced same key")
	}
}

func TestDeriveKeyFromPassword(t *testing.T) {
	password := []byte("user-password")
	salt, _ := GenerateRandomBytes(16)

	key := DeriveKeyFromPassword(password, salt)

	// Should return 32-byte key
	if len(key) != 32 {
		t.Errorf("Key length = %d, want 32", len(key))
	}

	// Should be deterministic
	key2 := DeriveKeyFromPassword(password, salt)
	if !bytes.Equal(key, key2) {
		t.Error("DeriveKeyFromPassword is not deterministic")
	}
}

func TestPBKDF2WithEmptyPassword(t *testing.T) {
	password := []byte("")
	salt := []byte("salt")
	iterations := 1000
	keyLen := 32

	key := DerivePBKDF2SHA256(password, salt, iterations, keyLen)

	if len(key) != keyLen {
		t.Errorf("Key length = %d, want %d", len(key), keyLen)
	}

	// Should still be deterministic
	key2 := DerivePBKDF2SHA256(password, salt, iterations, keyLen)
	if !bytes.Equal(key, key2) {
		t.Error("PBKDF2 with empty password is not deterministic")
	}
}

func TestPBKDF2WithEmptySalt(t *testing.T) {
	password := []byte("password")
	salt := []byte("")
	iterations := 1000
	keyLen := 32

	key := DerivePBKDF2SHA256(password, salt, iterations, keyLen)

	if len(key) != keyLen {
		t.Errorf("Key length = %d, want %d", len(key), keyLen)
	}
}

func TestPBKDF2LowIterations(t *testing.T) {
	password := []byte("password")
	salt := []byte("salt")
	keyLen := 32

	// Test with very low iteration count (not recommended for production)
	key := DerivePBKDF2SHA256(password, salt, 1, keyLen)

	if len(key) != keyLen {
		t.Errorf("Key length = %d, want %d", len(key), keyLen)
	}
}

func TestPBKDF2HighIterations(t *testing.T) {
	password := []byte("password")
	salt := []byte("salt")
	keyLen := 32

	// Test with recommended high iteration count
	key := DerivePBKDF2SHA256(password, salt, 210000, keyLen)

	if len(key) != keyLen {
		t.Errorf("Key length = %d, want %d", len(key), keyLen)
	}
}
