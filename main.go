package main

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	handlers "github.com/be-heroes/ultron/internal/handlers"
	ultron "github.com/be-heroes/ultron/pkg"
	algorithm "github.com/be-heroes/ultron/pkg/algorithm"
	mapper "github.com/be-heroes/ultron/pkg/mapper"
	services "github.com/be-heroes/ultron/pkg/services"
	"github.com/redis/go-redis/v9"
)

func main() {
	var redisClient *redis.Client

	redisServerAddress := os.Getenv(ultron.EnvRedisServerAddress)
	redisServerDatabase := os.Getenv(ultron.EnvRedisServerDatabase)
	redisServerDatabaseInt, err := strconv.Atoi(redisServerDatabase)
	if err != nil {
		redisServerDatabaseInt = 0
	}

	if redisServerAddress != "" {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     redisServerAddress,
			Password: os.Getenv(ultron.EnvRedisServerPassword),
			DB:       redisServerDatabaseInt,
		})
	}

	mapper := mapper.NewIMapper()
	algorithm := algorithm.NewIAlgorithm()
	cacheService := services.NewICacheService(nil, redisClient)
	certificateService := services.NewICertificateService()
	computeService := services.NewIComputeService(algorithm, cacheService, mapper)
	mutationHandler := handlers.NewIMutationHandler(computeService)
	validationHandler := handlers.NewIValidationHandler(computeService)

	log.Println("Initializing server")

	serverAddress := os.Getenv(ultron.EnvServerAddress)
	if serverAddress == "" {
		serverAddress = ":8443"
	}

	certificateOrganization := os.Getenv(ultron.EnvServerCertificateOrganization)
	if certificateOrganization == "" {
		certificateOrganization = "be-heroes"
	}

	certificateCommonName := os.Getenv(ultron.EnvServerCertificateCommonName)
	if certificateCommonName == "" {
		certificateCommonName = "ultron-service.default.svc"
	}

	certificateDnsNamesCSV := os.Getenv(ultron.EnvServerCertificateDnsNames)
	if certificateDnsNamesCSV == "" {
		certificateDnsNamesCSV = "ultron-service.default.svc,ultron-service,localhost"
	}

	certificateIpAddressesCSV := os.Getenv(ultron.EnvServerCertificateIpAddresses)
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

	certificateExportPath := os.Getenv(ultron.EnvServerCertificateExportPath)
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
