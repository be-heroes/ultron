package ultron

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"
)

const (
	BlockTypeCertificate   = "CERTIFICATE"
	BlockTypeRsaPrivateKey = "RSA PRIVATE KEY"
)

type CertificateService interface {
	GenerateSelfSignedCert(organization string, commonName string, dnsNames []string, ipAddresses []net.IP) (tls.Certificate, error)
	ExportCACert(caCert []byte, filePath string) error
}

type ICertificateService struct {
}

func NewICertificateService() *ICertificateService {
	return &ICertificateService{}
}

func (cs ICertificateService) GenerateSelfSignedCert(organization string, commonName string, dnsNames []string, ipAddresses []net.IP) (tls.Certificate, error) {
	if organization == "" || commonName == "" {
		return tls.Certificate{}, fmt.Errorf("organization and common name must be provided")
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return tls.Certificate{}, err
	}

	certTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{organization},
			CommonName:   commonName,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),

		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              dnsNames,
		IPAddresses:           ipAddresses,
	}

	certDERBytes, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: BlockTypeCertificate, Bytes: certDERBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: BlockTypeRsaPrivateKey, Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, err
	}

	return tlsCert, nil
}

func (cs ICertificateService) ExportCACert(caCert []byte, filePath string) error {
	if caCert == nil {
		return fmt.Errorf("CA certificate is nil")
	}

	certPEMBlock := pem.EncodeToMemory(&pem.Block{
		Type:  BlockTypeCertificate,
		Bytes: caCert,
	})

	if certPEMBlock == nil {
		return fmt.Errorf("failed to encode certificate to PEM format")
	}

	err := os.WriteFile(filePath, certPEMBlock, 0644)
	if err != nil {
		return fmt.Errorf("failed to write CA certificate to file: %w", err)
	}

	return nil
}
