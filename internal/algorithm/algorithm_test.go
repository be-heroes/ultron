package algorithm_test

import (
	"testing"

	ultron "ultron/internal"
	algorithm "ultron/internal/algorithm"
)

func TestResourceScore(t *testing.T) {
	// Arrange
	alg := algorithm.NewIAlgorithm()

	node := ultron.WeightedNode{
		TotalCPU:        8,
		AvailableCPU:    4,
		TotalMemory:     16,
		AvailableMemory: 8,
	}

	pod := ultron.WeightedPod{
		RequestedCPU:    2,
		RequestedMemory: 4,
	}

	// Act
	score := alg.ResourceScore(node, pod)
	expected := (4.0-2.0)/8.0 + (8.0-4.0)/16.0

	// Assert
	if score != expected {
		t.Errorf("ResourceScore was incorrect, got: %f, want: %f", score, expected)
	}
}

func TestStorageScore(t *testing.T) {
	// Arrange
	alg := algorithm.NewIAlgorithm()

	node := ultron.WeightedNode{
		DiskType: "SSD",
	}

	pod := ultron.WeightedPod{
		RequestedDiskType: "SSD",
	}

	// Act
	score1 := alg.StorageScore(node, pod)
	expected1 := 1.0

	pod.RequestedDiskType = "HDD"
	score2 := alg.StorageScore(node, pod)
	expected2 := 0.0

	// Assert
	if score1 != expected1 {
		t.Errorf("StorageScore was incorrect, got: %f, want: %f", score1, expected1)
	}

	if score2 != expected2 {
		t.Errorf("StorageScore was incorrect for mismatch, got: %f, want: %f", score2, expected2)
	}
}

func TestNetworkScore(t *testing.T) {
	// Arrange
	alg := algorithm.NewIAlgorithm()

	node := ultron.WeightedNode{
		NetworkType: "10G",
	}

	pod := ultron.WeightedPod{
		RequestedNetworkType: "10G",
	}

	// Act
	score1 := alg.NetworkScore(node, pod)
	expected1 := 1.0

	pod.RequestedNetworkType = "1G"
	score2 := alg.NetworkScore(node, pod)
	expected2 := 0.0

	// Assert
	if score1 != expected1 {
		t.Errorf("NetworkScore was incorrect, got: %f, want: %f", score1, expected1)
	}

	if score2 != expected2 {
		t.Errorf("NetworkScore was incorrect for mismatch, got: %f, want: %f", score2, expected2)
	}
}

func TestPriceScore(t *testing.T) {
	// Arrange
	alg := algorithm.NewIAlgorithm()

	node := ultron.WeightedNode{
		Price:       10.0,
		MedianPrice: 8.0,
	}

	// Act
	score1 := alg.PriceScore(node)
	expected1 := 1.0 - (node.MedianPrice / node.Price)

	node.Price = 0
	score2 := alg.PriceScore(node)
	expected2 := 0.0

	// Assert
	if score1 != expected1 {
		t.Errorf("PriceScore was incorrect, got: %f, want: %f", score1, expected1)
	}

	if score2 != expected2 {
		t.Errorf("PriceScore was incorrect for zero price, got: %f, want: %f", score2, expected2)
	}
}

func TestNodeScore(t *testing.T) {
	// Arrange
	alg := algorithm.NewIAlgorithm()

	node := ultron.WeightedNode{
		Price:            10.0,
		MedianPrice:      8.0,
		InterruptionRate: 0.1,
	}

	// Act
	score1 := alg.NodeScore(node)
	expected1 := node.InterruptionRate * (node.Price / node.MedianPrice)

	node.Price = 0
	score2 := alg.NodeScore(node)
	expected2 := 0.0

	// Assert
	if score1 != expected1 {
		t.Errorf("NodeScore was incorrect, got: %f, want: %f", score1, expected1)
	}

	if score2 != expected2 {
		t.Errorf("NodeScore was incorrect for zero price, got: %f, want: %f", score2, expected2)
	}
}

func TestPodScore(t *testing.T) {
	// Arrange
	alg := algorithm.NewIAlgorithm()

	pod := ultron.WeightedPod{
		Priority: ultron.PriorityHigh,
	}

	// Act
	score1 := alg.PodScore(pod)
	expected1 := 1.0

	pod.Priority = ultron.PriorityLow
	score2 := alg.PodScore(pod)
	expected2 := 0.0

	// Assert
	if score1 != expected1 {
		t.Errorf("PodScore was incorrect, got: %f, want: %f", score1, expected1)
	}

	if score2 != expected2 {
		t.Errorf("PodScore was incorrect for low priority, got: %f, want: %f", score2, expected2)
	}
}

func TestTotalScore(t *testing.T) {
	// Arrange
	alg := algorithm.NewIAlgorithm()

	node := ultron.WeightedNode{
		TotalCPU:         8,
		AvailableCPU:     4,
		TotalMemory:      16,
		AvailableMemory:  8,
		DiskType:         "SSD",
		NetworkType:      "10G",
		Price:            10.0,
		MedianPrice:      8.0,
		InterruptionRate: 0.1,
	}

	pod := ultron.WeightedPod{
		RequestedCPU:         2,
		RequestedMemory:      4,
		RequestedDiskType:    "SSD",
		RequestedNetworkType: "10G",
		Priority:             ultron.PriorityHigh,
	}

	// Act
	score := alg.TotalScore(node, pod)

	resourceScore := algorithm.Alpha * alg.ResourceScore(node, pod)
	storageScore := algorithm.Beta * alg.StorageScore(node, pod)
	networkScore := algorithm.Gamma * alg.NetworkScore(node, pod)
	priceScore := algorithm.Delta * alg.PriceScore(node)
	nodeScore := algorithm.Epsilon * alg.NodeScore(node)
	podScore := algorithm.Zeta * alg.PodScore(pod)

	expected := resourceScore + storageScore + networkScore + priceScore - nodeScore + podScore

	// Assert
	if score != expected {
		t.Errorf("TotalScore was incorrect, got: %f, want: %f", score, expected)
	}
}
