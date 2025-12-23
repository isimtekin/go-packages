package cryptoutils

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"testing"
	"time"
)

func TestGenerateCA(t *testing.T) {
	tests := []struct {
		name    string
		config  CertificateConfig
		wantErr bool
	}{
		{
			name: "valid CA",
			config: CertificateConfig{
				Subject:   pkix.Name{CommonName: "Test CA"},
				NotBefore: time.Now(),
				NotAfter:  time.Now().AddDate(10, 0, 0),
			},
			wantErr: false,
		},
		{
			name: "missing NotBefore",
			config: CertificateConfig{
				Subject:  pkix.Name{CommonName: "Test CA"},
				NotAfter: time.Now().AddDate(10, 0, 0),
			},
			wantErr: true,
		},
		{
			name: "missing NotAfter",
			config: CertificateConfig{
				Subject:   pkix.Name{CommonName: "Test CA"},
				NotBefore: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "NotAfter before NotBefore",
			config: CertificateConfig{
				Subject:   pkix.Name{CommonName: "Test CA"},
				NotBefore: time.Now().AddDate(1, 0, 0),
				NotAfter:  time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			privateKey, cert, err := GenerateCA(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateCA() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if privateKey == nil {
					t.Error("GenerateCA() privateKey is nil")
				}
				if cert == nil {
					t.Error("GenerateCA() cert is nil")
				}
				if !cert.IsCA {
					t.Error("GenerateCA() cert.IsCA should be true")
				}
				if cert.Subject.CommonName != tt.config.Subject.CommonName {
					t.Errorf("GenerateCA() CommonName = %v, want %v", cert.Subject.CommonName, tt.config.Subject.CommonName)
				}
			}
		})
	}
}

func TestGenerateCAWithRSA(t *testing.T) {
	tests := []struct {
		name    string
		bits    int
		config  CertificateConfig
		wantErr bool
	}{
		{
			name: "valid RSA 2048",
			bits: 2048,
			config: CertificateConfig{
				Subject:   pkix.Name{CommonName: "Test RSA CA"},
				NotBefore: time.Now(),
				NotAfter:  time.Now().AddDate(10, 0, 0),
			},
			wantErr: false,
		},
		{
			name: "RSA too small",
			bits: 1024,
			config: CertificateConfig{
				Subject:   pkix.Name{CommonName: "Test RSA CA"},
				NotBefore: time.Now(),
				NotAfter:  time.Now().AddDate(10, 0, 0),
			},
			wantErr: true,
		},
		{
			name: "missing validity",
			bits: 2048,
			config: CertificateConfig{
				Subject: pkix.Name{CommonName: "Test RSA CA"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			privateKey, cert, err := GenerateCAWithRSA(tt.bits, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateCAWithRSA() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if privateKey == nil {
					t.Error("GenerateCAWithRSA() privateKey is nil")
				}
				if cert == nil {
					t.Error("GenerateCAWithRSA() cert is nil")
				}
				if !cert.IsCA {
					t.Error("GenerateCAWithRSA() cert.IsCA should be true")
				}
			}
		})
	}
}

func TestCreateLicenseCertificate(t *testing.T) {
	// Generate CA first
	caKey, caCert, err := GenerateCA(CertificateConfig{
		Subject:   pkix.Name{CommonName: "Test CA"},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(10, 0, 0),
	})
	if err != nil {
		t.Fatalf("Failed to generate CA: %v", err)
	}

	tests := []struct {
		name    string
		license LicenseInfo
		config  CertificateConfig
		wantErr bool
	}{
		{
			name: "valid license",
			license: LicenseInfo{
				CustomerID: "cust-123",
				MaxApps:    10,
				Features:   []string{"pro", "enterprise"},
			},
			config: CertificateConfig{
				Subject:   pkix.Name{CommonName: "Customer: cust-123"},
				NotBefore: time.Now(),
				NotAfter:  time.Now().AddDate(1, 0, 0),
			},
			wantErr: false,
		},
		{
			name: "license with custom fields",
			license: LicenseInfo{
				CustomerID:   "cust-456",
				MaxApps:      5,
				Features:     []string{"basic"},
				CustomFields: map[string]string{"region": "EU", "tier": "gold"},
			},
			config: CertificateConfig{
				Subject:   pkix.Name{CommonName: "Customer: cust-456"},
				NotBefore: time.Now(),
				NotAfter:  time.Now().AddDate(2, 0, 0),
			},
			wantErr: false,
		},
		{
			name: "missing customerID",
			license: LicenseInfo{
				MaxApps:  10,
				Features: []string{"pro"},
			},
			config: CertificateConfig{
				Subject:   pkix.Name{CommonName: "Invalid"},
				NotBefore: time.Now(),
				NotAfter:  time.Now().AddDate(1, 0, 0),
			},
			wantErr: true,
		},
		{
			name: "negative maxApps",
			license: LicenseInfo{
				CustomerID: "cust-789",
				MaxApps:    -1,
			},
			config: CertificateConfig{
				Subject:   pkix.Name{CommonName: "Invalid"},
				NotBefore: time.Now(),
				NotAfter:  time.Now().AddDate(1, 0, 0),
			},
			wantErr: true,
		},
		{
			name: "missing validity",
			license: LicenseInfo{
				CustomerID: "cust-111",
				MaxApps:    10,
			},
			config: CertificateConfig{
				Subject: pkix.Name{CommonName: "Invalid"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cert, privateKey, err := CreateLicenseCertificate(caKey, caCert, tt.license, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateLicenseCertificate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if cert == nil {
					t.Error("CreateLicenseCertificate() cert is nil")
				}
				if privateKey == nil {
					t.Error("CreateLicenseCertificate() privateKey is nil")
				}
				if cert.Subject.CommonName != tt.config.Subject.CommonName {
					t.Errorf("CreateLicenseCertificate() CommonName = %v, want %v", cert.Subject.CommonName, tt.config.Subject.CommonName)
				}
			}
		})
	}
}

func TestCreateLicenseCertificateWithNilCA(t *testing.T) {
	license := LicenseInfo{
		CustomerID: "cust-123",
		MaxApps:    10,
	}
	config := CertificateConfig{
		Subject:   pkix.Name{CommonName: "Test"},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 0, 0),
	}

	// Test with nil CA cert
	_, _, err := CreateLicenseCertificate(nil, nil, license, config)
	if err == nil {
		t.Error("CreateLicenseCertificate() should fail with nil CA")
	}
}

func TestCreateLicenseCertificateWithNonCA(t *testing.T) {
	// Generate a CA first
	caKey, caCert, err := GenerateCA(CertificateConfig{
		Subject:   pkix.Name{CommonName: "Test CA"},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(10, 0, 0),
	})
	if err != nil {
		t.Fatalf("Failed to generate CA: %v", err)
	}

	// Create a non-CA certificate
	license := LicenseInfo{
		CustomerID: "cust-123",
		MaxApps:    10,
	}
	config := CertificateConfig{
		Subject:   pkix.Name{CommonName: "Non-CA Cert"},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 0, 0),
	}

	nonCACert, _, err := CreateLicenseCertificate(caKey, caCert, license, config)
	if err != nil {
		t.Fatalf("Failed to create license cert: %v", err)
	}

	// Try to use non-CA cert as CA
	_, _, err = CreateLicenseCertificate(caKey, nonCACert, license, config)
	if err == nil {
		t.Error("CreateLicenseCertificate() should fail with non-CA certificate")
	}
}

func TestVerifyLicenseCertificate(t *testing.T) {
	// Generate CA
	caKey, caCert, err := GenerateCA(CertificateConfig{
		Subject:   pkix.Name{CommonName: "Test CA"},
		NotBefore: time.Now().Add(-time.Hour),
		NotAfter:  time.Now().AddDate(10, 0, 0),
	})
	if err != nil {
		t.Fatalf("Failed to generate CA: %v", err)
	}

	// Create valid license certificate
	license := LicenseInfo{
		CustomerID: "cust-123",
		MaxApps:    10,
		Features:   []string{"pro"},
	}
	validCert, _, err := CreateLicenseCertificate(caKey, caCert, license, CertificateConfig{
		Subject:   pkix.Name{CommonName: "Customer: cust-123"},
		NotBefore: time.Now().Add(-time.Hour),
		NotAfter:  time.Now().AddDate(1, 0, 0),
	})
	if err != nil {
		t.Fatalf("Failed to create license cert: %v", err)
	}

	// Create expired certificate
	expiredCert, _, err := CreateLicenseCertificate(caKey, caCert, license, CertificateConfig{
		Subject:   pkix.Name{CommonName: "Expired"},
		NotBefore: time.Now().AddDate(-2, 0, 0),
		NotAfter:  time.Now().AddDate(-1, 0, 0),
	})
	if err != nil {
		t.Fatalf("Failed to create expired cert: %v", err)
	}

	// Create not-yet-valid certificate
	futureCert, _, err := CreateLicenseCertificate(caKey, caCert, license, CertificateConfig{
		Subject:   pkix.Name{CommonName: "Future"},
		NotBefore: time.Now().AddDate(1, 0, 0),
		NotAfter:  time.Now().AddDate(2, 0, 0),
	})
	if err != nil {
		t.Fatalf("Failed to create future cert: %v", err)
	}

	// Generate different CA
	_, differentCACert, err := GenerateCA(CertificateConfig{
		Subject:   pkix.Name{CommonName: "Different CA"},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(10, 0, 0),
	})
	if err != nil {
		t.Fatalf("Failed to generate different CA: %v", err)
	}

	tests := []struct {
		name    string
		cert    *x509.Certificate
		caCert  *x509.Certificate
		wantErr bool
		errType error
	}{
		{
			name:    "valid certificate",
			cert:    validCert,
			caCert:  caCert,
			wantErr: false,
		},
		{
			name:    "expired certificate",
			cert:    expiredCert,
			caCert:  caCert,
			wantErr: true,
			errType: ErrCertificateExpired,
		},
		{
			name:    "not yet valid certificate",
			cert:    futureCert,
			caCert:  caCert,
			wantErr: true,
			errType: ErrCertificateNotYetValid,
		},
		{
			name:    "wrong CA",
			cert:    validCert,
			caCert:  differentCACert,
			wantErr: true,
			errType: ErrVerificationFailed,
		},
		{
			name:    "nil certificate",
			cert:    nil,
			caCert:  caCert,
			wantErr: true,
			errType: ErrInvalidCertificate,
		},
		{
			name:    "nil CA",
			cert:    validCert,
			caCert:  nil,
			wantErr: true,
			errType: ErrInvalidCA,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyLicenseCertificate(tt.cert, tt.caCert)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyLicenseCertificate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExtractLicenseInfo(t *testing.T) {
	// Generate CA
	caKey, caCert, err := GenerateCA(CertificateConfig{
		Subject:   pkix.Name{CommonName: "Test CA"},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(10, 0, 0),
	})
	if err != nil {
		t.Fatalf("Failed to generate CA: %v", err)
	}

	// Create license with all fields
	originalLicense := LicenseInfo{
		CustomerID:   "cust-123",
		MaxApps:      10,
		Features:     []string{"pro", "enterprise", "api"},
		CustomFields: map[string]string{"region": "US", "tier": "platinum"},
	}

	cert, _, err := CreateLicenseCertificate(caKey, caCert, originalLicense, CertificateConfig{
		Subject:   pkix.Name{CommonName: "Customer: cust-123"},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 0, 0),
	})
	if err != nil {
		t.Fatalf("Failed to create license cert: %v", err)
	}

	// Extract license info
	extracted, err := ExtractLicenseInfo(cert)
	if err != nil {
		t.Fatalf("ExtractLicenseInfo() error = %v", err)
	}

	// Verify extracted data
	if extracted.CustomerID != originalLicense.CustomerID {
		t.Errorf("CustomerID = %v, want %v", extracted.CustomerID, originalLicense.CustomerID)
	}
	if extracted.MaxApps != originalLicense.MaxApps {
		t.Errorf("MaxApps = %v, want %v", extracted.MaxApps, originalLicense.MaxApps)
	}
	if len(extracted.Features) != len(originalLicense.Features) {
		t.Errorf("Features length = %v, want %v", len(extracted.Features), len(originalLicense.Features))
	}
	for i, f := range originalLicense.Features {
		if extracted.Features[i] != f {
			t.Errorf("Features[%d] = %v, want %v", i, extracted.Features[i], f)
		}
	}
	if extracted.CustomFields["region"] != originalLicense.CustomFields["region"] {
		t.Errorf("CustomFields[region] = %v, want %v", extracted.CustomFields["region"], originalLicense.CustomFields["region"])
	}
	if extracted.CustomFields["tier"] != originalLicense.CustomFields["tier"] {
		t.Errorf("CustomFields[tier] = %v, want %v", extracted.CustomFields["tier"], originalLicense.CustomFields["tier"])
	}
}

func TestExtractLicenseInfoFromNilCert(t *testing.T) {
	_, err := ExtractLicenseInfo(nil)
	if err != ErrInvalidCertificate {
		t.Errorf("ExtractLicenseInfo(nil) error = %v, want %v", err, ErrInvalidCertificate)
	}
}

func TestExtractLicenseInfoFromCertWithoutExtension(t *testing.T) {
	// Generate CA (CA cert doesn't have license extension)
	_, caCert, err := GenerateCA(CertificateConfig{
		Subject:   pkix.Name{CommonName: "Test CA"},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(10, 0, 0),
	})
	if err != nil {
		t.Fatalf("Failed to generate CA: %v", err)
	}

	_, err = ExtractLicenseInfo(caCert)
	if err != ErrLicenseExtNotFound {
		t.Errorf("ExtractLicenseInfo() error = %v, want %v", err, ErrLicenseExtNotFound)
	}
}

func TestPEMEncodingDecoding(t *testing.T) {
	// Generate CA
	_, caCert, err := GenerateCA(CertificateConfig{
		Subject:   pkix.Name{CommonName: "Test CA"},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(10, 0, 0),
	})
	if err != nil {
		t.Fatalf("Failed to generate CA: %v", err)
	}

	// Encode to PEM
	pemData := EncodeCertificateToPEM(caCert)
	if pemData == nil {
		t.Fatal("EncodeCertificateToPEM() returned nil")
	}

	// Decode from PEM
	decoded, err := DecodeCertificateFromPEM(pemData)
	if err != nil {
		t.Fatalf("DecodeCertificateFromPEM() error = %v", err)
	}

	// Compare
	if decoded.Subject.CommonName != caCert.Subject.CommonName {
		t.Errorf("CommonName = %v, want %v", decoded.Subject.CommonName, caCert.Subject.CommonName)
	}
	if !decoded.NotBefore.Equal(caCert.NotBefore) {
		t.Errorf("NotBefore = %v, want %v", decoded.NotBefore, caCert.NotBefore)
	}
	if !decoded.NotAfter.Equal(caCert.NotAfter) {
		t.Errorf("NotAfter = %v, want %v", decoded.NotAfter, caCert.NotAfter)
	}
}

func TestDecodeCertificateFromInvalidPEM(t *testing.T) {
	tests := []struct {
		name    string
		pemData []byte
	}{
		{
			name:    "empty data",
			pemData: []byte{},
		},
		{
			name:    "invalid PEM",
			pemData: []byte("not a valid PEM"),
		},
		{
			name:    "wrong PEM type",
			pemData: []byte("-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBg==\n-----END PRIVATE KEY-----"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeCertificateFromPEM(tt.pemData)
			if err == nil {
				t.Error("DecodeCertificateFromPEM() should return error for invalid input")
			}
		})
	}
}

func TestValidateLicenseInfo(t *testing.T) {
	tests := []struct {
		name    string
		info    *LicenseInfo
		wantErr bool
	}{
		{
			name: "valid license",
			info: &LicenseInfo{
				CustomerID: "cust-123",
				MaxApps:    10,
			},
			wantErr: false,
		},
		{
			name: "zero maxApps is valid",
			info: &LicenseInfo{
				CustomerID: "cust-123",
				MaxApps:    0,
			},
			wantErr: false,
		},
		{
			name:    "nil license",
			info:    nil,
			wantErr: true,
		},
		{
			name: "empty customerID",
			info: &LicenseInfo{
				CustomerID: "",
				MaxApps:    10,
			},
			wantErr: true,
		},
		{
			name: "negative maxApps",
			info: &LicenseInfo{
				CustomerID: "cust-123",
				MaxApps:    -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLicenseInfo(tt.info)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLicenseInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsLicenseExpired(t *testing.T) {
	// Generate CA
	caKey, caCert, err := GenerateCA(CertificateConfig{
		Subject:   pkix.Name{CommonName: "Test CA"},
		NotBefore: time.Now().AddDate(-1, 0, 0),
		NotAfter:  time.Now().AddDate(10, 0, 0),
	})
	if err != nil {
		t.Fatalf("Failed to generate CA: %v", err)
	}

	license := LicenseInfo{
		CustomerID: "cust-123",
		MaxApps:    10,
	}

	// Valid certificate
	validCert, _, _ := CreateLicenseCertificate(caKey, caCert, license, CertificateConfig{
		Subject:   pkix.Name{CommonName: "Valid"},
		NotBefore: time.Now().Add(-time.Hour),
		NotAfter:  time.Now().AddDate(1, 0, 0),
	})

	// Expired certificate
	expiredCert, _, _ := CreateLicenseCertificate(caKey, caCert, license, CertificateConfig{
		Subject:   pkix.Name{CommonName: "Expired"},
		NotBefore: time.Now().AddDate(-2, 0, 0),
		NotAfter:  time.Now().AddDate(-1, 0, 0),
	})

	if IsLicenseExpired(validCert) {
		t.Error("IsLicenseExpired() should return false for valid cert")
	}
	if !IsLicenseExpired(expiredCert) {
		t.Error("IsLicenseExpired() should return true for expired cert")
	}
	if !IsLicenseExpired(nil) {
		t.Error("IsLicenseExpired() should return true for nil cert")
	}
}

func TestIsLicenseNotYetValid(t *testing.T) {
	// Generate CA
	caKey, caCert, err := GenerateCA(CertificateConfig{
		Subject:   pkix.Name{CommonName: "Test CA"},
		NotBefore: time.Now().AddDate(-1, 0, 0),
		NotAfter:  time.Now().AddDate(10, 0, 0),
	})
	if err != nil {
		t.Fatalf("Failed to generate CA: %v", err)
	}

	license := LicenseInfo{
		CustomerID: "cust-123",
		MaxApps:    10,
	}

	// Valid certificate
	validCert, _, _ := CreateLicenseCertificate(caKey, caCert, license, CertificateConfig{
		Subject:   pkix.Name{CommonName: "Valid"},
		NotBefore: time.Now().Add(-time.Hour),
		NotAfter:  time.Now().AddDate(1, 0, 0),
	})

	// Future certificate
	futureCert, _, _ := CreateLicenseCertificate(caKey, caCert, license, CertificateConfig{
		Subject:   pkix.Name{CommonName: "Future"},
		NotBefore: time.Now().AddDate(1, 0, 0),
		NotAfter:  time.Now().AddDate(2, 0, 0),
	})

	if IsLicenseNotYetValid(validCert) {
		t.Error("IsLicenseNotYetValid() should return false for valid cert")
	}
	if !IsLicenseNotYetValid(futureCert) {
		t.Error("IsLicenseNotYetValid() should return true for future cert")
	}
	if !IsLicenseNotYetValid(nil) {
		t.Error("IsLicenseNotYetValid() should return true for nil cert")
	}
}

func TestHasFeature(t *testing.T) {
	info := &LicenseInfo{
		CustomerID: "cust-123",
		MaxApps:    10,
		Features:   []string{"pro", "enterprise", "api"},
	}

	tests := []struct {
		feature string
		want    bool
	}{
		{"pro", true},
		{"enterprise", true},
		{"api", true},
		{"basic", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.feature, func(t *testing.T) {
			if got := HasFeature(info, tt.feature); got != tt.want {
				t.Errorf("HasFeature(%q) = %v, want %v", tt.feature, got, tt.want)
			}
		})
	}

	// Test with nil info
	if HasFeature(nil, "pro") {
		t.Error("HasFeature(nil, ...) should return false")
	}

	// Test with empty features
	emptyInfo := &LicenseInfo{CustomerID: "cust"}
	if HasFeature(emptyInfo, "pro") {
		t.Error("HasFeature() should return false for empty features")
	}
}

func TestCertificateConfigValidate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		config  CertificateConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: CertificateConfig{
				Subject:   pkix.Name{CommonName: "Test"},
				NotBefore: now,
				NotAfter:  now.AddDate(1, 0, 0),
			},
			wantErr: false,
		},
		{
			name: "zero NotBefore",
			config: CertificateConfig{
				Subject:  pkix.Name{CommonName: "Test"},
				NotAfter: now.AddDate(1, 0, 0),
			},
			wantErr: true,
		},
		{
			name: "zero NotAfter",
			config: CertificateConfig{
				Subject:   pkix.Name{CommonName: "Test"},
				NotBefore: now,
			},
			wantErr: true,
		},
		{
			name: "NotAfter equals NotBefore",
			config: CertificateConfig{
				Subject:   pkix.Name{CommonName: "Test"},
				NotBefore: now,
				NotAfter:  now,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEndToEndLicenseWorkflow(t *testing.T) {
	// 1. Generate CA (done once, stored securely)
	caKey, caCert, err := GenerateCA(CertificateConfig{
		Subject:   pkix.Name{CommonName: "MyApp License CA", Organization: []string{"MyCompany"}},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(10, 0, 0),
	})
	if err != nil {
		t.Fatalf("Failed to generate CA: %v", err)
	}

	// 2. Create license for customer
	license := LicenseInfo{
		CustomerID: "cust-enterprise-001",
		MaxApps:    100,
		Features:   []string{"pro", "enterprise", "api", "sso"},
		CustomFields: map[string]string{
			"companyName": "Enterprise Corp",
			"region":      "US-EAST",
		},
	}

	licenseCert, _, err := CreateLicenseCertificate(caKey, caCert, license, CertificateConfig{
		Subject:   pkix.Name{CommonName: "License: cust-enterprise-001"},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 0, 0),
	})
	if err != nil {
		t.Fatalf("Failed to create license certificate: %v", err)
	}

	// 3. Export certificate to PEM (for distribution)
	certPEM := EncodeCertificateToPEM(licenseCert)
	caCertPEM := EncodeCertificateToPEM(caCert)

	// 4. Customer side: Load and verify certificate
	loadedCert, err := DecodeCertificateFromPEM(certPEM)
	if err != nil {
		t.Fatalf("Failed to load certificate: %v", err)
	}

	loadedCACert, err := DecodeCertificateFromPEM(caCertPEM)
	if err != nil {
		t.Fatalf("Failed to load CA certificate: %v", err)
	}

	// 5. Verify license
	if err := VerifyLicenseCertificate(loadedCert, loadedCACert); err != nil {
		t.Fatalf("License verification failed: %v", err)
	}

	// 6. Extract license info
	extractedLicense, err := ExtractLicenseInfo(loadedCert)
	if err != nil {
		t.Fatalf("Failed to extract license info: %v", err)
	}

	// 7. Use license info
	if extractedLicense.CustomerID != "cust-enterprise-001" {
		t.Errorf("CustomerID mismatch")
	}
	if extractedLicense.MaxApps != 100 {
		t.Errorf("MaxApps mismatch")
	}
	if !HasFeature(extractedLicense, "enterprise") {
		t.Errorf("Should have enterprise feature")
	}
	if !HasFeature(extractedLicense, "sso") {
		t.Errorf("Should have sso feature")
	}
	if HasFeature(extractedLicense, "premium") {
		t.Errorf("Should not have premium feature")
	}
	if extractedLicense.CustomFields["companyName"] != "Enterprise Corp" {
		t.Errorf("CustomField companyName mismatch")
	}
}

func TestRSACAWithECDSALicense(t *testing.T) {
	// Generate RSA CA
	caKey, caCert, err := GenerateCAWithRSA(2048, CertificateConfig{
		Subject:   pkix.Name{CommonName: "RSA CA"},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(10, 0, 0),
	})
	if err != nil {
		t.Fatalf("Failed to generate RSA CA: %v", err)
	}

	// Create ECDSA license certificate signed by RSA CA
	license := LicenseInfo{
		CustomerID: "cust-123",
		MaxApps:    10,
		Features:   []string{"basic"},
	}

	cert, privateKey, err := CreateLicenseCertificate(caKey, caCert, license, CertificateConfig{
		Subject:   pkix.Name{CommonName: "License"},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 0, 0),
	})
	if err != nil {
		t.Fatalf("Failed to create license: %v", err)
	}

	// Verify
	if err := VerifyLicenseCertificate(cert, caCert); err != nil {
		t.Errorf("Verification failed: %v", err)
	}

	// Check that license cert uses ECDSA key
	if privateKey == nil {
		t.Error("License private key should not be nil")
	}

	// Extract license info
	extracted, err := ExtractLicenseInfo(cert)
	if err != nil {
		t.Errorf("Failed to extract license: %v", err)
	}
	if extracted.CustomerID != "cust-123" {
		t.Errorf("CustomerID mismatch")
	}
}
