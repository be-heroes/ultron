package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	ultron "emma.ms/ultron-webhookserver/ultron"
	emma "github.com/emma-community/emma-go-sdk"
)

const (
	EnvironmentVariableKeyKubernetesConfig              = "KUBECONFIG"
	EnvironmentVariableKeyKubernetesServiceHost         = "KUBERNETES_SERVICE_HOST"
	EnvironmentVariableKeyKubernetesServicePort         = "KUBERNETES_SERVICE_PORT"
	EnvironmentVariableKeyEmmaClientId                  = "EMMA_CLIENT_ID"
	EnvironmentVariableKeyEmmaClientSecret              = "EMMA_CLIENT_SECRET"
	EnvironmentVariableKeyServerAddress                 = "SERVER_ADDRESS"
	EnvironmentVariableKeyServerCertificateOrganization = "SERVER_CERTIFICATE_ORGANIZATION"
	EnvironmentVariableKeyServerCertificateCommonName   = "SERVER_CERTIFICATE_COMMON_NAME"
	EnvironmentVariableKeyServerCertificateDnsNames     = "SERVER_CERTIFICATE_DNS_NAMES"
	EnvironmentVariableKeyServerCertificateIpAddresses  = "SERVER_CERTIFICATE_IP_ADDRESSES"
	EnvironmentVariableKeyServerCertificateExportPath   = "SERVER_CERTIFICATE_EXPORT_PATH"
)

func main() {
	kubernetesConfigPath := os.Getenv(EnvironmentVariableKeyKubernetesConfig)
	kubernetesMasterUrl := fmt.Sprintf("tcp://%s:%s", os.Getenv(EnvironmentVariableKeyKubernetesServiceHost), os.Getenv(EnvironmentVariableKeyKubernetesServicePort))
	emmaApiCredentials := emma.Credentials{ClientId: os.Getenv(EnvironmentVariableKeyEmmaClientId), ClientSecret: os.Getenv(EnvironmentVariableKeyEmmaClientSecret)}

	log.Println("Initializing cache")

	err := ultron.InitializeCache(emmaApiCredentials, kubernetesMasterUrl, kubernetesConfigPath)
	if err != nil {
		log.Fatalf("Failed to initialize cache with error: %v", err)
	}

	log.Println("Initialized cache")
	log.Println("Initializing server")

	serverAddress := os.Getenv(EnvironmentVariableKeyServerAddress)
	if serverAddress == "" {
		serverAddress = ":8443"
	}

	certificateOrganization := os.Getenv(EnvironmentVariableKeyServerCertificateOrganization)
	if certificateOrganization == "" {
		certificateOrganization = "emma"
	}

	certificateCommonName := os.Getenv(EnvironmentVariableKeyServerCertificateCommonName)
	if certificateCommonName == "" {
		certificateCommonName = "emma-ultron-webhookserver-service.default.svc"
	}

	certificateDnsNamesCSV := os.Getenv(EnvironmentVariableKeyServerCertificateDnsNames)
	if certificateDnsNamesCSV == "" {
		certificateDnsNamesCSV = "emma-ultron-webhookserver-service.default.svc,emma-ultron-webhookserver-service,localhost"
	}

	certificateIpAddressesCSV := os.Getenv(EnvironmentVariableKeyServerCertificateIpAddresses)
	if certificateIpAddressesCSV == "" {
		certificateIpAddressesCSV = "127.0.0.1"
	}

	var certificateIpAddresses []net.IP
	for _, ipAddress := range strings.Split(certificateIpAddressesCSV, ",") {
		certificateIpAddresses = append(certificateIpAddresses, net.ParseIP(ipAddress))
	}

	log.Println("Generating self-signed certificate")

	cert, err := ultron.GenerateSelfSignedCert(
		certificateOrganization,
		certificateCommonName,
		strings.Split(certificateDnsNamesCSV, ","),
		certificateIpAddresses)
	if err != nil {
		log.Fatalf("Failed to generate self-signed certificate: %v", err)
	}

	log.Println("Generated self-signed certificate")

	certificateExportPath := os.Getenv(EnvironmentVariableKeyServerCertificateExportPath)
	if certificateExportPath != "" {
		log.Println("Exporting CA certificate")

		err = ultron.ExportCACert(cert.Certificate[0], certificateExportPath)
		if err != nil {
			log.Fatalf("Failed to export CA certificate to file: %v", err)
		}

		log.Println("Exported CA certificate")
	}

	server := &http.Server{
		Addr: serverAddress,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
		Handler: http.HandlerFunc(ultron.MutatePods),
	}

	log.Println("Initialized server")

	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("Failed to listen and serve server: %v", err)
	}
}
