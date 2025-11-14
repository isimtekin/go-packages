package cryptoutils

import (
	"bytes"
	"testing"
)

func TestGenerateRandomBytes(t *testing.T) {
	sizes := []int{8, 16, 32, 64, 128, 256}

	for _, size := range sizes {
		t.Run(string(rune(size)), func(t *testing.T) {
			b, err := GenerateRandomBytes(size)
			if err != nil {
				t.Fatalf("GenerateRandomBytes failed: %v", err)
			}

			if len(b) != size {
				t.Errorf("Byte length = %d, want %d", len(b), size)
			}

			// Generate another to ensure they're different
			b2, _ := GenerateRandomBytes(size)
			if bytes.Equal(b, b2) {
				t.Error("Generated identical random bytes")
			}
		})
	}
}

func TestGenerateRandomBytesZeroSize(t *testing.T) {
	b, err := GenerateRandomBytes(0)
	if err != nil {
		t.Fatalf("GenerateRandomBytes(0) failed: %v", err)
	}
	if len(b) != 0 {
		t.Errorf("Byte length = %d, want 0", len(b))
	}
}

func TestGenerateRandomString(t *testing.T) {
	lengths := []int{8, 16, 22, 32, 64}

	for _, length := range lengths {
		t.Run(string(rune(length)), func(t *testing.T) {
			str, err := GenerateRandomString(length)
			if err != nil {
				t.Fatalf("GenerateRandomString failed: %v", err)
			}

			if len(str) < length-1 || len(str) > length {
				t.Errorf("String length = %d, want ~%d", len(str), length)
			}

			// Should be URL-safe (no +, /, or = characters)
			for _, c := range str {
				if c == '+' || c == '/' || c == '=' {
					t.Errorf("String contains non-URL-safe character: %c", c)
				}
			}

			// Generate another to ensure they're different
			str2, _ := GenerateRandomString(length)
			if str == str2 {
				t.Error("Generated identical random strings")
			}
		})
	}
}

func TestGenerateShortID(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"Default", 22},
		{"Short", 8},
		{"Medium", 16},
		{"Long", 32},
		{"Zero uses default", 0},
		{"Negative uses default", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := GenerateShortID(tt.length)
			if err != nil {
				t.Fatalf("GenerateShortID failed: %v", err)
			}

			expectedLen := tt.length
			if tt.length <= 0 {
				expectedLen = 22 // Default
			}

			if len(id) < expectedLen-1 || len(id) > expectedLen {
				t.Errorf("ID length = %d, want ~%d", len(id), expectedLen)
			}

			// Should be URL-safe
			for _, c := range id {
				if c == '+' || c == '/' || c == '=' {
					t.Errorf("ID contains non-URL-safe character: %c", c)
				}
			}
		})
	}
}

func TestGenerateShortIDUniqueness(t *testing.T) {
	ids := make(map[string]bool)
	count := 1000

	for i := 0; i < count; i++ {
		id, err := GenerateShortID(22)
		if err != nil {
			t.Fatalf("Failed to generate ID: %v", err)
		}

		if ids[id] {
			t.Error("Generated duplicate ID")
		}
		ids[id] = true
	}

	if len(ids) != count {
		t.Errorf("Generated %d unique IDs, want %d", len(ids), count)
	}
}

func TestGenerateSecureToken(t *testing.T) {
	token, err := GenerateSecureToken()
	if err != nil {
		t.Fatalf("GenerateSecureToken failed: %v", err)
	}

	// Should be 43 characters (32 bytes base64 URL-safe without padding)
	if len(token) != 43 {
		t.Errorf("Token length = %d, want 43", len(token))
	}

	// Should be URL-safe
	for _, c := range token {
		if c == '+' || c == '/' || c == '=' {
			t.Errorf("Token contains non-URL-safe character: %c", c)
		}
	}

	// Generate another to ensure uniqueness
	token2, _ := GenerateSecureToken()
	if token == token2 {
		t.Error("Generated identical tokens")
	}
}

func TestGenerateSecureTokenUniqueness(t *testing.T) {
	tokens := make(map[string]bool)
	count := 100

	for i := 0; i < count; i++ {
		token, err := GenerateSecureToken()
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		if tokens[token] {
			t.Error("Generated duplicate token")
		}
		tokens[token] = true
	}
}

func TestGenerateRandomInt(t *testing.T) {
	tests := []struct {
		name string
		max  int64
	}{
		{"Small range", 10},
		{"Medium range", 100},
		{"Large range", 1000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 100; i++ {
				n, err := GenerateRandomInt(tt.max)
				if err != nil {
					t.Fatalf("GenerateRandomInt failed: %v", err)
				}

				if n < 0 || n >= tt.max {
					t.Errorf("Random int %d out of range [0, %d)", n, tt.max)
				}
			}
		})
	}
}

func TestGenerateRandomIntInvalidRange(t *testing.T) {
	tests := []int64{0, -1, -100}

	for _, max := range tests {
		t.Run(string(rune(max)), func(t *testing.T) {
			_, err := GenerateRandomInt(max)
			if err != ErrInvalidRange {
				t.Errorf("Expected ErrInvalidRange, got %v", err)
			}
		})
	}
}

func TestGenerateRandomIntRange(t *testing.T) {
	tests := []struct {
		name string
		min  int64
		max  int64
	}{
		{"Positive range", 10, 20},
		{"Large range", 100, 1000},
		{"Negative to positive", -50, 50},
		{"Single value", 5, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 100; i++ {
				n, err := GenerateRandomIntRange(tt.min, tt.max)
				if err != nil {
					t.Fatalf("GenerateRandomIntRange failed: %v", err)
				}

				if n < tt.min || n >= tt.max {
					t.Errorf("Random int %d out of range [%d, %d)", n, tt.min, tt.max)
				}
			}
		})
	}
}

func TestGenerateRandomIntRangeInvalid(t *testing.T) {
	tests := []struct {
		name string
		min  int64
		max  int64
	}{
		{"Min >= Max", 10, 10},
		{"Min > Max", 20, 10},
		{"Negative range", -10, -20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GenerateRandomIntRange(tt.min, tt.max)
			if err != ErrInvalidRange {
				t.Errorf("Expected ErrInvalidRange, got %v", err)
			}
		})
	}
}

func TestGenerateRandomIntDistribution(t *testing.T) {
	// Test that distribution is reasonably uniform
	max := int64(10)
	counts := make(map[int64]int)
	iterations := 10000

	for i := 0; i < iterations; i++ {
		n, err := GenerateRandomInt(max)
		if err != nil {
			t.Fatalf("Failed to generate random int: %v", err)
		}
		counts[n]++
	}

	// Each number should appear roughly iterations/max times
	expected := iterations / int(max)
	tolerance := expected / 2 // Allow 50% deviation

	for i := int64(0); i < max; i++ {
		count := counts[i]
		if count < expected-tolerance || count > expected+tolerance {
			t.Logf("Warning: Number %d appeared %d times, expected ~%d", i, count, expected)
		}
	}

	// At least verify all numbers appeared
	if len(counts) != int(max) {
		t.Errorf("Only %d different numbers generated, want %d", len(counts), max)
	}
}
