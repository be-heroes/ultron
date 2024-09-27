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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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

	wNodes, err := getNodeDataFromK8s()
	if err != nil {
		log.Fatalf("Failed to fetch nodes: %v", err)
	}

	ultron.Cache.Set("weightedNodes", wNodes, cache.DefaultExpiration)
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

func getNodeDataFromK8s() ([]ultron.WeightedNode, error) {
	var config *rest.Config
	var err error

	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		kubeconfig := os.Getenv("KUBECONFIG")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var weightedNodes []ultron.WeightedNode

	for _, node := range nodes.Items {
		cpuAllocatable := node.Status.Allocatable[v1.ResourceCPU]
		memAllocatable := node.Status.Allocatable[v1.ResourceMemory]
		storageAllocatable := node.Status.Allocatable[v1.ResourceEphemeralStorage]
		availableCPU := cpuAllocatable.AsApproximateFloat64()
		availableMemory := (float64)(memAllocatable.Value() / (1024 * 1024 * 1024))
		availableStorage := (float64)(storageAllocatable.Value() / (1024 * 1024 * 1024))

		// TODO: Extract relevant labels for the node (e.g., node type or network configuration)
		// TODO: Implement logic to infer/fetch missing values or resort to sensible defaults
		hostname := node.Labels["kubernetes.io/hostname"]
		nodeType := "durable"
		diskType := "SSD"
		networkType := "isolated"
		nodePrice := 0.30
		nodeMedianPrice := 0.25
		nodeInteruptionRate := 0.05

		weightedNode := ultron.WeightedNode{
			Selector:         hostname,
			AvailableCPU:     availableCPU,
			TotalCPU:         availableCPU,    // TODO: Change usage of allocatable CPU as total CPU
			AvailableMemory:  availableMemory, // TODO: Change usage of allocatable memory as total memory
			TotalMemory:      availableMemory,
			AvailableStorage: availableStorage,
			DiskType:         diskType,
			NetworkType:      networkType,
			Price:            nodePrice,
			MedianPrice:      nodeMedianPrice,
			Type:             nodeType,
			InterruptionRate: nodeInteruptionRate,
		}

		weightedNodes = append(weightedNodes, weightedNode)
	}

	return weightedNodes, nil
}
