package algorithm_test

import (
	"testing"

	ultron "ultron/internal"
	algorithm "ultron/internal/algorithm"
)

func TestResourceFitScore(t *testing.T) {
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
	score := alg.ResourceFitScore(node, pod)
	expected := (4.0-2.0)/8.0 + (8.0-4.0)/16.0

	// Assert
	if score != expected {
		t.Errorf("ResourceFitScore was incorrect, got: %f, want: %f", score, expected)
	}
}

func TestDiskTypeScore(t *testing.T) {
	// Arrange
	alg := algorithm.NewIAlgorithm()

	node := ultron.WeightedNode{
		DiskType: "SSD",
	}

	pod := ultron.WeightedPod{
		RequestedDiskType: "SSD",
	}

	// Act
	score1 := alg.DiskTypeScore(node, pod)
	expected1 := 1.0

	pod.RequestedDiskType = "HDD"
	score2 := alg.DiskTypeScore(node, pod)
	expected2 := 0.0

	// Assert
	if score1 != expected1 {
		t.Errorf("DiskTypeScore was incorrect, got: %f, want: %f", score1, expected1)
	}

	if score2 != expected2 {
		t.Errorf("DiskTypeScore was incorrect for mismatch, got: %f, want: %f", score2, expected2)
	}
}

func TestNetworkTypeScore(t *testing.T) {
	// Arrange
	alg := algorithm.NewIAlgorithm()

	node := ultron.WeightedNode{
		NetworkType: "10G",
	}

	pod := ultron.WeightedPod{
		RequestedNetworkType: "10G",
	}

	// Act
	score1 := alg.NetworkTypeScore(node, pod)
	expected1 := 1.0

	pod.RequestedNetworkType = "1G"
	score2 := alg.NetworkTypeScore(node, pod)
	expected2 := 0.0

	// Assert
	if score1 != expected1 {
		t.Errorf("NetworkTypeScore was incorrect, got: %f, want: %f", score1, expected1)
	}

	if score2 != expected2 {
		t.Errorf("NetworkTypeScore was incorrect for mismatch, got: %f, want: %f", score2, expected2)
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

func TestNodeStabilityScore(t *testing.T) {
	// Arrange
	alg := algorithm.NewIAlgorithm()

	node := ultron.WeightedNode{
		Price:            10.0,
		MedianPrice:      8.0,
		InterruptionRate: 0.1,
	}

	// Act
	score1 := alg.NodeStabilityScore(node)
	expected1 := node.InterruptionRate * (node.Price / node.MedianPrice)

	node.Price = 0
	score2 := alg.NodeStabilityScore(node)
	expected2 := 0.0

	// Assert
	if score1 != expected1 {
		t.Errorf("NodeStabilityScore was incorrect, got: %f, want: %f", score1, expected1)
	}

	if score2 != expected2 {
		t.Errorf("NodeStabilityScore was incorrect for zero price, got: %f, want: %f", score2, expected2)
	}
}

func TestWorkloadPriorityScore(t *testing.T) {
	// Arrange
	alg := algorithm.NewIAlgorithm()

	pod := ultron.WeightedPod{
		Priority: ultron.PriorityHigh,
	}

	// Act
	score1 := alg.WorkloadPriorityScore(pod)
	expected1 := 1.0

	pod.Priority = ultron.PriorityLow
	score2 := alg.WorkloadPriorityScore(pod)
	expected2 := 0.0

	// Assert
	if score1 != expected1 {
		t.Errorf("WorkloadPriorityScore was incorrect, got: %f, want: %f", score1, expected1)
	}

	if score2 != expected2 {
		t.Errorf("WorkloadPriorityScore was incorrect for low priority, got: %f, want: %f", score2, expected2)
	}
}

func TestScore(t *testing.T) {
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
	score := alg.Score(node, pod)

	resourceFit := algorithm.Alpha * alg.ResourceFitScore(node, pod)
	diskType := algorithm.Beta * alg.DiskTypeScore(node, pod)
	networkType := algorithm.Gamma * alg.NetworkTypeScore(node, pod)
	price := algorithm.Delta * alg.PriceScore(node)
	stability := algorithm.Epsilon * alg.NodeStabilityScore(node)
	priority := algorithm.Zeta * alg.WorkloadPriorityScore(pod)

	expected := resourceFit + diskType + networkType + price - stability + priority

	// Assert
	if score != expected {
		t.Errorf("Score was incorrect, got: %f, want: %f", score, expected)
	}
}
