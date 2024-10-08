package handlers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	handlers "github.com/be-heroes/ultron/internal/handlers"
	ultron "github.com/be-heroes/ultron/pkg"

	emma "github.com/emma-community/emma-go-sdk"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func intPtr(i int32) *int32         { return &i }
func float32Ptr(f float32) *float32 { return &f }
func stringPtr(s string) *string    { return &s }

type MockComputeService struct{}

func (mcs *MockComputeService) MatchPodSpec(pod *corev1.Pod) (*ultron.WeightedNode, error) {
	return &ultron.WeightedNode{
		Selector: map[string]string{
			"node-type": "mock-node",
		},
	}, nil
}

func (mcs *MockComputeService) CalculateWeightedNodeMedianPrice(node ultron.WeightedNode) (float64, error) {
	return 0, nil
}

func (mcs *MockComputeService) ComputeConfigurationMatchesWeightedNodeRequirements(configuration ultron.ComputeConfiguration, node ultron.WeightedNode) bool {
	return true
}

func (mcs *MockComputeService) ComputeConfigurationMatchesWeightedPodRequirements(configuration ultron.ComputeConfiguration, pod ultron.WeightedPod) bool {
	return true
}

func (mcs *MockComputeService) MatchWeightedNodeToComputeConfiguration(node ultron.WeightedNode) (*ultron.ComputeConfiguration, error) {
	return &ultron.ComputeConfiguration{}, nil
}

func (mcs *MockComputeService) MatchWeightedPodToComputeConfiguration(node ultron.WeightedPod) (*ultron.ComputeConfiguration, error) {
	return &ultron.ComputeConfiguration{}, nil
}

func (mcs *MockComputeService) MatchWeightedPodToWeightedNode(pod ultron.WeightedPod) (*ultron.WeightedNode, error) {
	return nil, nil
}

func (cs MockComputeService) ComputeInteruptionRateForWeightedNode(wNode ultron.WeightedNode) (*ultron.WeightedInteruptionRate, error) {
	return &ultron.WeightedInteruptionRate{}, nil
}

func (cs MockComputeService) ComputeLatencyRateForWeightedNode(wNode ultron.WeightedNode) (*ultron.WeightedLatencyRate, error) {
	return &ultron.WeightedLatencyRate{}, nil
}

type MockAlgorithm struct{}

func (ma *MockAlgorithm) StorageScore(wNode ultron.WeightedNode, wPod ultron.WeightedPod) float64 {
	return 0
}

func (ma *MockAlgorithm) NetworkScore(wNode ultron.WeightedNode, wPod ultron.WeightedPod) float64 {
	return 0
}

func (ma *MockAlgorithm) NodeScore(wNode ultron.WeightedNode) float64 {
	return 0
}

func (ma *MockAlgorithm) PriceScore(wNode ultron.WeightedNode) float64 {
	return 0.2
}

func (ma *MockAlgorithm) ResourceScore(wNode ultron.WeightedNode, wPod ultron.WeightedPod) float64 {
	return wNode.AvailableCPU - wPod.RequestedCPU
}

func (ma *MockAlgorithm) PodScore(wPod ultron.WeightedPod) float64 {
	return 0
}

func (ma *MockAlgorithm) TotalScore(wNode ultron.WeightedNode, wPod ultron.WeightedPod) float64 {
	return wNode.AvailableCPU - wPod.RequestedCPU
}

type MockCache struct{}

func (mc *MockCache) AddCacheItem(key string, value interface{}, time time.Duration) error {
	return nil
}

func (mc *MockCache) GetCacheItem(key string) (interface{}, error) {
	return nil, nil
}

func (mc *MockCache) GetAllComputeConfigurations() ([]ultron.ComputeConfiguration, error) {
	return []ultron.ComputeConfiguration{
		{
			ComputeType: ultron.ComputeTypeDurable,
			VmConfiguration: emma.VmConfiguration{
				VCpu:     intPtr(2),
				RamGb:    intPtr(8),
				VolumeGb: intPtr(50),
				Cost: &emma.VmConfigurationCost{
					PricePerUnit: float32Ptr(0.2),
				},
				VolumeType:        stringPtr("SSD"),
				CloudNetworkTypes: []string{"isolated", "public"},
			},
		},
	}, nil
}

func (mc *MockCache) GetEphemeralComputeConfigurations() ([]ultron.ComputeConfiguration, error) {
	return nil, nil
}

func (mc *MockCache) GetDurableComputeConfigurations() ([]ultron.ComputeConfiguration, error) {
	return nil, nil
}

func (mc *MockCache) GetWeightedNodes() ([]ultron.WeightedNode, error) {
	return []ultron.WeightedNode{
		{
			AvailableCPU:     4,
			TotalCPU:         4,
			AvailableMemory:  8,
			TotalMemory:      8,
			AvailableStorage: 50,
			DiskType:         "SSD",
			NetworkType:      "isolated",
		},
	}, nil
}

type MockMapper struct{}

func (mm *MockMapper) MapPodToWeightedPod(pod *corev1.Pod) (ultron.WeightedPod, error) {
	return ultron.WeightedPod{
		RequestedCPU:         2,
		RequestedMemory:      4,
		RequestedStorage:     50,
		RequestedDiskType:    "SSD",
		RequestedNetworkType: "isolated",
	}, nil
}

func (mm *MockMapper) MapNodeToWeightedNode(node *corev1.Node) (ultron.WeightedNode, error) {
	return ultron.WeightedNode{
		AvailableCPU:     2,
		TotalCPU:         4,
		AvailableMemory:  8,
		TotalMemory:      16,
		AvailableStorage: 100,
		TotalStorage:     200,
		DiskType:         "SSD",
		NetworkType:      "5G",
	}, nil
}

func TestMutatePods_Success(t *testing.T) {
	mockComputeService := &MockComputeService{}
	handler := handlers.NewIMutationHandler(mockComputeService)

	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
		},
		Spec: corev1.PodSpec{},
	}
	rawPod, _ := json.Marshal(pod)

	admissionReviewReq := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:  "1234",
			Kind: metav1.GroupVersionKind{Kind: "Pod"},
			Object: runtime.RawExtension{
				Raw: rawPod,
			},
		},
	}

	reqBody, _ := json.Marshal(admissionReviewReq)
	req := httptest.NewRequest(http.MethodPost, "/mutate", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	handler.MutatePodSpec(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code 200, got %d", resp.StatusCode)
	}

	var admissionReviewResp admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &admissionReviewResp); err != nil {
		t.Fatalf("Expected valid AdmissionReview response, but got error: %v", err)
	}

	if admissionReviewResp.Response.UID != admissionReviewReq.Request.UID {
		t.Errorf("Expected response UID to match request UID, got %s", admissionReviewResp.Response.UID)
	}

	if admissionReviewResp.Response.Allowed != true {
		t.Errorf("Expected Allowed to be true, but got false")
	}

	patch := admissionReviewResp.Response.Patch
	if patch == nil {
		t.Fatalf("Expected non-nil patch, but got nil")
	}

	expectedPatch := `[{"op":"add","path":"/spec/nodeSelector","value":{"node-type":"mock-node"}}]`
	if string(patch) != expectedPatch {
		t.Errorf("Expected patch %s, but got %s", expectedPatch, string(patch))
	}
}

func TestMutatePods_InvalidBody(t *testing.T) {
	mockComputeService := &MockComputeService{}
	handler := handlers.NewIMutationHandler(mockComputeService)

	req := httptest.NewRequest(http.MethodPost, "/mutate", bytes.NewBuffer([]byte("invalid body")))
	w := httptest.NewRecorder()

	handler.MutatePodSpec(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status code 400, got %d", resp.StatusCode)
	}
}

func TestMutationHandleAdmissionReview_NonPodKind(t *testing.T) {
	mockComputeService := &MockComputeService{}
	handler := handlers.NewIMutationHandler(mockComputeService)

	admissionRequest := &admissionv1.AdmissionRequest{
		Kind: metav1.GroupVersionKind{Kind: "Service"},
	}

	admissionResponse, err := handler.HandleAdmissionReview(admissionRequest)
	if err != nil {
		t.Fatalf("HandleAdmissionReview returned an error: %v", err)
	}

	if admissionResponse.Allowed != true {
		t.Errorf("Expected Allowed to be true for non-pod kind, got false")
	}
}

func TestMutationHandleAdmissionReview_PodSpecFailure(t *testing.T) {
	mockComputeService := &MockComputeService{}
	handler := handlers.NewIMutationHandler(mockComputeService)

	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
		},
		Spec: corev1.PodSpec{},
	}
	rawPod, _ := json.Marshal(pod)

	admissionRequest := &admissionv1.AdmissionRequest{
		UID:  "1234",
		Kind: metav1.GroupVersionKind{Kind: "Pod"},
		Object: runtime.RawExtension{
			Raw: rawPod,
		},
	}

	admissionResponse, err := handler.HandleAdmissionReview(admissionRequest)
	if err != nil {
		t.Fatalf("HandleAdmissionReview returned an error: %v", err)
	}

	if admissionResponse.Allowed != true {
		t.Errorf("Expected Allowed to be true, but got false")
	}
}
