package main

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"strings"

	"go.uber.org/zap"

	handlers "github.com/be-heroes/ultron/internal/handlers"
	ultron "github.com/be-heroes/ultron/pkg"
	algorithm "github.com/be-heroes/ultron/pkg/algorithm"
	mapper "github.com/be-heroes/ultron/pkg/mapper"
	services "github.com/be-heroes/ultron/pkg/services"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	sugar := logger.Sugar()
	sugar.Info("Initializing Ultron")

	config, err := ultron.LoadConfig()
	if err != nil {
		sugar.Fatalf("Failed to load configuration: %v", err)
	}

	ctx := context.Background()
	redisClient := ultron.InitializeRedisClientFromConfig(ctx, config, sugar)
	mapperInstance := mapper.NewMapper()
	algorithmInstance := algorithm.NewAlgorithm()
	cacheService := services.NewCacheService(nil, redisClient)
	certificateService := services.NewCertificateService()
	computeService := services.NewComputeService(algorithmInstance, cacheService, mapperInstance)
	mutationHandler := handlers.NewMutationHandler(computeService)
	validationHandler := handlers.NewValidationHandler(computeService, redisClient)

	sugar.Info("Initialized Ultron")
	sugar.Info("Generating self-signed certificate")

	cert, err := certificateService.GenerateSelfSignedCert(
		config.CertificateOrganization,
		config.CertificateCommonName,
		strings.Split(config.CertificateDnsNamesCSV, ","),
		parseCertificateIpAddresses(config.CertificateIpAddressesCSV),
	)
	if err != nil {
		sugar.Fatalf("Failed to generate self-signed certificate: %v", err)
	}

	sugar.Info("Generated self-signed certificate")

	if config.CertificateExportPath != "" {
		sugar.Info("Exporting CA certificate")

		err = certificateService.ExportCACert(cert.Certificate[0], config.CertificateExportPath)
		if err != nil {
			sugar.Fatalf("Failed to export CA certificate to file: %v", err)
		}

		sugar.Info("Exported CA certificate")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", mutationHandler.MutatePodSpec)
	mux.HandleFunc("/validate", validationHandler.ValidatePodSpec)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	sugar.Info("Starting Ultron")

	server := &http.Server{
		Addr: config.ServerAddress,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
		Handler: mux,
	}

	if err := server.ListenAndServeTLS("", ""); err != nil {
		sugar.Fatalf("Failed to listen and serve Ultron: %v", err)
	}
}

func parseCertificateIpAddresses(csv string) []net.IP {
	var certificateIpAddresses []net.IP

	for _, ipAddress := range strings.Split(csv, ",") {
		if ipAddress != "" {
			certificateIpAddresses = append(certificateIpAddresses, net.ParseIP(ipAddress))
		}
	}

	return certificateIpAddresses
}
