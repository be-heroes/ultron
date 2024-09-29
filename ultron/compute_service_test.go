package ultron_test

import (
	"math"
	"testing"

	ultron "emma.ms/ultron/ultron"
	corev1 "k8s.io/api/core/v1"
)

func TestComputePodSpec_Success(t *testing.T) {
	// Arrange
	mockAlgorithm := &MockAlgorithm{}
	mockCache := &MockCache{}
	mockMapper := &MockMapper{}

	service := ultron.NewIComputeService(mockAlgorithm, mockCache, mockMapper)

	pod := &corev1.Pod{}

	// Act
	wNode, err := service.ComputePodSpec(pod)

	// Assert
	if err != nil {
		t.Fatalf("ComputePodSpec returned an error: %v", err)
	}

	if wNode == nil {
		t.Fatal("Expected wNode to not be nil")
	}

	if wNode.AvailableCPU != 4 {
		t.Errorf("Expected AvailableCPU to be 4, got %f", wNode.AvailableCPU)
	}

	if wNode.DiskType != "SSD" {
		t.Errorf("Expected DiskType to be SSD, got %s", wNode.DiskType)
	}

	if wNode.NetworkType != "isolated" {
		t.Errorf("Expected NetworkType to be isolated, got %s", wNode.NetworkType)
	}
}

func TestComputePodSpec_NoWeightedNode(t *testing.T) {
	// Arrange
	mockCache := &MockCache{}
	mockAlgorithm := &MockAlgorithm{}
	mockMapper := &MockMapper{}

	service := ultron.NewIComputeService(mockAlgorithm, mockCache, mockMapper)

	pod := &corev1.Pod{}

	// Act
	wNode, err := service.ComputePodSpec(pod)

	// Assert
	if err != nil {
		t.Fatalf("ComputePodSpec returned an error: %v", err)
	}

	if wNode == nil {
		t.Fatal("Expected wNode to not be nil")
	}
}

func TestMatchWeightedPodToComputeConfiguration_Success(t *testing.T) {
	// Arrange
	mockAlgorithm := &MockAlgorithm{}
	mockCache := &MockCache{}
	mockMapper := &MockMapper{}

	service := ultron.NewIComputeService(mockAlgorithm, mockCache, mockMapper)

	wPod := ultron.WeightedPod{
		RequestedCPU:         1,
		RequestedMemory:      2,
		RequestedStorage:     30,
		RequestedDiskType:    "SSD",
		RequestedNetworkType: "isolated",
	}

	// Act
	computeConfig, err := service.MatchWeightedPodToComputeConfiguration(wPod)

	// Assert
	if err != nil {
		t.Fatalf("MatchWeightedPodToComputeConfiguration returned an error: %v", err)
	}

	if computeConfig == nil {
		t.Fatal("Expected computeConfig to not be nil")
	}

	if *computeConfig.VCpu != 2 {
		t.Errorf("Expected VCpu to be 2, got %d", *computeConfig.VCpu)
	}

	if computeConfig.ComputeType != ultron.ComputeTypeDurable {
		t.Errorf("Expected ComputeType to be durable, got %s", computeConfig.ComputeType)
	}
}

func TestCalculateWeightedNodeMedianPrice_Success(t *testing.T) {
	// Arrange
	mockAlgorithm := &MockAlgorithm{}
	mockCache := &MockCache{}
	mockMapper := &MockMapper{}

	service := ultron.NewIComputeService(mockAlgorithm, mockCache, mockMapper)

	wNode := ultron.WeightedNode{
		AvailableCPU:     4,
		AvailableMemory:  8,
		AvailableStorage: 50,
		DiskType:         "SSD",
		NetworkType:      "isolated",
	}

	// Act
	medianPrice, err := service.CalculateWeightedNodeMedianPrice(wNode)

	// Assert
	if err != nil {
		t.Fatalf("CalculateWeightedNodeMedianPrice returned an error: %v", err)
	}

	expectedPrice := 0.2
	if math.Abs(medianPrice-expectedPrice) > 0.01 {
		t.Errorf("Expected median price to be %f, got %f", expectedPrice, medianPrice)
	}
}

func TestMatchWeightedPodToWeightedNode_Success(t *testing.T) {
	// Arrange
	mockAlgorithm := &MockAlgorithm{}
	mockCache := &MockCache{}
	mockMapper := &MockMapper{}

	service := ultron.NewIComputeService(mockAlgorithm, mockCache, mockMapper)

	wPod := ultron.WeightedPod{
		RequestedCPU:    2,
		RequestedMemory: 4,
	}

	// Act
	wNode, err := service.MatchWeightedPodToWeightedNode(wPod)

	// Assert
	if err != nil {
		t.Fatalf("MatchWeightedPodToWeightedNode returned an error: %v", err)
	}

	if wNode == nil {
		t.Fatal("Expected wNode to not be nil")
	}

	if wNode.AvailableCPU != 4 {
		t.Errorf("Expected AvailableCPU to be 4, got %f", wNode.AvailableCPU)
	}
}

func TestMatchWeightedPodToComputeConfiguration_NoMatch(t *testing.T) {
	// Arrange
	mockCache := &MockCache{}
	mockAlgorithm := &MockAlgorithm{}
	mockMapper := &MockMapper{}

	service := ultron.NewIComputeService(mockAlgorithm, mockCache, mockMapper)

	wPod := ultron.WeightedPod{
		RequestedCPU: 4,
	}

	// Act
	computeConfig, err := service.MatchWeightedPodToComputeConfiguration(wPod)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	if computeConfig != nil {
		t.Fatalf("Expected no matching configuration, but got: %v", computeConfig)
	}
}
