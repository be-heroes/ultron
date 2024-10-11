package services_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"math/big"
	"net"
	"os"
	"testing"
	"time"

	services "github.com/be-heroes/ultron/pkg/services"
	"github.com/stretchr/testify/assert"
)

func TestGenerateSelfSignedCert_Success(t *testing.T) {
	// Arrange
	certService := services.NewCertificateService()
	organization := "TestOrg"
	commonName := "test.com"
	dnsNames := []string{"test.com", "www.test.com"}
	ipAddresses := []net.IP{net.ParseIP("127.0.0.1")}

	// Act
	cert, err := certService.GenerateSelfSignedCert(organization, commonName, dnsNames, ipAddresses)

	// Assert
	assert.NoError(t, err, "GenerateSelfSignedCert should not return an error")
	assert.NotEmpty(t, cert.Certificate, "Expected certificate to be generated, but got none")
	assert.NotNil(t, cert.PrivateKey, "Expected private key to be generated, but got nil")
}

func TestGenerateSelfSignedCert_MissingOrgAndCommonName(t *testing.T) {
	// Arrange
	certService := services.NewCertificateService()
	organization := ""
	commonName := ""
	dnsNames := []string{"test.com"}
	ipAddresses := []net.IP{net.ParseIP("127.0.0.1")}

	// Act
	_, err := certService.GenerateSelfSignedCert(organization, commonName, dnsNames, ipAddresses)

	// Assert
	assert.Error(t, err, "Expected error for missing organization and common name, but got none")
}

func TestGenerateSelfSignedCert_EmptyDNSAndIP(t *testing.T) {
	// Arrange
	certService := services.NewCertificateService()
	organization := "TestOrg"
	commonName := "test.com"
	dnsNames := []string{}
	ipAddresses := []net.IP{}

	// Act
	cert, err := certService.GenerateSelfSignedCert(organization, commonName, dnsNames, ipAddresses)

	// Assert
	assert.NoError(t, err, "GenerateSelfSignedCert should not return an error")
	assert.NotEmpty(t, cert.Certificate, "Expected certificate to be generated, but got none")
}

func TestExportCACert_Success(t *testing.T) {
	// Arrange
	certService := services.NewCertificateService()

	// Act & Assert
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
	}
	caCertDER, _ := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)

	filePath := "test_ca_cert.pem"
	err := certService.ExportCACert(caCertDER, filePath)

	// Assert
	assert.NoError(t, err, "ExportCACert should not return an error")

	_, err = os.Stat(filePath)
	assert.NoError(t, err, "Expected CA certificate file to be created, but it does not exist")

	// Cleanup
	os.Remove(filePath)
}

func TestExportCACert_NilCert(t *testing.T) {
	// Arrange
	certService := services.NewCertificateService()

	// Act
	err := certService.ExportCACert(nil, "dummy.pem")

	// Assert
	assert.Error(t, err, "Expected error for nil certificate, but got none")
}

func TestExportCACert_FailToWriteFile(t *testing.T) {
	// Arrange
	certService := services.NewCertificateService()

	// Act
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
	}
	caCertDER, _ := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)

	filePath := "/invalid_path/test_ca_cert.pem"
	err := certService.ExportCACert(caCertDER, filePath)

	// Assert
	assert.Error(t, err, "Expected error for invalid file path, but got none")
}
