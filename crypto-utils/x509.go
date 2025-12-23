package cryptoutils

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

// LicenseOID is the custom OID for license information extension
// 1.3.6.1.4.1.99999.1.1 - Private Enterprise Number placeholder
var LicenseOID = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 99999, 1, 1}

// LicenseInfo contains the license information stored in a certificate extension
type LicenseInfo struct {
	CustomerID   string            `json:"customerId"`
	MaxApps      int               `json:"maxApps"`
	Features     []string          `json:"features"`
	IssuedAt     time.Time         `json:"issuedAt"`
	CustomFields map[string]string `json:"customFields,omitempty"`
}

// CertificateConfig contains configuration for certificate generation
type CertificateConfig struct {
	Subject   pkix.Name
	NotBefore time.Time     // Required: Certificate validity start
	NotAfter  time.Time     // Required: Certificate validity end
	IsCA      bool          // Whether this is a CA certificate
	KeyUsage  x509.KeyUsage // Key usage flags
}

// Validate checks if the certificate config is valid
func (c *CertificateConfig) Validate() error {
	if c.NotBefore.IsZero() {
		return fmt.Errorf("%w: NotBefore is required", ErrInvalidValidityPeriod)
	}
	if c.NotAfter.IsZero() {
		return fmt.Errorf("%w: NotAfter is required", ErrInvalidValidityPeriod)
	}
	if !c.NotAfter.After(c.NotBefore) {
		return fmt.Errorf("%w: NotAfter must be after NotBefore", ErrInvalidValidityPeriod)
	}
	return nil
}

// GenerateCA generates a self-signed CA certificate with ECDSA P-256 key
func GenerateCA(config CertificateConfig) (*ecdsa.PrivateKey, *x509.Certificate, error) {
	if err := config.Validate(); err != nil {
		return nil, nil, err
	}

	// Generate ECDSA key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate ECDSA key: %w", err)
	}

	// Generate serial number
	serialNumber, err := generateSerialNumber()
	if err != nil {
		return nil, nil, err
	}

	// Prepare CA certificate template
	template := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               config.Subject,
		NotBefore:             config.NotBefore,
		NotAfter:              config.NotAfter,
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		MaxPathLen:            1,
	}

	if config.KeyUsage != 0 {
		template.KeyUsage = config.KeyUsage
	}

	// Create self-signed certificate
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create CA certificate: %w", err)
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	return privateKey, cert, nil
}

// GenerateCAWithRSA generates a self-signed CA certificate with RSA key
func GenerateCAWithRSA(bits int, config CertificateConfig) (*rsa.PrivateKey, *x509.Certificate, error) {
	if err := config.Validate(); err != nil {
		return nil, nil, err
	}

	if bits < 2048 {
		return nil, nil, fmt.Errorf("RSA key size must be at least 2048 bits")
	}

	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Generate serial number
	serialNumber, err := generateSerialNumber()
	if err != nil {
		return nil, nil, err
	}

	// Prepare CA certificate template
	template := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               config.Subject,
		NotBefore:             config.NotBefore,
		NotAfter:              config.NotAfter,
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		MaxPathLen:            1,
	}

	if config.KeyUsage != 0 {
		template.KeyUsage = config.KeyUsage
	}

	// Create self-signed certificate
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create CA certificate: %w", err)
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	return privateKey, cert, nil
}

// CreateLicenseCertificate creates a license certificate signed by the CA
func CreateLicenseCertificate(
	caKey crypto.Signer,
	caCert *x509.Certificate,
	license LicenseInfo,
	config CertificateConfig,
) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	if err := config.Validate(); err != nil {
		return nil, nil, err
	}

	if caCert == nil {
		return nil, nil, ErrInvalidCA
	}

	if !caCert.IsCA {
		return nil, nil, fmt.Errorf("%w: certificate is not a CA", ErrInvalidCA)
	}

	// Set IssuedAt if not set
	if license.IssuedAt.IsZero() {
		license.IssuedAt = time.Now()
	}

	// Validate license info
	if err := ValidateLicenseInfo(&license); err != nil {
		return nil, nil, err
	}

	// Encode license info to JSON
	licenseJSON, err := json.Marshal(license)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal license info: %w", err)
	}

	// ASN.1 encode the JSON bytes
	licenseASN1, err := asn1.Marshal(licenseJSON)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to ASN.1 encode license info: %w", err)
	}

	// Generate key pair for the license certificate
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate license key: %w", err)
	}

	// Generate serial number
	serialNumber, err := generateSerialNumber()
	if err != nil {
		return nil, nil, err
	}

	// Create license extension
	licenseExt := pkix.Extension{
		Id:       LicenseOID,
		Critical: false,
		Value:    licenseASN1,
	}

	// Prepare certificate template
	template := &x509.Certificate{
		SerialNumber:   serialNumber,
		Subject:        config.Subject,
		NotBefore:      config.NotBefore,
		NotAfter:       config.NotAfter,
		KeyUsage:       x509.KeyUsageDigitalSignature,
		ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		ExtraExtensions: []pkix.Extension{licenseExt},
	}

	if config.KeyUsage != 0 {
		template.KeyUsage = config.KeyUsage
	}

	// Create certificate signed by CA
	certDER, err := x509.CreateCertificate(rand.Reader, template, caCert, &privateKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create license certificate: %w", err)
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse license certificate: %w", err)
	}

	return cert, privateKey, nil
}

// VerifyLicenseCertificate verifies a license certificate against the CA certificate
func VerifyLicenseCertificate(cert *x509.Certificate, caCert *x509.Certificate) error {
	if cert == nil {
		return ErrInvalidCertificate
	}
	if caCert == nil {
		return ErrInvalidCA
	}

	// Check validity period
	now := time.Now()
	if now.Before(cert.NotBefore) {
		return ErrCertificateNotYetValid
	}
	if now.After(cert.NotAfter) {
		return ErrCertificateExpired
	}

	// Create a certificate pool with the CA certificate
	roots := x509.NewCertPool()
	roots.AddCert(caCert)

	// Verify the certificate chain
	opts := x509.VerifyOptions{
		Roots:       roots,
		CurrentTime: now,
		KeyUsages:   []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	_, err := cert.Verify(opts)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrVerificationFailed, err)
	}

	return nil
}

// ExtractLicenseInfo extracts the license information from a certificate
func ExtractLicenseInfo(cert *x509.Certificate) (*LicenseInfo, error) {
	if cert == nil {
		return nil, ErrInvalidCertificate
	}

	// Find the license extension
	for _, ext := range cert.Extensions {
		if ext.Id.Equal(LicenseOID) {
			// Decode ASN.1 wrapped JSON
			var jsonBytes []byte
			_, err := asn1.Unmarshal(ext.Value, &jsonBytes)
			if err != nil {
				return nil, fmt.Errorf("%w: failed to decode ASN.1: %v", ErrInvalidLicenseData, err)
			}

			// Parse JSON
			var license LicenseInfo
			if err := json.Unmarshal(jsonBytes, &license); err != nil {
				return nil, fmt.Errorf("%w: failed to parse JSON: %v", ErrInvalidLicenseData, err)
			}

			return &license, nil
		}
	}

	return nil, ErrLicenseExtNotFound
}

// EncodeCertificateToPEM encodes a certificate to PEM format
func EncodeCertificateToPEM(cert *x509.Certificate) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})
}

// DecodeCertificateFromPEM decodes a certificate from PEM format
func DecodeCertificateFromPEM(pemData []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("%w: failed to decode PEM block", ErrInvalidCertificate)
	}
	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("%w: unexpected PEM type: %s", ErrInvalidCertificate, block.Type)
	}
	return x509.ParseCertificate(block.Bytes)
}

// ValidateLicenseInfo validates the license information
func ValidateLicenseInfo(info *LicenseInfo) error {
	if info == nil {
		return fmt.Errorf("%w: license info is nil", ErrInvalidLicenseData)
	}
	if info.CustomerID == "" {
		return fmt.Errorf("%w: customerID is required", ErrInvalidLicenseData)
	}
	if info.MaxApps < 0 {
		return fmt.Errorf("%w: maxApps cannot be negative", ErrInvalidLicenseData)
	}
	return nil
}

// IsLicenseExpired checks if a certificate has expired
func IsLicenseExpired(cert *x509.Certificate) bool {
	if cert == nil {
		return true
	}
	return time.Now().After(cert.NotAfter)
}

// IsLicenseNotYetValid checks if a certificate is not yet valid
func IsLicenseNotYetValid(cert *x509.Certificate) bool {
	if cert == nil {
		return true
	}
	return time.Now().Before(cert.NotBefore)
}

// HasFeature checks if a license has a specific feature
func HasFeature(info *LicenseInfo, feature string) bool {
	if info == nil {
		return false
	}
	for _, f := range info.Features {
		if f == feature {
			return true
		}
	}
	return false
}

// generateSerialNumber generates a random serial number for certificates
func generateSerialNumber() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}
	return serialNumber, nil
}
