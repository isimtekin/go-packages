package cryptoutils

import (
	"crypto/rand"
	"errors"
	"math/big"
)

// Password generation

const (
	lowercaseLetters = "abcdefghijklmnopqrstuvwxyz"
	uppercaseLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits           = "0123456789"
	specialChars     = "!@#$%^&*()_+-=[]{}|;:,.<>?"
)

// PasswordOptions configures password generation
type PasswordOptions struct {
	Length         int
	IncludeLower   bool
	IncludeUpper   bool
	IncludeDigits  bool
	IncludeSpecial bool
}

// DefaultPasswordOptions returns recommended password generation options
func DefaultPasswordOptions() PasswordOptions {
	return PasswordOptions{
		Length:         16,
		IncludeLower:   true,
		IncludeUpper:   true,
		IncludeDigits:  true,
		IncludeSpecial: true,
	}
}

// GeneratePassword generates a cryptographically secure random password
func GeneratePassword(opts PasswordOptions) (string, error) {
	if opts.Length <= 0 {
		return "", errors.New("password length must be positive")
	}

	// Build character set
	charset := ""
	if opts.IncludeLower {
		charset += lowercaseLetters
	}
	if opts.IncludeUpper {
		charset += uppercaseLetters
	}
	if opts.IncludeDigits {
		charset += digits
	}
	if opts.IncludeSpecial {
		charset += specialChars
	}

	if len(charset) == 0 {
		return "", errors.New("at least one character type must be included")
	}

	password := make([]byte, opts.Length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < opts.Length; i++ {
		num, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		password[i] = charset[num.Int64()]
	}

	return string(password), nil
}

// GenerateStrongPassword generates a strong password with default options (16 chars, all types)
func GenerateStrongPassword() (string, error) {
	return GeneratePassword(DefaultPasswordOptions())
}

// GenerateSimplePassword generates a simple alphanumeric password (no special chars)
func GenerateSimplePassword(length int) (string, error) {
	opts := PasswordOptions{
		Length:         length,
		IncludeLower:   true,
		IncludeUpper:   true,
		IncludeDigits:  true,
		IncludeSpecial: false,
	}
	return GeneratePassword(opts)
}

// GeneratePIN generates a numeric PIN of specified length
func GeneratePIN(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("PIN length must be positive")
	}

	pin := make([]byte, length)
	digitsLen := big.NewInt(int64(len(digits)))

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, digitsLen)
		if err != nil {
			return "", err
		}
		pin[i] = digits[num.Int64()]
	}

	return string(pin), nil
}
