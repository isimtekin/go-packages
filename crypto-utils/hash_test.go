package cryptoutils

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestHashSHA256(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string // Expected hex output
	}{
		{
			name:     "Empty string",
			data:     []byte(""),
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "Hello World",
			data:     []byte("Hello World"),
			expected: "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e",
		},
		{
			name:     "Binary data",
			data:     []byte{0x00, 0x01, 0xFF},
			expected: "6aa184c33af02a8c4754c1a49f0b5b0a85c8a5b6a0f5e5c2c7a7e7f7f7f7f7f7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := HashSHA256(tt.data)
			if len(hash) != 32 {
				t.Errorf("Hash length = %d, want 32", len(hash))
			}

			hashHex := HashSHA256Hex(tt.data)
			if len(hashHex) != 64 {
				t.Errorf("Hex hash length = %d, want 64", len(hashHex))
			}

			// Verify hex encoding is correct
			expectedBytes, _ := hex.DecodeString(hashHex)
			if !bytes.Equal(hash, expectedBytes) {
				t.Error("Hex encoding doesn't match binary hash")
			}
		})
	}
}

func TestHashSHA512(t *testing.T) {
	data := []byte("Test data")

	hash := HashSHA512(data)
	if len(hash) != 64 {
		t.Errorf("Hash length = %d, want 64", len(hash))
	}

	hashHex := HashSHA512Hex(data)
	if len(hashHex) != 128 {
		t.Errorf("Hex hash length = %d, want 128", len(hashHex))
	}

	expectedBytes, _ := hex.DecodeString(hashHex)
	if !bytes.Equal(hash, expectedBytes) {
		t.Error("Hex encoding doesn't match binary hash")
	}
}

func TestHashSHA384(t *testing.T) {
	data := []byte("Test data")

	hash := HashSHA384(data)
	if len(hash) != 48 {
		t.Errorf("Hash length = %d, want 48", len(hash))
	}

	hashHex := HashSHA384Hex(data)
	if len(hashHex) != 96 {
		t.Errorf("Hex hash length = %d, want 96", len(hashHex))
	}

	expectedBytes, _ := hex.DecodeString(hashHex)
	if !bytes.Equal(hash, expectedBytes) {
		t.Error("Hex encoding doesn't match binary hash")
	}
}

func TestHashDeterministic(t *testing.T) {
	data := []byte("Deterministic test")

	hash1 := HashSHA256(data)
	hash2 := HashSHA256(data)

	if !bytes.Equal(hash1, hash2) {
		t.Error("Hash function is not deterministic")
	}
}

func TestHMACSHA256(t *testing.T) {
	key := []byte("secret-key")
	data := []byte("message to authenticate")

	mac := HMACSHA256(key, data)
	if len(mac) != 32 {
		t.Errorf("HMAC length = %d, want 32", len(mac))
	}

	macHex := HMACSHA256Hex(key, data)
	if len(macHex) != 64 {
		t.Errorf("HMAC hex length = %d, want 64", len(macHex))
	}

	expectedBytes, _ := hex.DecodeString(macHex)
	if !bytes.Equal(mac, expectedBytes) {
		t.Error("Hex encoding doesn't match binary MAC")
	}
}

func TestHMACSHA512(t *testing.T) {
	key := []byte("secret-key")
	data := []byte("message to authenticate")

	mac := HMACSHA512(key, data)
	if len(mac) != 64 {
		t.Errorf("HMAC length = %d, want 64", len(mac))
	}

	macHex := HMACSHA512Hex(key, data)
	if len(macHex) != 128 {
		t.Errorf("HMAC hex length = %d, want 128", len(macHex))
	}

	expectedBytes, _ := hex.DecodeString(macHex)
	if !bytes.Equal(mac, expectedBytes) {
		t.Error("Hex encoding doesn't match binary MAC")
	}
}

func TestVerifyHMACSHA256(t *testing.T) {
	key := []byte("secret-key")
	data := []byte("message")

	mac := HMACSHA256(key, data)

	// Valid MAC should verify
	if !VerifyHMACSHA256(key, data, mac) {
		t.Error("Valid MAC failed verification")
	}

	// Tampered data should fail
	tamperedData := []byte("tampered message")
	if VerifyHMACSHA256(key, tamperedData, mac) {
		t.Error("Tampered data passed verification")
	}

	// Wrong key should fail
	wrongKey := []byte("wrong-key")
	if VerifyHMACSHA256(wrongKey, data, mac) {
		t.Error("Wrong key passed verification")
	}

	// Tampered MAC should fail
	tamperedMAC := make([]byte, len(mac))
	copy(tamperedMAC, mac)
	tamperedMAC[0] ^= 0xFF
	if VerifyHMACSHA256(key, data, tamperedMAC) {
		t.Error("Tampered MAC passed verification")
	}
}

func TestVerifyHMACSHA512(t *testing.T) {
	key := []byte("secret-key")
	data := []byte("message")

	mac := HMACSHA512(key, data)

	// Valid MAC should verify
	if !VerifyHMACSHA512(key, data, mac) {
		t.Error("Valid MAC failed verification")
	}

	// Tampered data should fail
	tamperedData := []byte("tampered")
	if VerifyHMACSHA512(key, tamperedData, mac) {
		t.Error("Tampered data passed verification")
	}

	// Wrong key should fail
	wrongKey := []byte("wrong")
	if VerifyHMACSHA512(wrongKey, data, mac) {
		t.Error("Wrong key passed verification")
	}
}

func TestHMACDeterministic(t *testing.T) {
	key := []byte("key")
	data := []byte("data")

	mac1 := HMACSHA256(key, data)
	mac2 := HMACSHA256(key, data)

	if !bytes.Equal(mac1, mac2) {
		t.Error("HMAC is not deterministic")
	}
}

func TestHMACWithDifferentKeys(t *testing.T) {
	key1 := []byte("key1")
	key2 := []byte("key2")
	data := []byte("data")

	mac1 := HMACSHA256(key1, data)
	mac2 := HMACSHA256(key2, data)

	if bytes.Equal(mac1, mac2) {
		t.Error("Different keys produced same MAC")
	}
}

func TestHMACWithEmptyData(t *testing.T) {
	key := []byte("key")
	data := []byte("")

	mac := HMACSHA256(key, data)
	if len(mac) != 32 {
		t.Errorf("HMAC length = %d, want 32", len(mac))
	}

	if !VerifyHMACSHA256(key, data, mac) {
		t.Error("Failed to verify HMAC of empty data")
	}
}

func TestHMACWithEmptyKey(t *testing.T) {
	key := []byte("")
	data := []byte("data")

	mac := HMACSHA256(key, data)
	if len(mac) != 32 {
		t.Errorf("HMAC length = %d, want 32", len(mac))
	}

	if !VerifyHMACSHA256(key, data, mac) {
		t.Error("Failed to verify HMAC with empty key")
	}
}
