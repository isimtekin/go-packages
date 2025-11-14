# crypto-utils

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

Comprehensive, idiomatic Go cryptography utilities package providing AES encryption, RSA operations, ECDSA signing, ECDH key exchange, hashing, key derivation, secure random generation, and encoding helpers.

## Features

### 🔐 Symmetric Encryption
- **AES-GCM** - Authenticated encryption with automatic nonce handling
- **AES-CBC** - Block cipher mode with PKCS7 padding
- Support for AES-128, AES-192, and AES-256

### 🔑 Asymmetric Cryptography
- **RSA** - Key generation (2048/3072/4096 bit), OAEP encryption, PSS signing
- **ECDSA** - P-256/P-384/P-521 curves, signing/verification with SHA-256
- **ECDH** - X25519, P-256, P-384, P-521 key exchange
- PEM encoding/decoding for all key types

### 🔨 Hashing & MAC
- **SHA-256, SHA-384, SHA-512** - Cryptographic hash functions
- **HMAC-SHA256, HMAC-SHA512** - Message authentication codes
- Hex encoding helpers

### 🎲 Secure Random Generation
- Cryptographically secure random bytes
- Random strings and short IDs (URL-safe base64)
- Secure token generation (256-bit)
- Random integers and ranges
- Password generation (customizable character sets)
- PIN generation

### 🛡️ Key Derivation
- **PBKDF2** with SHA-256/SHA-512
- Configurable iterations (default: 210,000)
- Salt-based password-to-key derivation

### 📦 Encoding
- Base64 (standard and URL-safe)
- Raw base64 (without padding)
- Convenient encode/decode helpers

## Installation

```bash
go get github.com/isimtekin/go-packages/crypto-utils@v0.0.1
```

## Quick Start

### AES-GCM Encryption

```go
package main

import (
    "fmt"
    "log"

    cryptoutils "github.com/isimtekin/go-packages/crypto-utils"
)

func main() {
    // Generate a 256-bit key
    key, _ := cryptoutils.GenerateRandomBytes(32)
    plaintext := []byte("Secret message")

    // Encrypt
    ciphertext, err := cryptoutils.EncryptAESGCM(key, plaintext)
    if err != nil {
        log.Fatal(err)
    }

    // Decrypt
    decrypted, err := cryptoutils.DecryptAESGCM(key, ciphertext)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Decrypted: %s\n", decrypted)
}
```

### RSA Encryption (OAEP)

```go
// Generate RSA key pair
privateKey, err := cryptoutils.GenerateRSAKeyPair(2048)
if err != nil {
    log.Fatal(err)
}

plaintext := []byte("Confidential data")

// Encrypt with public key
ciphertext, err := cryptoutils.EncryptRSAOAEP(&privateKey.PublicKey, plaintext)
if err != nil {
    log.Fatal(err)
}

// Decrypt with private key
decrypted, err := cryptoutils.DecryptRSAOAEP(privateKey, ciphertext)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Decrypted: %s\n", decrypted)
```

### Digital Signatures (RSA-PSS)

```go
// Generate key pair
privateKey, _ := cryptoutils.GenerateRSAKeyPair(2048)
message := []byte("Document to sign")

// Sign
signature, err := cryptoutils.SignRSAPSS(privateKey, message)
if err != nil {
    log.Fatal(err)
}

// Verify
err = cryptoutils.VerifyRSAPSS(&privateKey.PublicKey, message, signature)
if err != nil {
    log.Fatal("Invalid signature:", err)
}

fmt.Println("Signature verified!")
```

### ECDSA Signing

```go
// Generate ECDSA key pair (P-256)
privateKey, _ := cryptoutils.GenerateECDSAKeyPair()
message := []byte("Message to sign")

// Sign and get bytes (R || S format)
signature, err := cryptoutils.SignECDSAToBytes(privateKey, message)
if err != nil {
    log.Fatal(err)
}

// Verify from bytes
valid := cryptoutils.VerifyECDSAFromBytes(&privateKey.PublicKey, message, signature)
if !valid {
    log.Fatal("Invalid signature")
}

fmt.Println("ECDSA signature verified!")
```

### ECDH Key Exchange

```go
// Alice generates key pair
alicePriv, _ := cryptoutils.GenerateECDHKeyPair() // X25519
alicePub := alicePriv.PublicKey()

// Bob generates key pair
bobPriv, _ := cryptoutils.GenerateECDHKeyPair()
bobPub := bobPriv.PublicKey()

// Alice derives shared secret
aliceShared, _ := cryptoutils.DeriveSharedSecret(alicePriv, bobPub)

// Bob derives shared secret
bobShared, _ := cryptoutils.DeriveSharedSecret(bobPriv, alicePub)

// aliceShared == bobShared ✓
fmt.Printf("Shared secret established: %x\n", aliceShared)
```

### Password-Based Key Derivation

```go
password := []byte("user-password")
salt, _ := cryptoutils.GenerateRandomBytes(16)

// Derive a 256-bit key using PBKDF2-SHA256
key := cryptoutils.DerivePBKDF2SHA256(password, salt, 210000, 32)

// Use the key for AES encryption
ciphertext, _ := cryptoutils.EncryptAESGCM(key, []byte("secret data"))
```

### Secure Password Generation

```go
// Generate a strong password (16 chars, all types)
password, _ := cryptoutils.GenerateStrongPassword()
fmt.Println("Password:", password)

// Custom password options
opts := cryptoutils.PasswordOptions{
    Length:         20,
    IncludeLower:   true,
    IncludeUpper:   true,
    IncludeDigits:  true,
    IncludeSpecial: false, // No special characters
}
simplePassword, _ := cryptoutils.GeneratePassword(opts)
fmt.Println("Simple password:", simplePassword)

// Generate a 6-digit PIN
pin, _ := cryptoutils.GeneratePIN(6)
fmt.Println("PIN:", pin)
```

### Secure Random IDs and Tokens

```go
// Generate short ID (22 chars, URL-safe)
id, _ := cryptoutils.GenerateShortID(22)
fmt.Println("Short ID:", id)

// Generate secure token (256-bit, base64 URL-safe)
token, _ := cryptoutils.GenerateSecureToken()
fmt.Println("Token:", token)

// Random integer in range [1, 100]
randomNum, _ := cryptoutils.GenerateRandomIntRange(1, 101)
fmt.Println("Random number:", randomNum)
```

### Hashing and HMAC

```go
data := []byte("data to hash")

// SHA-256
hash := cryptoutils.HashSHA256Hex(data)
fmt.Println("SHA-256:", hash)

// SHA-512
hash512 := cryptoutils.HashSHA512Hex(data)
fmt.Println("SHA-512:", hash512)

// HMAC-SHA256
key := []byte("secret-key")
mac := cryptoutils.HMACSHA256(key, data)

// Verify HMAC
valid := cryptoutils.VerifyHMACSHA256(key, data, mac)
fmt.Println("HMAC valid:", valid)
```

### Base64 Encoding

```go
data := []byte("Data to encode")

// Standard base64
encoded := cryptoutils.EncodeBase64(data)
decoded, _ := cryptoutils.DecodeBase64(encoded)

// URL-safe base64 (no padding)
urlEncoded := cryptoutils.EncodeBase64RawURL(data)
urlDecoded, _ := cryptoutils.DecodeBase64RawURL(urlEncoded)

fmt.Printf("Decoded: %s\n", decoded)
```

## API Reference

### AES Encryption

| Function | Description |
|----------|-------------|
| `EncryptAESGCM(key, plaintext []byte) ([]byte, error)` | Encrypt with AES-GCM (nonce prepended) |
| `DecryptAESGCM(key, ciphertext []byte) ([]byte, error)` | Decrypt AES-GCM ciphertext |
| `EncryptAESCBC(key, plaintext []byte) ([]byte, error)` | Encrypt with AES-CBC (IV prepended) |
| `DecryptAESCBC(key, ciphertext []byte) ([]byte, error)` | Decrypt AES-CBC ciphertext |

**Key sizes:** 16 bytes (AES-128), 24 bytes (AES-192), 32 bytes (AES-256)

### RSA Operations

| Function | Description |
|----------|-------------|
| `GenerateRSAKeyPair(bits int) (*rsa.PrivateKey, error)` | Generate RSA key pair (min 2048 bits) |
| `EncryptRSAOAEP(publicKey *rsa.PublicKey, plaintext []byte) ([]byte, error)` | Encrypt with RSA-OAEP SHA-256 |
| `DecryptRSAOAEP(privateKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error)` | Decrypt RSA-OAEP ciphertext |
| `SignRSAPSS(privateKey *rsa.PrivateKey, message []byte) ([]byte, error)` | Sign with RSA-PSS SHA-256 |
| `VerifyRSAPSS(publicKey *rsa.PublicKey, message, signature []byte) error` | Verify RSA-PSS signature |
| `EncodeRSAPrivateKeyToPEM(key *rsa.PrivateKey) []byte` | Encode private key to PEM |
| `EncodeRSAPublicKeyToPEM(key *rsa.PublicKey) ([]byte, error)` | Encode public key to PEM |
| `DecodeRSAPrivateKeyFromPEM(pemData []byte) (*rsa.PrivateKey, error)` | Decode private key from PEM |
| `DecodeRSAPublicKeyFromPEM(pemData []byte) (*rsa.PublicKey, error)` | Decode public key from PEM |

### ECDSA Operations

| Function | Description |
|----------|-------------|
| `GenerateECDSAKeyPair() (*ecdsa.PrivateKey, error)` | Generate ECDSA key pair (P-256) |
| `GenerateECDSAKeyPairWithCurve(curve elliptic.Curve) (*ecdsa.PrivateKey, error)` | Generate with specific curve |
| `SignECDSA(privateKey *ecdsa.PrivateKey, message []byte) (r, s *big.Int, err error)` | Sign message |
| `VerifyECDSA(publicKey *ecdsa.PublicKey, message []byte, r, s *big.Int) bool` | Verify signature |
| `SignECDSAToBytes(privateKey *ecdsa.PrivateKey, message []byte) ([]byte, error)` | Sign to byte array |
| `VerifyECDSAFromBytes(publicKey *ecdsa.PublicKey, message, signature []byte) bool` | Verify from bytes |
| `EncodeECDSAPrivateKeyToPEM(key *ecdsa.PrivateKey) ([]byte, error)` | Encode to PEM |
| `EncodeECDSAPublicKeyToPEM(key *ecdsa.PublicKey) ([]byte, error)` | Encode to PEM |
| `DecodeECDSAPrivateKeyFromPEM(pemData []byte) (*ecdsa.PrivateKey, error)` | Decode from PEM |
| `DecodeECDSAPublicKeyFromPEM(pemData []byte) (*ecdsa.PublicKey, error)` | Decode from PEM |

### ECDH Key Exchange

| Function | Description |
|----------|-------------|
| `GenerateECDHKeyPair() (*ecdh.PrivateKey, error)` | Generate X25519 key pair |
| `GenerateECDHKeyPairP256() (*ecdh.PrivateKey, error)` | Generate P-256 key pair |
| `GenerateECDHKeyPairP384() (*ecdh.PrivateKey, error)` | Generate P-384 key pair |
| `GenerateECDHKeyPairP521() (*ecdh.PrivateKey, error)` | Generate P-521 key pair |
| `DeriveSharedSecret(privateKey *ecdh.PrivateKey, publicKey *ecdh.PublicKey) ([]byte, error)` | Derive shared secret |
| `EncodeECDHPrivateKey(key *ecdh.PrivateKey) []byte` | Encode private key |
| `EncodeECDHPublicKey(key *ecdh.PublicKey) []byte` | Encode public key |
| `DecodeECDHPrivateKeyX25519(data []byte) (*ecdh.PrivateKey, error)` | Decode X25519 private key |
| `DecodeECDHPublicKeyX25519(data []byte) (*ecdh.PublicKey, error)` | Decode X25519 public key |

### Hashing

| Function | Description |
|----------|-------------|
| `HashSHA256(data []byte) []byte` | SHA-256 hash |
| `HashSHA256Hex(data []byte) string` | SHA-256 hash as hex string |
| `HashSHA512(data []byte) []byte` | SHA-512 hash |
| `HashSHA512Hex(data []byte) string` | SHA-512 hash as hex string |
| `HashSHA384(data []byte) []byte` | SHA-384 hash |
| `HashSHA384Hex(data []byte) string` | SHA-384 hash as hex string |

### HMAC

| Function | Description |
|----------|-------------|
| `HMACSHA256(key, data []byte) []byte` | HMAC-SHA256 |
| `HMACSHA256Hex(key, data []byte) string` | HMAC-SHA256 as hex |
| `VerifyHMACSHA256(key, data, expectedMAC []byte) bool` | Verify HMAC-SHA256 (constant time) |
| `HMACSHA512(key, data []byte) []byte` | HMAC-SHA512 |
| `HMACSHA512Hex(key, data []byte) string` | HMAC-SHA512 as hex |
| `VerifyHMACSHA512(key, data, expectedMAC []byte) bool` | Verify HMAC-SHA512 (constant time) |

### Key Derivation

| Function | Description |
|----------|-------------|
| `DerivePBKDF2SHA256(password, salt []byte, iterations, keyLen int) []byte` | PBKDF2 with SHA-256 |
| `DerivePBKDF2SHA512(password, salt []byte, iterations, keyLen int) []byte` | PBKDF2 with SHA-512 |
| `DeriveKeyFromPassword(password, salt []byte) []byte` | Convenience function (210k iterations, 32-byte key) |

**Recommended iterations:** 210,000+ (adjustable based on security requirements)

### Random Generation

| Function | Description |
|----------|-------------|
| `GenerateRandomBytes(n int) ([]byte, error)` | Generate n random bytes |
| `GenerateRandomString(n int) (string, error)` | Generate random string (base64 URL-safe) |
| `GenerateShortID(length int) (string, error)` | Generate short ID (default 22 chars) |
| `GenerateSecureToken() (string, error)` | Generate 256-bit token (43 chars) |
| `GenerateRandomInt(max int64) (int64, error)` | Random int in [0, max) |
| `GenerateRandomIntRange(min, max int64) (int64, error)` | Random int in [min, max) |

### Password Generation

| Function | Description |
|----------|-------------|
| `GeneratePassword(opts PasswordOptions) (string, error)` | Generate password with options |
| `GenerateStrongPassword() (string, error)` | Generate strong 16-char password |
| `GenerateSimplePassword(length int) (string, error)` | Generate alphanumeric password |
| `GeneratePIN(length int) (string, error)` | Generate numeric PIN |
| `DefaultPasswordOptions() PasswordOptions` | Get default password options |

**PasswordOptions:**
```go
type PasswordOptions struct {
    Length         int  // Password length
    IncludeLower   bool // Include lowercase letters
    IncludeUpper   bool // Include uppercase letters
    IncludeDigits  bool // Include digits
    IncludeSpecial bool // Include special characters
}
```

### Base64 Encoding

| Function | Description |
|----------|-------------|
| `EncodeBase64(data []byte) string` | Standard base64 encoding |
| `DecodeBase64(encoded string) ([]byte, error)` | Standard base64 decoding |
| `EncodeBase64URL(data []byte) string` | URL-safe base64 with padding |
| `DecodeBase64URL(encoded string) ([]byte, error)` | URL-safe base64 decoding |
| `EncodeBase64RawURL(data []byte) string` | URL-safe base64 without padding |
| `DecodeBase64RawURL(encoded string) ([]byte, error)` | URL-safe base64 decoding (no padding) |
| `EncodeBase64Raw(data []byte) string` | Standard base64 without padding |
| `DecodeBase64Raw(encoded string) ([]byte, error)` | Standard base64 decoding (no padding) |

## Security Considerations

### Key Sizes
- **AES:** Use 256-bit keys (32 bytes) for maximum security
- **RSA:** Minimum 2048 bits, 3072+ recommended for long-term security
- **ECDSA/ECDH:** P-256 provides ~128-bit security, P-384 provides ~192-bit

### Random Generation
- All random functions use `crypto/rand` for cryptographically secure randomness
- Never use `math/rand` for security-sensitive operations

### Key Derivation
- Use at least 210,000 PBKDF2 iterations (OWASP recommendation as of 2023)
- Always use a unique random salt for each password
- Salt should be at least 16 bytes

### Password Storage
- Never store passwords in plaintext
- Use PBKDF2, bcrypt, scrypt, or Argon2 for password hashing
- Store the salt alongside the hash

### Authentication
- HMAC verification uses constant-time comparison to prevent timing attacks
- Always verify signatures before trusting signed data

## Error Handling

```go
import cryptoutils "github.com/isimtekin/go-packages/crypto-utils"

// Check for specific errors
key := []byte("invalid-key")
_, err := cryptoutils.EncryptAESGCM(key, plaintext)
if err == cryptoutils.ErrInvalidKeySize {
    // Handle invalid key size
}
```

**Common Errors:**
- `ErrInvalidKeySize` - Key size doesn't match AES requirements
- `ErrInvalidPadding` - Invalid PKCS7 padding
- `ErrInvalidRange` - Invalid range for random int generation
- `ErrKeyGenerationFailed` - Failed to generate cryptographic key
- `ErrSignatureVerification` - Signature verification failed

## Testing

```bash
# Run all tests
go test -v ./...

# Run with race detection
go test -v -race ./...

# Run with coverage
go test -v -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Examples

See the examples in the test files:
- `aes_test.go` - AES encryption examples
- `crypto_test.go` - Comprehensive crypto operation examples

## Dependencies

- `golang.org/x/crypto/pbkdf2` - PBKDF2 key derivation
- Go standard library (`crypto/*`, `encoding/*`)

## Performance Notes

- AES-GCM is faster than AES-CBC and provides authentication
- ECDSA operations are faster than RSA for signing
- X25519 (ECDH) is faster and safer than RSA for key exchange
- SHA-256 is faster than SHA-512 on 32-bit architectures

## License

MIT License - see [LICENSE](../LICENSE) file for details.

## Related Packages

- [env-util](../env-util) - Environment variable utilities
- [mongo-client](../mongo-client) - MongoDB client wrapper
- [redis-client](../redis-client) - Redis client wrapper
- [nats-client](../nats-client) - NATS messaging client

## Resources

- [Go Cryptography Documentation](https://pkg.go.dev/crypto)
- [OWASP Cryptographic Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Cryptographic_Storage_Cheat_Sheet.html)
- [NIST Recommendations](https://www.keylength.com/)
