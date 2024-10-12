package algorithm_test

import (
	"testing"

	ultron "github.com/be-heroes/ultron/pkg"
	algorithm "github.com/be-heroes/ultron/pkg/algorithm"
	"github.com/stretchr/testify/assert"
)

func TestResourceScore(t *testing.T) {
	// Arrange
	alg := algorithm.NewAlgorithm()

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
	score := alg.ResourceScore(&node, &pod)
	expected := (4.0-2.0)/8.0 + (8.0-4.0)/16.0

	// Assert
	assert.Equal(t, expected, score, "ResourceScore was incorrect")
}

func TestStorageScore(t *testing.T) {
	// Arrange
	alg := algorithm.NewAlgorithm()

	node := ultron.WeightedNode{
		DiskType: "SSD",
	}

	pod := ultron.WeightedPod{
		RequestedDiskType: "SSD",
	}

	// Act
	score1 := alg.StorageScore(&node, &pod)
	expected1 := 1.0

	pod.RequestedDiskType = "HDD"
	score2 := alg.StorageScore(&node, &pod)
	expected2 := 0.0

	// Assert
	assert.Equal(t, expected1, score1, "StorageScore was incorrect")
	assert.Equal(t, expected2, score2, "StorageScore was incorrect for mismatch")
}

func TestNetworkScore(t *testing.T) {
	// Arrange
	alg := algorithm.NewAlgorithm()

	node := ultron.WeightedNode{
		NetworkType: "10G",
	}

	pod := ultron.WeightedPod{
		RequestedNetworkType: "10G",
	}

	// Act
	score1 := alg.NetworkScore(&node, &pod)
	expected1 := 1.0

	pod.RequestedNetworkType = "1G"
	score2 := alg.NetworkScore(&node, &pod)
	expected2 := 0.0

	// Assert
	assert.Equal(t, expected1, score1, "NetworkScore was incorrect")
	assert.Equal(t, expected2, score2, "NetworkScore was incorrect for mismatch")
}

func TestPriceScore(t *testing.T) {
	// Arrange
	alg := algorithm.NewAlgorithm()

	node := ultron.WeightedNode{
		Price:       10.0,
		MedianPrice: 8.0,
	}

	// Act
	score1 := alg.PriceScore(&node)
	expected1 := 1.0 - (node.MedianPrice / node.Price)

	node.Price = 0
	score2 := alg.PriceScore(&node)
	expected2 := 0.0

	// Assert
	assert.Equal(t, expected1, score1, "PriceScore was incorrect")
	assert.Equal(t, expected2, score2, "PriceScore was incorrect for zero price")
}

func TestNodeScore(t *testing.T) {
	// Arrange
	alg := algorithm.NewAlgorithm()

	node := ultron.WeightedNode{
		Price:            10.0,
		MedianPrice:      8.0,
		InterruptionRate: ultron.WeightedInteruptionRate{Value: 0.1},
	}

	// Act
	score1 := alg.NodeScore(&node)
	expected1 := node.Price / (node.MedianPrice + node.InterruptionRate.Value)

	node.Price = 0
	score2 := alg.NodeScore(&node)
	expected2 := 0.0

	// Assert
	assert.Equal(t, expected1, score1, "NodeScore was incorrect")
	assert.Equal(t, expected2, score2, "NodeScore was incorrect for zero price")
}

func TestPodScore(t *testing.T) {
	// Arrange
	alg := algorithm.NewAlgorithm()

	pod := ultron.WeightedPod{
		Priority: ultron.PriorityHigh,
	}

	// Act
	score1 := alg.PodScore(&pod)
	expected1 := 1.0

	pod.Priority = ultron.PriorityLow
	score2 := alg.PodScore(&pod)
	expected2 := 0.0

	// Assert
	assert.Equal(t, expected1, score1, "PodScore was incorrect")
	assert.Equal(t, expected2, score2, "PodScore was incorrect for low priority")
}

func TestTotalScore(t *testing.T) {
	// Arrange
	alg := algorithm.NewAlgorithm()

	node := ultron.WeightedNode{
		TotalCPU:         8,
		AvailableCPU:     4,
		TotalMemory:      16,
		AvailableMemory:  8,
		DiskType:         "SSD",
		NetworkType:      "10G",
		Price:            10.0,
		MedianPrice:      8.0,
		InterruptionRate: ultron.WeightedInteruptionRate{Value: 0.1},
	}

	pod := ultron.WeightedPod{
		RequestedCPU:         2,
		RequestedMemory:      4,
		RequestedDiskType:    "SSD",
		RequestedNetworkType: "10G",
		Priority:             ultron.PriorityHigh,
	}

	// Act
	score := alg.TotalScore(&node, &pod)

	resourceScore := algorithm.Alpha * alg.ResourceScore(&node, &pod)
	storageScore := algorithm.Beta * alg.StorageScore(&node, &pod)
	networkScore := algorithm.Gamma * alg.NetworkScore(&node, &pod)
	priceScore := algorithm.Delta * alg.PriceScore(&node)
	nodeScore := algorithm.Epsilon * alg.NodeScore(&node)
	podScore := algorithm.Zeta * alg.PodScore(&pod)

	expected := resourceScore + storageScore + networkScore + priceScore - nodeScore + podScore

	// Assert
	assert.Equal(t, expected, score, "TotalScore was incorrect")
}
