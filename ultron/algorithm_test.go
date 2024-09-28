package ultron

import (
	"testing"
)

func TestResourceFitScore(t *testing.T) {
	node := WeightedNode{
		AvailableCPU:    8,
		TotalCPU:        16,
		AvailableMemory: 32,
		TotalMemory:     64,
	}

	pod := WeightedPod{
		RequestedCPU:    4,
		RequestedMemory: 16,
	}

	expected := (8.0-4.0)/16.0 + (32.0-16.0)/64.0
	result := ResourceFitScore(node, pod)
	if result != expected {
		t.Errorf("Expected %f, but got %f", expected, result)
	}
}

func TestDiskTypeScore(t *testing.T) {
	node := WeightedNode{
		DiskType: "ssd",
	}
	pod := WeightedPod{
		RequestedDiskType: "ssd",
	}

	result := DiskTypeScore(node, pod)
	if result != 1.0 {
		t.Errorf("Expected 1.0, but got %f", result)
	}

	pod.RequestedDiskType = "hdd"
	result = DiskTypeScore(node, pod)
	if result != 0.0 {
		t.Errorf("Expected 0.0, but got %f", result)
	}
}

func TestNetworkTypeScore(t *testing.T) {
	node := WeightedNode{
		NetworkType: "5G",
	}
	pod := WeightedPod{
		RequestedNetworkType: "5G",
	}

	result := NetworkTypeScore(node, pod)
	if result != 1.0 {
		t.Errorf("Expected 1.0, but got %f", result)
	}

	pod.RequestedNetworkType = "4G"
	result = NetworkTypeScore(node, pod)
	if result != 0.0 {
		t.Errorf("Expected 0.0, but got %f", result)
	}
}

func TestPriceScore(t *testing.T) {
	node := WeightedNode{
		Price:       1.0,
		MedianPrice: 2.0,
	}

	expected := 1.0 - (node.MedianPrice / node.Price)
	result := PriceScore(node)
	if result != expected {
		t.Errorf("Expected %f, but got %f", expected, result)
	}
}

func TestNodeStabilityScore(t *testing.T) {
	node := WeightedNode{
		InterruptionRate: 0.2,
		Price:            1.0,
		MedianPrice:      2.0,
	}

	expected := node.InterruptionRate * (node.Price / node.MedianPrice)
	result := NodeStabilityScore(node)
	if result != expected {
		t.Errorf("Expected %f, but got %f", expected, result)
	}
}

func TestWorkloadPriorityScore(t *testing.T) {
	pod := WeightedPod{
		Priority: PriorityHigh,
	}

	result := WorkloadPriorityScore(pod)
	if result != 1.0 {
		t.Errorf("Expected 1.0, but got %f", result)
	}

	pod.Priority = PriorityLow
	result = WorkloadPriorityScore(pod)
	if result != 0.0 {
		t.Errorf("Expected 0.0, but got %f", result)
	}
}

func TestScore(t *testing.T) {
	node := WeightedNode{
		AvailableCPU:     8,
		TotalCPU:         16,
		AvailableMemory:  32,
		TotalMemory:      64,
		DiskType:         "ssd",
		NetworkType:      "5G",
		MedianPrice:      2.0,
		Price:            1.0,
		InterruptionRate: 0.1,
	}

	pod := WeightedPod{
		RequestedCPU:         4,
		RequestedMemory:      16,
		RequestedDiskType:    "ssd",
		RequestedNetworkType: "5G",
		Priority:             PriorityHigh,
	}

	resourceFit := Alpha * ((8.0-4.0)/16.0 + (32.0-16.0)/64.0)
	diskType := Beta * 1.0
	networkType := Gamma * 1.0
	price := Delta * (1.0 - (2.0 / 1.0))
	stability := Epsilon * (0.1 * (1.0 / 2.0))
	priority := Zeta * 1.0

	expected := resourceFit + diskType + networkType + price - stability + priority
	result := Score(node, pod)
	if result != expected {
		t.Errorf("Expected %f, but got %f", expected, result)
	}
}
