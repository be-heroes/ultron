package ultron

// import (
// 	"errors"
// 	"math"
// 	"testing"

// 	emma "github.com/emma-community/emma-go-sdk"
// 	corev1 "k8s.io/api/core/v1"
// )

// // Helper function to return pointers
// func int32Ptr(val int32) *int32 {
// 	return &val
// }

// func float32Ptr(val float32) *float32 {
// 	return &val
// }

// func stringPtr(val string) *string {
// 	return &val
// }

// func mockGetAllComputeConfigurationsFromCache() ([]ComputeConfiguration, error) {
// 	return []ComputeConfiguration{
// 		{
// 			VmConfiguration: emma.VmConfiguration{
// 				VCpu:       int32Ptr(4),
// 				RamGb:      int32Ptr(16),
// 				VolumeGb:   int32Ptr(100),
// 				VolumeType: stringPtr("SSD"),
// 				Cost: &emma.VmConfigurationCost{
// 					PricePerUnit: float32Ptr(0.2),
// 				},
// 				CloudNetworkTypes: []string{"isolated"}},
// 		},
// 	}, nil
// }

// func mockGetWeightedNodesFromCache() ([]WeightedNode, error) {
// 	return []WeightedNode{
// 		{
// 			AvailableCPU:     4,
// 			TotalCPU:         4,
// 			AvailableMemory:  16,
// 			TotalMemory:      16,
// 			AvailableStorage: 100,
// 			DiskType:         "SSD",
// 			NetworkType:      "isolated",
// 			Price:            0.2,
// 		},
// 	}, nil
// }

// func mockMapPodToWeightedPod(pod *corev1.Pod) (WeightedPod, error) {
// 	return WeightedPod{
// 		RequestedCPU:         2,
// 		RequestedMemory:      4,
// 		RequestedStorage:     10,
// 		RequestedDiskType:    "SSD",
// 		RequestedNetworkType: "isolated",
// 	}, nil
// }

// func TestComputePodSpec(t *testing.T) {
// 	pod := &corev1.Pod{
// 		Spec: corev1.PodSpec{},
// 	}

// 	actualMapPodToWeightedPod := MapPodToWeightedPod
// 	MapPodToWeightedPod = mockMapPodToWeightedPod
// 	actualGetWeightedNodesFromCache := GetWeightedNodesFromCache
// 	GetWeightedNodesFromCache = mockGetWeightedNodesFromCache
// 	actualGetAllComputeConfigurationsFromCache := GetAllComputeConfigurationsFromCache
// 	GetAllComputeConfigurationsFromCache = mockGetAllComputeConfigurationsFromCache

// 	weightedNode, err := ComputePodSpec(pod)

// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	if weightedNode == nil {
// 		t.Fatal("Expected a non-nil weighted node, but got nil")
// 	}

// 	if weightedNode.AvailableCPU != 4 {
// 		t.Errorf("Expected AvailableCPU to be 4, but got %f", weightedNode.AvailableCPU)
// 	}

// 	MapPodToWeightedPod = actualMapPodToWeightedPod
// 	GetWeightedNodesFromCache = actualGetWeightedNodesFromCache
// 	GetAllComputeConfigurationsFromCache = actualGetAllComputeConfigurationsFromCache
// }

// func TestMatchWeightedPodToWeightedNode(t *testing.T) {
// 	actualGetWeightedNodesFromCache := GetWeightedNodesFromCache
// 	GetWeightedNodesFromCache = mockGetWeightedNodesFromCache

// 	mockPod := WeightedPod{
// 		RequestedCPU:         4,
// 		RequestedMemory:      16,
// 		RequestedStorage:     50,
// 		RequestedDiskType:    "SSD",
// 		RequestedNetworkType: "isolated",
// 	}

// 	Score = func(node WeightedNode, pod WeightedPod) float64 {
// 		if node.AvailableCPU == 8 {
// 			return 10.0
// 		}
// 		return 5.0
// 	}

// 	matchedNode, err := MatchWeightedPodToWeightedNode(mockPod)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	if matchedNode.AvailableCPU != 4 {
// 		t.Errorf("Expected AvailableCPU to be 4, but got %f", matchedNode.AvailableCPU)
// 	}

// 	GetWeightedNodesFromCache = actualGetWeightedNodesFromCache
// }

// func TestMatchWeightedPodToWeightedNode_NoMatch(t *testing.T) {
// 	actualGetWeightedNodesFromCache := GetWeightedNodesFromCache
// 	GetWeightedNodesFromCache = func() ([]WeightedNode, error) {
// 		return []WeightedNode{
// 			{
// 				AvailableCPU:    2, // Less than required CPU
// 				AvailableMemory: 8, // Less than required memory
// 			},
// 		}, nil
// 	}

// 	mockPod := WeightedPod{
// 		RequestedCPU:    4,
// 		RequestedMemory: 16,
// 	}

// 	matchedNode, err := MatchWeightedPodToWeightedNode(mockPod)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	if matchedNode.AvailableCPU != 0 {
// 		t.Errorf("Expected no matching node, but found one with AvailableCPU: %f", matchedNode.AvailableCPU)
// 	}

// 	GetWeightedNodesFromCache = actualGetWeightedNodesFromCache
// }

// func TestMatchWeightedPodToWeightedNode_Error(t *testing.T) {
// 	actualGetWeightedNodesFromCache := GetWeightedNodesFromCache
// 	GetWeightedNodesFromCache = func() ([]WeightedNode, error) {
// 		return nil, errors.New("cache error")
// 	}

// 	mockPod := WeightedPod{
// 		RequestedCPU:    4,
// 		RequestedMemory: 16,
// 	}

// 	_, err := MatchWeightedPodToWeightedNode(mockPod)
// 	if err == nil {
// 		t.Fatalf("Expected an error, got none")
// 	}

// 	if err.Error() != "cache error" {
// 		t.Errorf("Expected error message 'cache error', got %v", err)
// 	}

// 	GetWeightedNodesFromCache = actualGetWeightedNodesFromCache
// }

// func TestMatchWeightedPodToComputeConfiguration(t *testing.T) {
// 	wPod := WeightedPod{
// 		RequestedCPU:         2,
// 		RequestedMemory:      4,
// 		RequestedStorage:     10,
// 		RequestedDiskType:    "SSD",
// 		RequestedNetworkType: "isolated",
// 	}

// 	actualGetAllComputeConfigurationsFromCache := GetAllComputeConfigurationsFromCache
// 	GetAllComputeConfigurationsFromCache = mockGetAllComputeConfigurationsFromCache

// 	computeConfig, err := MatchWeightedPodToComputeConfiguration(wPod)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	if computeConfig == nil {
// 		t.Fatal("Expected a non-nil ComputeConfiguration, but got nil")
// 	}

// 	if *computeConfig.VCpu != 4 {
// 		t.Errorf("Expected VCpu to be 4, but got %d", *computeConfig.VCpu)
// 	}

// 	GetAllComputeConfigurationsFromCache = actualGetAllComputeConfigurationsFromCache
// }

// func TestCalculateWeightedNodeMedianPrice(t *testing.T) {
// 	wNode := WeightedNode{
// 		AvailableCPU:     4,
// 		TotalCPU:         4,
// 		AvailableMemory:  16,
// 		TotalMemory:      16,
// 		AvailableStorage: 100,
// 		DiskType:         "SSD",
// 		NetworkType:      "isolated",
// 	}

// 	actualGetAllComputeConfigurationsFromCache := GetAllComputeConfigurationsFromCache
// 	GetAllComputeConfigurationsFromCache = mockGetAllComputeConfigurationsFromCache
// 	actualGetWeightedNodesFromCache := GetWeightedNodesFromCache
// 	GetWeightedNodesFromCache = mockGetWeightedNodesFromCache

// 	medianPrice, err := CalculateWeightedNodeMedianPrice(wNode)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}

// 	expectedPrice := 0.2
// 	if math.Abs(medianPrice-expectedPrice) > 0.001 {
// 		t.Errorf("Expected median price to be %f, but got %f", expectedPrice, medianPrice)
// 	}

// 	GetAllComputeConfigurationsFromCache = actualGetAllComputeConfigurationsFromCache
// 	GetWeightedNodesFromCache = actualGetWeightedNodesFromCache
// }
