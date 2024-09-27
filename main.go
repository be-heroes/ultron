package main

import (
	"context"
	"crypto/tls"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"

	ultron "emma.ms/ultron-webhookserver/ultron"
	emmaSdk "github.com/emma-community/emma-go-sdk"
	"github.com/patrickmn/go-cache"
)

func main() {
	populateCache()

	cert, err := ultron.GenerateSelfSignedCert(
		"emma",
		"emma-ultron-webhookserver-service.default.svc",
		[]string{"emma-ultron-webhookserver-service.default.svc", "emma-ultron-webhookserver-service", "localhost"},
		[]net.IP{net.ParseIP("127.0.0.1")})
	if err != nil {
		log.Fatalf("Failed to generate self-signed certificate: %v", err)
	}

	certificateExportPath := os.Getenv("EMMA_WEBHOOKSERVER_CERTIFICATE_EXPORT_PATH")

	if certificateExportPath != "" {
		err = writeCACertificateToFile(cert.Certificate[0], certificateExportPath)
		if err != nil {
			log.Fatalf("Failed to write CA certificate to file: %v", err)
		}
	}

	var address string

	if os.Getenv("EMMA_WEBHOOKSERVER_ADDRESS") != "" {
		address = os.Getenv("EMMA_WEBHOOKSERVER_ADDRESS")
	} else {
		address = ":8443"
	}

	server := &http.Server{
		Addr: address,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
		Handler: http.HandlerFunc(ultron.MutatePods),
	}

	log.Println("Starting webhook server with self-signed certificate...")

	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("Failed to listen and serve webhook server: %v", err)
	}
}

func populateCache() {
	apiClient := emmaSdk.NewAPIClient(emmaSdk.NewConfiguration())
	credentials := emmaSdk.Credentials{ClientId: os.Getenv("EMMA_CLIENT_ID"), ClientSecret: os.Getenv("EMMA_CLIENT_SECRET")}
	token, resp, err := apiClient.AuthenticationAPI.IssueToken(context.Background()).Credentials(credentials).Execute()

	if err != nil {
		log.Fatalf("Failed to fetch token: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		log.Fatalf("Failed to fetch token: %v", string(body))
	}

	auth := context.WithValue(context.Background(), emmaSdk.ContextAccessToken, token.GetAccessToken())
	durableConfigs, resp, err := apiClient.ComputeInstancesConfigurationsAPI.GetVmConfigs(auth).Execute()

	if err != nil {
		log.Fatalf("Failed to fetch vm configs: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		log.Fatalf("Failed to fetch vm configs: %v", string(body))
	}

	spotConfigs, resp, err := apiClient.ComputeInstancesConfigurationsAPI.GetSpotConfigs(auth).Execute()

	if err != nil {
		log.Fatalf("Failed to fetch spot configs: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		log.Fatalf("Failed to fetch spot configs: %v", string(body))
	}

	// TODO: Get Nodes from actual API server
	weightedNodes := []ultron.WeightedNode{
		{
			Selector: "kubernetes.io/hostname:node1", AvailableCPU: 8, TotalCPU: 16, AvailableMemory: 32, TotalMemory: 64, AvailableStorage: 100,
			DiskType: "SSD", NetworkType: "isolated", Price: 0.50, MedianPrice: 0.40, Type: "spot", InterruptionRate: 0.2,
		},
		{
			Selector: "kubernetes.io/hostname:node2", AvailableCPU: 4, TotalCPU: 8, AvailableMemory: 16, TotalMemory: 32, AvailableStorage: 100,
			DiskType: "SSDPlus", NetworkType: "multi-cloud", Price: 0.30, MedianPrice: 0.35, Type: "durable", InterruptionRate: 0.01,
		},
	}

	ultron.Cache.Set("weightedNodes", weightedNodes, cache.DefaultExpiration)
	ultron.Cache.Set("durableConfigs", durableConfigs.Content, cache.DefaultExpiration)
	ultron.Cache.Set("spotConfigs", spotConfigs.Content, cache.DefaultExpiration)
}

func writeCACertificateToFile(caCert []byte, filePath string) error {
	certPEMBlock := pem.EncodeToMemory(&pem.Block{
		Type:  ultron.CERTIFICATE_BLOCK_TYPE,
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
