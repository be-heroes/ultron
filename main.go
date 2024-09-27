package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	ultron "emma.ms/ultron-webhookserver/ultron"
	emmaSdk "github.com/emma-community/emma-go-sdk"
)

func main() {
	emmaApiCredentials := emmaSdk.Credentials{ClientId: os.Getenv("EMMA_CLIENT_ID"), ClientSecret: os.Getenv("EMMA_CLIENT_SECRET")}
	kubernetesConfigPath := os.Getenv("KUBECONFIG ")
	kubernetesMasterUrl := fmt.Sprintf("tcp://%s:%s", os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT"))

	ultron.InitializeCache(emmaApiCredentials, kubernetesMasterUrl, kubernetesConfigPath)

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
		err = ultron.WriteCACertificateToFile(cert.Certificate[0], certificateExportPath)
		if err != nil {
			log.Fatalf("Failed to write CA certificate to file: %v", err)
		}
	}

	address := os.Getenv("EMMA_WEBHOOKSERVER_ADDRESS")

	if address == "" {
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
