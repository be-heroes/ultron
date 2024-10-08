package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"strings"

	ultron "ultron/internal"
	algorithm "ultron/internal/algorithm"
	handlers "ultron/internal/handlers"
	kubernetes "ultron/internal/kubernetes"
	mapper "ultron/internal/mapper"
	services "ultron/internal/services"

	emma "github.com/emma-community/emma-go-sdk"
)

const (
	EnvironmentVariableKeyKubernetesConfig              = "KUBECONFIG"
	EnvironmentVariableKeyKubernetesServiceHost         = "KUBERNETES_SERVICE_HOST"
	EnvironmentVariableKeyKubernetesServicePort         = "KUBERNETES_SERVICE_PORT"
	EnvironmentVariableKeyEmmaClientId                  = "EMMA_CLIENT_ID"
	EnvironmentVariableKeyEmmaClientSecret              = "EMMA_CLIENT_SECRET"
	EnvironmentVariableKeyServerAddress                 = "ULTRON_SERVER_ADDRESS"
	EnvironmentVariableKeyServerCertificateOrganization = "ULTRON_SERVER_CERTIFICATE_ORGANIZATION"
	EnvironmentVariableKeyServerCertificateCommonName   = "ULTRON_SERVER_CERTIFICATE_COMMON_NAME"
	EnvironmentVariableKeyServerCertificateDnsNames     = "ULTRON_SERVER_CERTIFICATE_DNS_NAMES"
	EnvironmentVariableKeyServerCertificateIpAddresses  = "ULTRON_SERVER_CERTIFICATE_IP_ADDRESSES"
	EnvironmentVariableKeyServerCertificateExportPath   = "ULTRON_SERVER_CERTIFICATE_EXPORT_PATH"
)

func main() {
	kubernetesConfigPath := os.Getenv(EnvironmentVariableKeyKubernetesConfig)
	kubernetesMasterUrl := fmt.Sprintf("tcp://%s:%s", os.Getenv(EnvironmentVariableKeyKubernetesServiceHost), os.Getenv(EnvironmentVariableKeyKubernetesServicePort))
	emmaApiCredentials := emma.Credentials{ClientId: os.Getenv(EnvironmentVariableKeyEmmaClientId), ClientSecret: os.Getenv(EnvironmentVariableKeyEmmaClientSecret)}
	mapper := mapper.NewIMapper()
	algorithm := algorithm.NewIAlgorithm()
	cacheService := services.NewICacheService(nil, nil)
	certificateService := services.NewICertificateService()
	computeService := services.NewIComputeService(algorithm, cacheService, mapper)
	mutationHandler := handlers.NewIMutationHandler(computeService)
	kubernetesClient := kubernetes.NewIKubernetesClient(kubernetesMasterUrl, kubernetesConfigPath, mapper, computeService)

	// TODO: Move cache initialization to ultron-attendant
	log.Println("Initializing cache")

	apiClient := emma.NewAPIClient(emma.NewConfiguration())
	token, resp, err := apiClient.AuthenticationAPI.IssueToken(context.Background()).Credentials(emmaApiCredentials).Execute()
	if err != nil {
		log.Fatalf("Failed to issue access token with error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		_, err := io.ReadAll(resp.Body)

		log.Fatalf("Failed to read access token data with error: %v", err)
	}

	auth := context.WithValue(context.Background(), emma.ContextAccessToken, token.GetAccessToken())
	durableConfigs, resp, err := apiClient.ComputeInstancesConfigurationsAPI.GetVmConfigs(auth).Size(math.MaxInt32).Execute()
	if err != nil {
		log.Fatalf("Failed to fetch durable compute configurations with error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		_, err := io.ReadAll(resp.Body)

		log.Fatalf("Failed to read durable compute configurations data with error: %v", err)
	}

	ephemeralConfigs, resp, err := apiClient.ComputeInstancesConfigurationsAPI.GetSpotConfigs(auth).Size(math.MaxInt32).Execute()
	if err != nil {
		log.Fatalf("Failed to fetch ephemeral compute configurations with error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		_, err := io.ReadAll(resp.Body)

		log.Fatalf("Failed to read ephemeral compute configurations data with error: %v", err)
	}

	cacheService.AddCacheItem(ultron.CacheKeyDurableVmConfigurations, durableConfigs.Content, 0)
	cacheService.AddCacheItem(ultron.CacheKeySpotVmConfigurations, ephemeralConfigs.Content, 0)

	wNodes, err := kubernetesClient.GetWeightedNodes()
	if err != nil {
		log.Fatalf("Failed to get weighted nodes with error: %v", err)
	}

	cacheService.AddCacheItem(ultron.CacheKeyWeightedNodes, wNodes, 0)

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

	server := &http.Server{
		Addr: serverAddress,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
		Handler: http.HandlerFunc(mutationHandler.MutatePods),
	}

	log.Println("Initialized server")

	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("Failed to listen and serve server: %v", err)
	}
}
