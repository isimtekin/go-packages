package cryptoutils

import (
	"bytes"
	"testing"
)

func TestBase64Encoding(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"Empty", []byte("")},
		{"Short", []byte("Hi")},
		{"ASCII", []byte("Hello, World!")},
		{"Binary", []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}},
		{"Long", bytes.Repeat([]byte("x"), 1000)},
		{"Unicode", []byte("Hello 世界 🌍")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := EncodeBase64(tt.data)
			decoded, err := DecodeBase64(encoded)
			if err != nil {
				t.Fatalf("DecodeBase64 failed: %v", err)
			}

			if !bytes.Equal(decoded, tt.data) {
				t.Errorf("Decoded data doesn't match original")
			}
		})
	}
}

func TestBase64URLEncoding(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"Empty", []byte("")},
		{"Short", []byte("Hi")},
		{"Medium", []byte("Test data for URL encoding")},
		{"Binary", []byte{0x00, 0xFF, 0x7F, 0x80}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := EncodeBase64URL(tt.data)
			decoded, err := DecodeBase64URL(encoded)
			if err != nil {
				t.Fatalf("DecodeBase64URL failed: %v", err)
			}

			if !bytes.Equal(decoded, tt.data) {
				t.Errorf("Decoded data doesn't match original")
			}

			// URL-safe encoding should not contain + or /
			// but may contain = for padding
			for _, c := range encoded {
				if c == '+' || c == '/' {
					t.Errorf("URL encoding contains non-URL-safe character: %c", c)
				}
			}
		})
	}
}

func TestBase64RawURLEncoding(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"Empty", []byte("")},
		{"Short", []byte("Hi")},
		{"Medium", []byte("Test data")},
		{"Binary", []byte{0xFF, 0xFE, 0xFD}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := EncodeBase64RawURL(tt.data)
			decoded, err := DecodeBase64RawURL(encoded)
			if err != nil {
				t.Fatalf("DecodeBase64RawURL failed: %v", err)
			}

			if !bytes.Equal(decoded, tt.data) {
				t.Errorf("Decoded data doesn't match original")
			}

			// Raw encoding should not contain padding (=)
			for _, c := range encoded {
				if c == '=' {
					t.Error("Raw encoding should not contain padding")
				}
				if c == '+' || c == '/' {
					t.Errorf("URL encoding contains non-URL-safe character: %c", c)
				}
			}
		})
	}
}

func TestBase64RawEncoding(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"Empty", []byte("")},
		{"Short", []byte("A")},
		{"Medium", []byte("Test")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := EncodeBase64Raw(tt.data)
			decoded, err := DecodeBase64Raw(encoded)
			if err != nil {
				t.Fatalf("DecodeBase64Raw failed: %v", err)
			}

			if !bytes.Equal(decoded, tt.data) {
				t.Errorf("Decoded data doesn't match original")
			}

			// Raw encoding should not contain padding
			for _, c := range encoded {
				if c == '=' {
					t.Error("Raw encoding should not contain padding")
				}
			}
		})
	}
}

func TestBase64DecodeErrors(t *testing.T) {
	tests := []struct {
		name    string
		encoded string
	}{
		{"Invalid characters", "!!!invalid!!!"},
		{"Partial encoding", "ABC"},
		{"Corrupted", "YWJj===="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeBase64(tt.encoded)
			if err == nil {
				t.Error("Expected error for invalid base64")
			}
		})
	}
}

func TestBase64URLDecodeErrors(t *testing.T) {
	tests := []struct {
		name    string
		encoded string
	}{
		{"Invalid characters", "!!!invalid!!!"},
		{"Wrong encoding", "YWJj+/=="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeBase64URL(tt.encoded)
			if err == nil {
				t.Error("Expected error for invalid base64 URL")
			}
		})
	}
}

func TestBase64EncodingConsistency(t *testing.T) {
	data := []byte("Consistency test")

	// Multiple encodings of same data should be identical
	encoded1 := EncodeBase64(data)
	encoded2 := EncodeBase64(data)

	if encoded1 != encoded2 {
		t.Error("Base64 encoding is not consistent")
	}
}

func TestBase64URLvsStandard(t *testing.T) {
	// Data that will produce + and / in standard base64
	data := []byte{0xFB, 0xFF, 0xBF}

	standard := EncodeBase64(data)
	urlSafe := EncodeBase64URL(data)

	// Standard base64 might contain + or /
	// URL-safe should not
	hasNonURLSafe := false
	for _, c := range standard {
		if c == '+' || c == '/' {
			hasNonURLSafe = true
			break
		}
	}

	for _, c := range urlSafe {
		if c == '+' || c == '/' {
			t.Error("URL-safe encoding contains + or /")
		}
	}

	// Both should decode to same data
	decoded1, _ := DecodeBase64(standard)
	decoded2, _ := DecodeBase64URL(urlSafe)

	if !bytes.Equal(decoded1, decoded2) {
		t.Error("Standard and URL-safe decodings don't match")
	}

	if !bytes.Equal(decoded1, data) {
		t.Error("Decoded data doesn't match original")
	}

	// Log for informational purposes
	if hasNonURLSafe {
		t.Logf("Standard: %s (contains + or /)", standard)
		t.Logf("URL-safe: %s (no + or /)", urlSafe)
	}
}

func TestBase64LargeData(t *testing.T) {
	// Test with larger data
	data := bytes.Repeat([]byte("Large data test "), 1000)

	encoded := EncodeBase64(data)
	decoded, err := DecodeBase64(encoded)
	if err != nil {
		t.Fatalf("Failed to decode large data: %v", err)
	}

	if !bytes.Equal(decoded, data) {
		t.Error("Large data encoding/decoding failed")
	}
}

func TestBase64EmptyString(t *testing.T) {
	data := []byte("")

	encoded := EncodeBase64(data)
	if encoded != "" {
		t.Errorf("Empty data should encode to empty string, got: %s", encoded)
	}

	decoded, err := DecodeBase64(encoded)
	if err != nil {
		t.Fatalf("Failed to decode empty string: %v", err)
	}

	if len(decoded) != 0 {
		t.Errorf("Decoded empty string has length %d, want 0", len(decoded))
	}
}

func TestBase64AllBytes(t *testing.T) {
	// Test with all possible byte values
	data := make([]byte, 256)
	for i := 0; i < 256; i++ {
		data[i] = byte(i)
	}

	encoded := EncodeBase64(data)
	decoded, err := DecodeBase64(encoded)
	if err != nil {
		t.Fatalf("Failed to decode all bytes: %v", err)
	}

	if !bytes.Equal(decoded, data) {
		t.Error("All bytes encoding/decoding failed")
	}
}
