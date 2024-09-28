package ultron

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

const (
	BlockTypeCertificate   = "CERTIFICATE"
	BlockTypeRsaPrivateKey = "RSA PRIVATE KEY"
)

func GenerateSelfSignedCert(organization string, commonName string, dnsNames []string, ipAddresses []net.IP) (tls.Certificate, error) {
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

func ExportCACert(caCert []byte, filePath string) error {
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

	log.Printf("CA certificate written to %s", filePath)

	return nil
}
