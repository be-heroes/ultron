package services_test

import (
	"math"
	"testing"

	"github.com/be-heroes/ultron/mocks"
	ultron "github.com/be-heroes/ultron/pkg"
	services "github.com/be-heroes/ultron/pkg/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
)

func int64Ptr(i int64) *int64       { return &i }
func float64Ptr(f float64) *float64 { return &f }
func stringPtr(s string) *string    { return &s }

func TestComputePodSpec_Success(t *testing.T) {
	// Arrange
	mockAlgorithm := new(mocks.IAlgorithm)
	mockCache := new(mocks.ICacheService)
	mockMapper := new(mocks.IMapper)

	service := services.NewComputeService(mockAlgorithm, mockCache, mockMapper)

	pod := &corev1.Pod{}

	mockMapper.On("MapPodToWeightedPod", pod).Return(ultron.WeightedPod{
		Annotations: map[string]string{
			ultron.AnnotationDiskType:      "SSD",
			ultron.AnnotationNetworkType:   "isolated",
			ultron.AnnotationStorageSizeGb: "50",
		},
		Weights: map[string]float64{
			ultron.WeightKeyCpuRequested:     2,
			ultron.WeightKeyMemoryRequested:  4,
			ultron.WeightKeyStorageRequested: 50,
		},
	}, nil)

	mockCache.On("GetWeightedNodes").Return([]ultron.WeightedNode{
		{
			Annotations: map[string]string{
				ultron.AnnotationDiskType:      "SSD",
				ultron.AnnotationNetworkType:   "isolated",
				ultron.AnnotationStorageSizeGb: "50",
			},
			Weights: map[string]float64{
				ultron.WeightKeyCpuAvailable:     4,
				ultron.WeightKeyCpuTotal:         4,
				ultron.WeightKeyMemoryAvailable:  8,
				ultron.WeightKeyMemoryTotal:      8,
				ultron.WeightKeyStorageAvailable: 50,
			},
		},
	}, nil)

	mockAlgorithm.On("TotalScore", mock.AnythingOfType("*pkg.WeightedNode"), mock.AnythingOfType("*pkg.WeightedPod")).Return(1.0)

	// Act
	wNode, err := service.MatchPodSpec(pod)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, wNode)
	assert.Equal(t, 4.0, wNode.Weights[ultron.WeightKeyCpuAvailable])
	assert.Equal(t, "SSD", wNode.Annotations[ultron.AnnotationDiskType])
	assert.Equal(t, "isolated", wNode.Annotations[ultron.AnnotationNetworkType])

	mockMapper.AssertExpectations(t)
	mockCache.AssertExpectations(t)
	mockAlgorithm.AssertExpectations(t)
}

func TestComputePodSpec_NoWeightedNode(t *testing.T) {
	// Arrange
	mockAlgorithm := new(mocks.IAlgorithm)
	mockCache := new(mocks.ICacheService)
	mockMapper := new(mocks.IMapper)

	service := services.NewComputeService(mockAlgorithm, mockCache, mockMapper)

	pod := &corev1.Pod{}

	mockMapper.On("MapPodToWeightedPod", pod).Return(ultron.WeightedPod{}, nil)

	mockCache.On("GetWeightedNodes").Return([]ultron.WeightedNode{}, nil)

	mockAlgorithm.On("TotalScore", mock.Anything, mock.Anything).Maybe().Return(0.0)

	// Act
	wNode, err := service.MatchPodSpec(pod)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, wNode, "Expected wNode to be empty when no nodes are available")

	mockMapper.AssertExpectations(t)
	mockCache.AssertExpectations(t)
	mockAlgorithm.AssertExpectations(t)
}

func TestMatchWeightedPodToComputeConfiguration_Success(t *testing.T) {
	// Arrange
	mockAlgorithm := new(mocks.IAlgorithm)
	mockCache := new(mocks.ICacheService)
	mockMapper := new(mocks.IMapper)

	service := services.NewComputeService(mockAlgorithm, mockCache, mockMapper)

	wPod := ultron.WeightedPod{
		Annotations: map[string]string{
			ultron.AnnotationDiskType:      "SSD",
			ultron.AnnotationNetworkType:   "isolated",
			ultron.AnnotationStorageSizeGb: "30",
		},
		Weights: map[string]float64{
			ultron.WeightKeyCpuRequested:     1,
			ultron.WeightKeyMemoryRequested:  2,
			ultron.WeightKeyStorageRequested: 30,
		},
	}

	mockCache.On("GetAllComputeConfigurations").Return([]ultron.ComputeConfiguration{
		{
			ComputeType: ultron.ComputeTypeDurable,
			VCpu:        int64Ptr(2),
			RamGb:       int64Ptr(8),
			VolumeGb:    int64Ptr(50),
			Cost: &ultron.ComputeCost{
				PricePerUnit: float64Ptr(0.2),
			},
			VolumeType:        stringPtr("SSD"),
			CloudNetworkTypes: []string{"isolated", "public"},
		},
	}, nil)

	// Act
	computeConfig, err := service.MatchWeightedPodToComputeConfiguration(&wPod)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, computeConfig)
	assert.Equal(t, int64(2), *computeConfig.VCpu)
	assert.Equal(t, ultron.ComputeTypeDurable, computeConfig.ComputeType)

	mockCache.AssertExpectations(t)
}

func TestCalculateWeightedNodeMedianPrice_Success(t *testing.T) {
	// Arrange
	mockAlgorithm := new(mocks.IAlgorithm)
	mockCache := new(mocks.ICacheService)
	mockMapper := new(mocks.IMapper)

	service := services.NewComputeService(mockAlgorithm, mockCache, mockMapper)

	wNode := ultron.WeightedNode{
		Annotations: map[string]string{
			ultron.AnnotationDiskType:      "SSD",
			ultron.AnnotationNetworkType:   "isolated",
			ultron.AnnotationStorageSizeGb: "50",
		},
		Weights: map[string]float64{
			ultron.WeightKeyCpuAvailable:     4,
			ultron.WeightKeyCpuTotal:         4,
			ultron.WeightKeyMemoryAvailable:  8,
			ultron.WeightKeyMemoryTotal:      8,
			ultron.WeightKeyStorageAvailable: 50,
			ultron.WeightKeyPrice:            10.0,
			ultron.WeightKeyPriceMedian:      5.0,
		},
	}

	mockCache.On("GetAllComputeConfigurations").Return([]ultron.ComputeConfiguration{
		{
			ComputeType: ultron.ComputeTypeDurable,
			VCpu:        int64Ptr(2),
			RamGb:       int64Ptr(8),
			VolumeGb:    int64Ptr(50),
			Cost: &ultron.ComputeCost{
				PricePerUnit: float64Ptr(0.2),
			},
			VolumeType:        stringPtr("SSD"),
			CloudNetworkTypes: []string{"isolated", "public"},
		},
	}, nil)

	// Act
	medianPrice, err := service.CalculateWeightedNodeMedianPrice(&wNode)

	// Assert
	assert.NoError(t, err)
	assert.False(t, math.IsNaN(medianPrice), "Median price should not be NaN")
	assert.InDelta(t, 0.2, medianPrice, 0.01, "Median price did not match expected value")

	mockCache.AssertExpectations(t)
	mockAlgorithm.AssertExpectations(t)
}

func TestMatchWeightedPodToWeightedNode_Success(t *testing.T) {
	// Arrange
	mockAlgorithm := new(mocks.IAlgorithm)
	mockCache := new(mocks.ICacheService)
	mockMapper := new(mocks.IMapper)

	service := services.NewComputeService(mockAlgorithm, mockCache, mockMapper)

	wPod := ultron.WeightedPod{
		Weights: map[string]float64{
			ultron.WeightKeyCpuRequested:    2,
			ultron.WeightKeyMemoryRequested: 4,
		},
	}

	mockCache.On("GetWeightedNodes").Return([]ultron.WeightedNode{
		{
			Annotations: map[string]string{
				ultron.AnnotationDiskType:    "SSD",
				ultron.AnnotationNetworkType: "isolated",
			},
			Weights: map[string]float64{
				ultron.WeightKeyCpuAvailable:    4,
				ultron.WeightKeyMemoryAvailable: 8,
			},
		},
	}, nil)

	mockAlgorithm.On("TotalScore", mock.AnythingOfType("*pkg.WeightedNode"), mock.AnythingOfType("*pkg.WeightedPod")).Return(1.0)

	// Act
	wNode, err := service.MatchWeightedPodToWeightedNode(&wPod)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, wNode, "Expected a weighted node to be found")
	assert.Equal(t, 4.0, wNode.Weights[ultron.WeightKeyCpuAvailable])

	mockCache.AssertExpectations(t)
	mockAlgorithm.AssertExpectations(t)
}
