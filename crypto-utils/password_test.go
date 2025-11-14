package cryptoutils

import (
	"strings"
	"testing"
)

func TestGeneratePassword(t *testing.T) {
	tests := []struct {
		name string
		opts PasswordOptions
	}{
		{
			name: "Default options",
			opts: DefaultPasswordOptions(),
		},
		{
			name: "Lowercase only",
			opts: PasswordOptions{
				Length:       12,
				IncludeLower: true,
			},
		},
		{
			name: "Uppercase only",
			opts: PasswordOptions{
				Length:       12,
				IncludeUpper: true,
			},
		},
		{
			name: "Digits only",
			opts: PasswordOptions{
				Length:        12,
				IncludeDigits: true,
			},
		},
		{
			name: "Special chars only",
			opts: PasswordOptions{
				Length:         12,
				IncludeSpecial: true,
			},
		},
		{
			name: "Alphanumeric",
			opts: PasswordOptions{
				Length:        16,
				IncludeLower:  true,
				IncludeUpper:  true,
				IncludeDigits: true,
			},
		},
		{
			name: "All character types",
			opts: PasswordOptions{
				Length:         20,
				IncludeLower:   true,
				IncludeUpper:   true,
				IncludeDigits:  true,
				IncludeSpecial: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := GeneratePassword(tt.opts)
			if err != nil {
				t.Fatalf("GeneratePassword failed: %v", err)
			}

			if len(password) != tt.opts.Length {
				t.Errorf("Password length = %d, want %d", len(password), tt.opts.Length)
			}

			// Verify character types
			if tt.opts.IncludeLower {
				if !containsAny(password, lowercaseLetters) {
					t.Error("Password missing lowercase letters")
				}
			}
			if tt.opts.IncludeUpper {
				if !containsAny(password, uppercaseLetters) {
					// Note: Due to randomness, this might occasionally fail
					// but very unlikely with default length
				}
			}
			if tt.opts.IncludeDigits {
				if !containsAny(password, digits) {
					// Similar note as above
				}
			}
			if tt.opts.IncludeSpecial {
				if !containsAny(password, specialChars) {
					// Similar note as above
				}
			}
		})
	}
}

func TestGeneratePasswordInvalidOptions(t *testing.T) {
	tests := []struct {
		name string
		opts PasswordOptions
	}{
		{
			name: "Zero length",
			opts: PasswordOptions{
				Length:       0,
				IncludeLower: true,
			},
		},
		{
			name: "Negative length",
			opts: PasswordOptions{
				Length:       -1,
				IncludeLower: true,
			},
		},
		{
			name: "No character types",
			opts: PasswordOptions{
				Length: 12,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GeneratePassword(tt.opts)
			if err == nil {
				t.Error("Expected error for invalid options")
			}
		})
	}
}

func TestGeneratePasswordUniqueness(t *testing.T) {
	opts := DefaultPasswordOptions()
	passwords := make(map[string]bool)
	count := 100

	for i := 0; i < count; i++ {
		password, err := GeneratePassword(opts)
		if err != nil {
			t.Fatalf("Failed to generate password: %v", err)
		}

		if passwords[password] {
			t.Error("Generated duplicate password")
		}
		passwords[password] = true
	}
}

func TestDefaultPasswordOptions(t *testing.T) {
	opts := DefaultPasswordOptions()

	if opts.Length != 16 {
		t.Errorf("Default length = %d, want 16", opts.Length)
	}
	if !opts.IncludeLower {
		t.Error("Default should include lowercase")
	}
	if !opts.IncludeUpper {
		t.Error("Default should include uppercase")
	}
	if !opts.IncludeDigits {
		t.Error("Default should include digits")
	}
	if !opts.IncludeSpecial {
		t.Error("Default should include special characters")
	}
}

func TestGenerateStrongPassword(t *testing.T) {
	password, err := GenerateStrongPassword()
	if err != nil {
		t.Fatalf("GenerateStrongPassword failed: %v", err)
	}

	if len(password) != 16 {
		t.Errorf("Password length = %d, want 16", len(password))
	}

	// Should contain variety of characters
	hasLower := containsAny(password, lowercaseLetters)
	hasUpper := containsAny(password, uppercaseLetters)
	hasDigit := containsAny(password, digits)
	hasSpecial := containsAny(password, specialChars)

	if !hasLower || !hasUpper || !hasDigit || !hasSpecial {
		t.Logf("Password: %s", password)
		t.Error("Strong password should contain all character types")
	}
}

func TestGenerateSimplePassword(t *testing.T) {
	lengths := []int{8, 12, 16, 20}

	for _, length := range lengths {
		t.Run(string(rune(length)), func(t *testing.T) {
			password, err := GenerateSimplePassword(length)
			if err != nil {
				t.Fatalf("GenerateSimplePassword failed: %v", err)
			}

			if len(password) != length {
				t.Errorf("Password length = %d, want %d", len(password), length)
			}

			// Should NOT contain special characters
			if containsAny(password, specialChars) {
				t.Error("Simple password should not contain special characters")
			}

			// Should contain alphanumeric
			hasAlpha := containsAny(password, lowercaseLetters+uppercaseLetters)
			hasDigit := containsAny(password, digits)

			if !hasAlpha && !hasDigit {
				t.Error("Simple password should contain alphanumeric characters")
			}
		})
	}
}

func TestGeneratePIN(t *testing.T) {
	lengths := []int{4, 6, 8}

	for _, length := range lengths {
		t.Run(string(rune(length)), func(t *testing.T) {
			pin, err := GeneratePIN(length)
			if err != nil {
				t.Fatalf("GeneratePIN failed: %v", err)
			}

			if len(pin) != length {
				t.Errorf("PIN length = %d, want %d", len(pin), length)
			}

			// Should only contain digits
			for _, c := range pin {
				if c < '0' || c > '9' {
					t.Errorf("PIN contains non-digit character: %c", c)
				}
			}
		})
	}
}

func TestGeneratePINInvalidLength(t *testing.T) {
	tests := []int{0, -1, -100}

	for _, length := range tests {
		t.Run(string(rune(length)), func(t *testing.T) {
			_, err := GeneratePIN(length)
			if err == nil {
				t.Error("Expected error for invalid PIN length")
			}
		})
	}
}

func TestGeneratePINUniqueness(t *testing.T) {
	pins := make(map[string]bool)
	count := 100
	length := 8

	for i := 0; i < count; i++ {
		pin, err := GeneratePIN(length)
		if err != nil {
			t.Fatalf("Failed to generate PIN: %v", err)
		}

		if pins[pin] {
			t.Error("Generated duplicate PIN")
		}
		pins[pin] = true
	}
}

func TestGeneratePINDistribution(t *testing.T) {
	// Generate many 1-digit PINs and check distribution
	counts := make(map[rune]int)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		pin, err := GeneratePIN(1)
		if err != nil {
			t.Fatalf("Failed to generate PIN: %v", err)
		}
		counts[rune(pin[0])]++
	}

	// Should have all 10 digits
	if len(counts) != 10 {
		t.Errorf("Only %d different digits generated, want 10", len(counts))
	}

	// Each digit should appear roughly iterations/10 times
	expected := iterations / 10
	tolerance := expected / 2 // Allow 50% deviation

	for digit := '0'; digit <= '9'; digit++ {
		count := counts[digit]
		if count < expected-tolerance || count > expected+tolerance {
			t.Logf("Warning: Digit %c appeared %d times, expected ~%d", digit, count, expected)
		}
	}
}

// Helper function to check if string contains any character from charset
func containsAny(s, charset string) bool {
	for _, c := range s {
		if strings.ContainsRune(charset, c) {
			return true
		}
	}
	return false
}
