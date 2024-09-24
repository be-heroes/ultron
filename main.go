package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"time"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
)

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

	nodeType := callExternalAPI(pod)

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

func callExternalAPI(pod corev1.Pod) string {
	if val, ok := pod.Labels["high-performance"]; ok && val == "true" {
		return "high-performance-node"
	}
	return "custom-node"
}

func main() {
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

func generateSelfSignedCert() (tls.Certificate, error) {
	// Generate a private key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	// Create a certificate template
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

	// Set the DNS names and IP addresses the certificate will be valid for
	certTemplate.DNSNames = []string{
		"emma-ultron-webhookserver-service.default.svc",
		"emma-ultron-webhookserver-service",
		"localhost",
	}

	certTemplate.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}

	// Create the certificate
	certDERBytes, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	// Encode the certificate and private key to PEM format
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDERBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	// Load the certificate into a tls.Certificate
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, err
	}

	return tlsCert, nil
}
