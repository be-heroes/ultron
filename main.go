package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	ultron "emma.ms/ultron-webhookserver/ultron"
	emmaSdk "github.com/emma-community/emma-go-sdk"
	"github.com/patrickmn/go-cache"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
)

var apiClient *emmaSdk.APIClient = emmaSdk.NewAPIClient(emmaSdk.NewConfiguration())
var credentials = emmaSdk.Credentials{ClientId: os.Getenv("EMMA_CLIENT_ID"), ClientSecret: os.Getenv("EMMA_CLIENT_SECRET")}
var memCache = cache.New(10*time.Minute, 20*time.Minute)

func main() {
	populateCache()

	// TODO: Export the ca bundle to a file for k8s to use
	cert, err := generateSelfSignedCert()
	if err != nil {
		log.Fatalf("Failed to generate self-signed certificate: %v", err)
	}

	server := &http.Server{
		Addr: ":8443",
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
		Handler: http.HandlerFunc(mutatePods),
	}

	log.Println("Starting webhook server with self-signed certificate...")

	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("Failed to listen and serve webhook server: %v", err)
	}
}

func populateCache() {
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

	// TODO: Get Nodes from cluster API server
	activeNodes := []ultron.Node{
		{
			Name: "node1", AvailableCPU: 8, TotalCPU: 16, AvailableMemory: 32, TotalMemory: 64, AvailableStorage: 100,
			DiskType: "SSD", NetworkType: "low-latency", Price: 0.50, MedianPrice: 0.40, Type: "spot", InterruptionRate: 0.2,
		},
		{
			Name: "node2", AvailableCPU: 4, TotalCPU: 8, AvailableMemory: 16, TotalMemory: 32, AvailableStorage: 100,
			DiskType: "HDD", NetworkType: "high-bandwidth", Price: 0.30, MedianPrice: 0.35, Type: "durable", InterruptionRate: 0.01,
		},
	}

	// TODO: Map durableConfigs and spotConfigs
	memCache.Set("activeNodes", activeNodes, cache.DefaultExpiration)
	memCache.Set("durableConfigs", durableConfigs.Content, cache.DefaultExpiration)
	memCache.Set("spotConfigs", spotConfigs.Content, cache.DefaultExpiration)
}

func generateSelfSignedCert() (tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return tls.Certificate{}, err
	}

	certTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"emma"},
			CommonName:   "emma-ultron-webhookserver-service.default.svc",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),

		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certTemplate.DNSNames = []string{
		"emma-ultron-webhookserver-service.default.svc",
		"emma-ultron-webhookserver-service",
		"localhost",
	}

	certTemplate.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}

	certDERBytes, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDERBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, err
	}

	return tlsCert, nil
}

func mutatePods(w http.ResponseWriter, r *http.Request) {
	var admissionReviewReq admissionv1.AdmissionReview
	var admissionReviewResp admissionv1.AdmissionReview

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Could not read request body: %v", err)
		http.Error(w, "could not read request body", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &admissionReviewReq); err != nil {
		log.Printf("Could not unmarshal request: %v", err)
		http.Error(w, "could not unmarshal request", http.StatusBadRequest)
		return
	}

	admissionResponse := handleAdmissionReview(admissionReviewReq.Request)

	admissionReviewResp.Response = admissionResponse
	admissionReviewResp.Kind = admissionReviewReq.Kind
	admissionReviewResp.APIVersion = admissionReviewReq.APIVersion
	admissionReviewResp.Response.UID = admissionReviewReq.Request.UID

	respBytes, err := json.Marshal(admissionReviewResp)
	if err != nil {
		log.Printf("Could not marshal response: %v", err)
		http.Error(w, "could not marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBytes)
}

func handleAdmissionReview(request *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	if request.Kind.Kind != "Pod" {
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	var pod corev1.Pod
	if err := json.Unmarshal(request.Object.Raw, &pod); err != nil {
		log.Printf("Could not unmarshal pod object: %v", err)
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	nodeType := calculateNodeType(pod)

	if pod.Spec.NodeSelector == nil {
		pod.Spec.NodeSelector = make(map[string]string)
	}

	pod.Spec.NodeSelector["node.kubernetes.io/instance-type"] = nodeType
	patchBytes, err := json.Marshal([]map[string]interface{}{
		{
			"op":    "add",
			"path":  "/spec/nodeSelector",
			"value": pod.Spec.NodeSelector,
		},
	})
	if err != nil {
		log.Printf("Could not create patch: %v", err)
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	return &admissionv1.AdmissionResponse{
		Allowed:   true,
		Patch:     patchBytes,
		PatchType: func() *admissionv1.PatchType { pt := admissionv1.PatchTypeJSONPatch; return &pt }(),
	}
}

func calculateNodeType(pod corev1.Pod) string {
	durableConfigsInterface, found := memCache.Get("durableConfigs")

	if !found {
		log.Fatalf("Failed to get durableConfigs from cache")
	}

	_ = durableConfigsInterface.([]emmaSdk.VmConfiguration)

	spotConfigsInterface, found := memCache.Get("spotConfigs")

	if !found {
		log.Fatalf("Failed to get spotConfigs from cache")
	}

	_ = spotConfigsInterface.([]emmaSdk.VmConfiguration)

	activeNodesInterface, found := memCache.Get("activeNodes")

	if !found {
		log.Fatalf("Failed to get activeNodes from cache")
	}

	_ = activeNodesInterface.([]ultron.Node)

	mappedPod, err := mapK8sPodToUltronPod(pod)
	if err != nil {
		log.Fatalf("Error mapping pod: %v", err)
	}

	return ultron.FindBestNode(mappedPod, activeNodesInterface.([]ultron.Node)).Name
}

func mapK8sPodToUltronPod(k8sPod corev1.Pod) (ultron.Pod, error) {
	// Get resource requests and limits
	cpuRequest := k8sPod.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU]
	memRequest := k8sPod.Spec.Containers[0].Resources.Requests[corev1.ResourceMemory]
	cpuLimit := k8sPod.Spec.Containers[0].Resources.Limits[corev1.ResourceCPU]
	memLimit := k8sPod.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory]

	// Convert Kubernetes Quantity types to float64
	cpuRequestFloat, err := strconv.ParseFloat(cpuRequest.AsDec().String(), 64)
	if err != nil {
		return ultron.Pod{}, fmt.Errorf("failed to parse CPU request: %v", err)
	}

	memRequestFloat, err := strconv.ParseFloat(memRequest.AsDec().String(), 64)
	if err != nil {
		return ultron.Pod{}, fmt.Errorf("failed to parse memory request: %v", err)
	}

	cpuLimitFloat, err := strconv.ParseFloat(cpuLimit.AsDec().String(), 64)
	if err != nil {
		return ultron.Pod{}, fmt.Errorf("failed to parse CPU limit: %v", err)
	}

	memLimitFloat, err := strconv.ParseFloat(memLimit.AsDec().String(), 64)
	if err != nil {
		return ultron.Pod{}, fmt.Errorf("failed to parse memory limit: %v", err)
	}

	// For this example, we assume the pod needs "SSD" disk type and "low-latency" network
	// You can customize this based on specific Pod annotations or labels
	requestedDiskType := "SSD"
	requestedNetworkType := "low-latency"
	priority := "HighPriority"

	// Return the mapped Pod
	return ultron.Pod{
		Name:                 k8sPod.Name,
		RequestedCPU:         cpuRequestFloat,
		RequestedMemory:      memRequestFloat,
		RequestedStorage:     10, // Assume a default value for storage
		RequestedDiskType:    requestedDiskType,
		RequestedNetworkType: requestedNetworkType,
		LimitCPU:             cpuLimitFloat,
		LimitMemory:          memLimitFloat,
		Priority:             priority,
	}, nil
}
