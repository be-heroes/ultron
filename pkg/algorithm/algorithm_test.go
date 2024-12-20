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
		Weights: map[string]float64{
			ultron.WeightKeyCpuTotal:        8,
			ultron.WeightKeyCpuAvailable:    4,
			ultron.WeightKeyMemoryTotal:     16,
			ultron.WeightKeyMemoryAvailable: 8,
		},
	}

	pod := ultron.WeightedPod{
		Weights: map[string]float64{
			ultron.WeightKeyCpuRequested:    2,
			ultron.WeightKeyMemoryRequested: 4,
		},
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
		Annotations: map[string]string{
			ultron.AnnotationDiskType: "SSD",
		},
	}

	pod := ultron.WeightedPod{
		Annotations: map[string]string{
			ultron.AnnotationDiskType: "SSD",
		},
	}

	// Act
	score1 := alg.StorageScore(&node, &pod)
	expected1 := 1.0

	pod.Annotations[ultron.AnnotationDiskType] = "HDD"
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
		Annotations: map[string]string{
			ultron.AnnotationNetworkType: "10G",
		},
	}

	pod := ultron.WeightedPod{
		Annotations: map[string]string{
			ultron.AnnotationNetworkType: "10G",
		},
	}

	// Act
	score1 := alg.NetworkScore(&node, &pod)
	expected1 := 1.0

	pod.Annotations[ultron.AnnotationNetworkType] = "1G"
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
		Weights: map[string]float64{
			ultron.WeightKeyPrice:       10.0,
			ultron.WeightKeyPriceMedian: 8.0,
		},
	}

	// Act
	score1 := alg.PriceScore(&node)
	expected1 := 1.0 - (node.Weights[ultron.WeightKeyPriceMedian] / node.Weights[ultron.WeightKeyPrice])

	node.Weights[ultron.WeightKeyPrice] = 0
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
		Weights: map[string]float64{
			ultron.WeightKeyPrice:       10.0,
			ultron.WeightKeyPriceMedian: 8.0,
		},
		InterruptionRate: ultron.WeightedInteruptionRate{Weight: 0.1},
	}

	// Act
	score1 := alg.NodeScore(&node)
	expected1 := node.Weights[ultron.WeightKeyPrice] / (node.Weights[ultron.WeightKeyPriceMedian] + node.InterruptionRate.Weight)

	node.Weights[ultron.WeightKeyPrice] = 0
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
		Annotations: map[string]string{
			ultron.AnnotationWorkloadPriority: ultron.WorkloadPriorityHigh.String(),
		},
	}

	// Act
	score1 := alg.PodScore(&pod)
	expected1 := 1.0

	pod.Annotations[ultron.AnnotationWorkloadPriority] = ultron.WorkloadPriorityLow.String()
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
		Annotations: map[string]string{
			ultron.AnnotationDiskType:    "SSD",
			ultron.AnnotationNetworkType: "10G",
		},
		Weights: map[string]float64{
			ultron.WeightKeyCpuTotal:        8,
			ultron.WeightKeyCpuAvailable:    4,
			ultron.WeightKeyMemoryTotal:     16,
			ultron.WeightKeyMemoryAvailable: 8,
			ultron.WeightKeyPrice:           10.0,
			ultron.WeightKeyPriceMedian:     8.0,
		},
		InterruptionRate: ultron.WeightedInteruptionRate{Weight: 0.1},
	}

	pod := ultron.WeightedPod{
		Annotations: map[string]string{
			ultron.AnnotationDiskType:         "SSD",
			ultron.AnnotationNetworkType:      "10G",
			ultron.AnnotationWorkloadPriority: ultron.WorkloadPriorityHigh.String(),
		},
		Weights: map[string]float64{
			ultron.WeightKeyCpuRequested:    2,
			ultron.WeightKeyMemoryRequested: 4,
		},
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
