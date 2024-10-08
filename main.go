package main

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	handlers "github.com/be-heroes/ultron/internal/handlers"
	algorithm "github.com/be-heroes/ultron/pkg/algorithm"
	mapper "github.com/be-heroes/ultron/pkg/mapper"
	services "github.com/be-heroes/ultron/pkg/services"
)

const (
	EnvironmentVariableKeyKubernetesConfig              = "KUBECONFIG"
	EnvironmentVariableKeyKubernetesServiceHost         = "KUBERNETES_SERVICE_HOST"
	EnvironmentVariableKeyKubernetesServicePort         = "KUBERNETES_SERVICE_PORT"
	EnvironmentVariableKeyServerAddress                 = "ULTRON_SERVER_ADDRESS"
	EnvironmentVariableKeyServerCertificateOrganization = "ULTRON_SERVER_CERTIFICATE_ORGANIZATION"
	EnvironmentVariableKeyServerCertificateCommonName   = "ULTRON_SERVER_CERTIFICATE_COMMON_NAME"
	EnvironmentVariableKeyServerCertificateDnsNames     = "ULTRON_SERVER_CERTIFICATE_DNS_NAMES"
	EnvironmentVariableKeyServerCertificateIpAddresses  = "ULTRON_SERVER_CERTIFICATE_IP_ADDRESSES"
	EnvironmentVariableKeyServerCertificateExportPath   = "ULTRON_SERVER_CERTIFICATE_EXPORT_PATH"
)

func main() {
	mapper := mapper.NewIMapper()
	algorithm := algorithm.NewIAlgorithm()
	cacheService := services.NewICacheService(nil, nil)
	certificateService := services.NewICertificateService()
	computeService := services.NewIComputeService(algorithm, cacheService, mapper)
	mutationHandler := handlers.NewIMutationHandler(computeService)
	validationHandler := handlers.NewIValidationHandler(computeService)

	log.Println("Initializing server")

	serverAddress := os.Getenv(EnvironmentVariableKeyServerAddress)
	if serverAddress == "" {
		serverAddress = ":8443"
	}

	certificateOrganization := os.Getenv(EnvironmentVariableKeyServerCertificateOrganization)
	if certificateOrganization == "" {
		certificateOrganization = "be-heroes"
	}

	certificateCommonName := os.Getenv(EnvironmentVariableKeyServerCertificateCommonName)
	if certificateCommonName == "" {
		certificateCommonName = "ultron-service.default.svc"
	}

	certificateDnsNamesCSV := os.Getenv(EnvironmentVariableKeyServerCertificateDnsNames)
	if certificateDnsNamesCSV == "" {
		certificateDnsNamesCSV = "ultron-service.default.svc,ultron-service,localhost"
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

	cert, err := certificateService.GenerateSelfSignedCert(
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

		err = certificateService.ExportCACert(cert.Certificate[0], certificateExportPath)
		if err != nil {
			log.Fatalf("Failed to export CA certificate to file: %v", err)
		}

		log.Println("Exported CA certificate")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/mutate", mutationHandler.MutatePodSpec)
	mux.HandleFunc("/validate", validationHandler.ValidatePodSpec)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	server := &http.Server{
		Addr: serverAddress,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
		Handler: mux,
	}

	log.Println("Initialized server")

	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("Failed to listen and serve server: %v", err)
	}
}
