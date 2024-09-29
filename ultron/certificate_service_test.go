package ultron

// import (
// 	"crypto/x509"
// 	"encoding/pem"
// 	"net"
// 	"os"
// 	"testing"
// )

// func TestGenerateSelfSignedCert(t *testing.T) {
// 	organization := "Test Organization"
// 	commonName := "test.com"
// 	dnsNames := []string{"test.com", "www.test.com"}
// 	ipAddresses := []net.IP{net.ParseIP("127.0.0.1")}

// 	cert, err := GenerateSelfSignedCert(organization, commonName, dnsNames, ipAddresses)
// 	if err != nil {
// 		t.Fatalf("Failed to generate self-signed certificate: %v", err)
// 	}

// 	if len(cert.Certificate) == 0 {
// 		t.Fatal("Expected a non-empty certificate, got an empty one")
// 	}

// 	parsedCert, err := x509.ParseCertificate(cert.Certificate[0])
// 	if err != nil {
// 		t.Fatalf("Failed to parse generated certificate: %v", err)
// 	}

// 	if parsedCert.Subject.Organization[0] != organization {
// 		t.Errorf("Expected organization %s, got %s", organization, parsedCert.Subject.Organization[0])
// 	}
// 	if parsedCert.Subject.CommonName != commonName {
// 		t.Errorf("Expected common name %s, got %s", commonName, parsedCert.Subject.CommonName)
// 	}

// 	if len(parsedCert.DNSNames) != 2 || parsedCert.DNSNames[0] != "test.com" {
// 		t.Errorf("Expected DNSNames to include 'test.com', got %v", parsedCert.DNSNames)
// 	}

// 	if len(parsedCert.IPAddresses) != 1 || !parsedCert.IPAddresses[0].Equal(net.ParseIP("127.0.0.1")) {
// 		t.Errorf("Expected IPAddresses to include 127.0.0.1, got %v", parsedCert.IPAddresses)
// 	}
// }

// func TestGenerateSelfSignedCert_Error(t *testing.T) {
// 	_, err := GenerateSelfSignedCert("", "", nil, nil)
// 	if err == nil {
// 		t.Error("Expected error but got none")
// 	}
// }

// func TestExportCACert(t *testing.T) {
// 	organization := "Test Organization"
// 	commonName := "test.com"
// 	dnsNames := []string{"test.com"}
// 	ipAddresses := []net.IP{net.ParseIP("127.0.0.1")}

// 	cert, err := GenerateSelfSignedCert(organization, commonName, dnsNames, ipAddresses)
// 	if err != nil {
// 		t.Fatalf("Failed to generate self-signed certificate for testing: %v", err)
// 	}

// 	certPEM := pem.EncodeToMemory(&pem.Block{Type: BlockTypeCertificate, Bytes: cert.Certificate[0]})

// 	tempFilePath := "ca_test_cert.pem"
// 	defer os.Remove(tempFilePath)

// 	err = ExportCACert(certPEM, tempFilePath)
// 	if err != nil {
// 		t.Fatalf("Failed to export CA certificate: %v", err)
// 	}

// 	_, err = os.Stat(tempFilePath)
// 	if os.IsNotExist(err) {
// 		t.Fatalf("CA certificate file not created: %v", err)
// 	}
// }

// func TestExportCACert_Error(t *testing.T) {
// 	err := ExportCACert(nil, "invalid_cert.pem")
// 	if err == nil {
// 		t.Error("Expected an error when passing nil CA certificate, but got none")
// 	}
// }
