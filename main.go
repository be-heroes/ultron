package main

import (
	"context"
	"crypto/tls"
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
		sugar.Fatalf("Failed to load Ultron configuration: %v", err)
	}

	ctx := context.Background()
	redisClient := ultron.InitializeRedisClientFromConfig(ctx, config, sugar)
	mapper := mapper.NewMapper()
	algorithm := algorithm.NewAlgorithm()
	cacheService := services.NewCacheService(nil, redisClient)
	certificateService := services.NewCertificateService()
	computeService := services.NewComputeService(algorithm, cacheService, mapper)
	mutationHandler := handlers.NewMutationHandler(computeService)
	validationHandler := handlers.NewValidationHandler(computeService, mapper, redisClient)

	sugar.Info("Initialized Ultron")
	sugar.Info("Generating self-signed certificate")

	cert, err := certificateService.GenerateSelfSignedCert(
		config.CertificateOrganization,
		config.CertificateCommonName,
		strings.Split(config.CertificateDnsNamesCSV, ","),
		ultron.ParseCsvIpAddressString(config.CertificateIpAddressesCSV),
	)
	if err != nil {
		sugar.Fatalf("Failed to generate self-signed certificate: %v", err)
	}

	sugar.Info("Generated self-signed certificate")

	if config.CertificateExportPath != "" {
		sugar.Info("Exporting CA certificate to path: %s", config.CertificateExportPath)

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
		sugar.Fatalf("Failed to start Ultron: %v", err)
	}
}
