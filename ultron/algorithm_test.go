package ultron_test

import (
	"testing"

	ultron "emma.ms/ultron/ultron"
)

func TestResourceFitScore(t *testing.T) {
	alg := ultron.NewIAlgorithm()

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

	score := alg.ResourceFitScore(node, pod)
	expected := (4.0-2.0)/8.0 + (8.0-4.0)/16.0

	if score != expected {
		t.Errorf("ResourceFitScore was incorrect, got: %f, want: %f", score, expected)
	}
}

func TestDiskTypeScore(t *testing.T) {
	alg := ultron.NewIAlgorithm()

	node := ultron.WeightedNode{
		DiskType: "SSD",
	}

	pod := ultron.WeightedPod{
		RequestedDiskType: "SSD",
	}

	score := alg.DiskTypeScore(node, pod)
	expected := 1.0

	if score != expected {
		t.Errorf("DiskTypeScore was incorrect, got: %f, want: %f", score, expected)
	}

	// Test for mismatch
	pod.RequestedDiskType = "HDD"
	score = alg.DiskTypeScore(node, pod)
	expected = 0.0

	if score != expected {
		t.Errorf("DiskTypeScore was incorrect for mismatch, got: %f, want: %f", score, expected)
	}
}

func TestNetworkTypeScore(t *testing.T) {
	alg := ultron.NewIAlgorithm()

	node := ultron.WeightedNode{
		NetworkType: "10G",
	}

	pod := ultron.WeightedPod{
		RequestedNetworkType: "10G",
	}

	score := alg.NetworkTypeScore(node, pod)
	expected := 1.0

	if score != expected {
		t.Errorf("NetworkTypeScore was incorrect, got: %f, want: %f", score, expected)
	}

	// Test for mismatch
	pod.RequestedNetworkType = "1G"
	score = alg.NetworkTypeScore(node, pod)
	expected = 0.0

	if score != expected {
		t.Errorf("NetworkTypeScore was incorrect for mismatch, got: %f, want: %f", score, expected)
	}
}

func TestPriceScore(t *testing.T) {
	alg := ultron.NewIAlgorithm()

	node := ultron.WeightedNode{
		Price:       10.0,
		MedianPrice: 8.0,
	}

	score := alg.PriceScore(node)
	expected := 1.0 - (node.MedianPrice / node.Price)

	if score != expected {
		t.Errorf("PriceScore was incorrect, got: %f, want: %f", score, expected)
	}

	// Test for zero price
	node.Price = 0
	score = alg.PriceScore(node)
	expected = 0.0

	if score != expected {
		t.Errorf("PriceScore was incorrect for zero price, got: %f, want: %f", score, expected)
	}
}

func TestNodeStabilityScore(t *testing.T) {
	alg := ultron.NewIAlgorithm()

	node := ultron.WeightedNode{
		Price:            10.0,
		MedianPrice:      8.0,
		InterruptionRate: 0.1,
	}

	score := alg.NodeStabilityScore(node)
	expected := node.InterruptionRate * (node.Price / node.MedianPrice)

	if score != expected {
		t.Errorf("NodeStabilityScore was incorrect, got: %f, want: %f", score, expected)
	}

	// Test for zero price or median price
	node.Price = 0
	score = alg.NodeStabilityScore(node)
	expected = 0.0

	if score != expected {
		t.Errorf("NodeStabilityScore was incorrect for zero price, got: %f, want: %f", score, expected)
	}
}

func TestWorkloadPriorityScore(t *testing.T) {
	alg := ultron.NewIAlgorithm()

	pod := ultron.WeightedPod{
		Priority: ultron.PriorityHigh,
	}

	score := alg.WorkloadPriorityScore(pod)
	expected := 1.0

	if score != expected {
		t.Errorf("WorkloadPriorityScore was incorrect, got: %f, want: %f", score, expected)
	}

	// Test for low priority
	pod.Priority = ultron.PriorityLow
	score = alg.WorkloadPriorityScore(pod)
	expected = 0.0

	if score != expected {
		t.Errorf("WorkloadPriorityScore was incorrect for low priority, got: %f, want: %f", score, expected)
	}
}

func TestScore(t *testing.T) {
	alg := ultron.NewIAlgorithm()

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

	score := alg.Score(node, pod)

	resourceFit := ultron.Alpha * alg.ResourceFitScore(node, pod)
	diskType := ultron.Beta * alg.DiskTypeScore(node, pod)
	networkType := ultron.Gamma * alg.NetworkTypeScore(node, pod)
	price := ultron.Delta * alg.PriceScore(node)
	stability := ultron.Epsilon * alg.NodeStabilityScore(node)
	priority := ultron.Zeta * alg.WorkloadPriorityScore(pod)

	expected := resourceFit + diskType + networkType + price - stability + priority

	if score != expected {
		t.Errorf("Score was incorrect, got: %f, want: %f", score, expected)
	}
}
