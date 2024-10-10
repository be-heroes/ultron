package services_test

import (
	"math"
	"testing"

	"github.com/be-heroes/ultron/mocks" // Import the generated mocks
	ultron "github.com/be-heroes/ultron/pkg"
	services "github.com/be-heroes/ultron/pkg/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
)

func intPtr(i int32) *int32         { return &i }
func float32Ptr(f float32) *float32 { return &f }
func stringPtr(s string) *string    { return &s }

func TestComputePodSpec_Success(t *testing.T) {
	// Arrange
	mockAlgorithm := new(mocks.IAlgorithm)
	mockCache := new(mocks.ICacheService)
	mockMapper := new(mocks.IMapper)

	service := services.NewComputeService(mockAlgorithm, mockCache, mockMapper)

	pod := &corev1.Pod{}

	// Set up mock behavior for Mapper
	mockMapper.On("MapPodToWeightedPod", pod).Return(ultron.WeightedPod{
		RequestedCPU:         2,
		RequestedMemory:      4,
		RequestedStorage:     50,
		RequestedDiskType:    "SSD",
		RequestedNetworkType: "isolated",
	}, nil)

	// Set up mock behavior for Cache
	mockCache.On("GetWeightedNodes").Return([]ultron.WeightedNode{
		{
			AvailableCPU:     4,
			TotalCPU:         4,
			AvailableMemory:  8,
			TotalMemory:      8,
			AvailableStorage: 50,
			DiskType:         "SSD",
			NetworkType:      "isolated",
		},
	}, nil)

	// Set up mock behavior for Algorithm
	mockAlgorithm.On("TotalScore", mock.AnythingOfType("*pkg.WeightedNode"), mock.AnythingOfType("*pkg.WeightedPod")).Return(1.0)

	// Act
	wNode, err := service.MatchPodSpec(pod)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, wNode)
	assert.Equal(t, 4.0, wNode.AvailableCPU)
	assert.Equal(t, "SSD", wNode.DiskType)
	assert.Equal(t, "isolated", wNode.NetworkType)

	// Assert expectations for all mocks
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

	// Set up mock behavior for Mapper
	mockMapper.On("MapPodToWeightedPod", pod).Return(ultron.WeightedPod{}, nil)

	// Set up mock behavior for Cache to return no nodes
	mockCache.On("GetWeightedNodes").Return([]ultron.WeightedNode{}, nil)

	// Since no nodes are returned, TotalScore shouldn't be called, but add a maybe return for safety
	mockAlgorithm.On("TotalScore", mock.Anything, mock.Anything).Maybe().Return(0.0)

	// Act
	wNode, err := service.MatchPodSpec(pod)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, wNode, "Expected wNode to be empty when no nodes are available")

	// Assert expectations for all mocks
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
		RequestedCPU:         1,
		RequestedMemory:      2,
		RequestedStorage:     30,
		RequestedDiskType:    "SSD",
		RequestedNetworkType: "isolated",
	}

	// Set up mock behavior for Cache
	mockCache.On("GetAllComputeConfigurations").Return([]ultron.ComputeConfiguration{
		{
			ComputeType: ultron.ComputeTypeDurable,
			VCpu:        intPtr(2),
			RamGb:       intPtr(8),
			VolumeGb:    intPtr(50),
			Cost: &ultron.ComputeCost{
				PricePerUnit: float32Ptr(0.2),
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
	assert.Equal(t, int32(2), *computeConfig.VCpu)
	assert.Equal(t, ultron.ComputeTypeDurable, computeConfig.ComputeType)

	// Assert expectations for all mocks
	mockCache.AssertExpectations(t)
}

func TestCalculateWeightedNodeMedianPrice_Success(t *testing.T) {
	// Arrange
	mockAlgorithm := new(mocks.IAlgorithm)
	mockCache := new(mocks.ICacheService)
	mockMapper := new(mocks.IMapper)

	service := services.NewComputeService(mockAlgorithm, mockCache, mockMapper)

	wNode := ultron.WeightedNode{
		AvailableCPU:     4,
		AvailableMemory:  8,
		AvailableStorage: 50,
		DiskType:         "SSD",
		NetworkType:      "isolated",
		Price:            10.0,
		MedianPrice:      5.0,
	}

	mockCache.On("GetAllComputeConfigurations").Return([]ultron.ComputeConfiguration{
		{
			ComputeType: ultron.ComputeTypeDurable,
			VCpu:        intPtr(2),
			RamGb:       intPtr(8),
			VolumeGb:    intPtr(50),
			Cost: &ultron.ComputeCost{
				PricePerUnit: float32Ptr(0.2),
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

	// Assert expectations for all mocks
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
		RequestedCPU:    2,
		RequestedMemory: 4,
	}

	// Set up mock behavior for Cache
	mockCache.On("GetWeightedNodes").Return([]ultron.WeightedNode{
		{
			AvailableCPU:    4,
			AvailableMemory: 8,
			DiskType:        "SSD",
			NetworkType:     "isolated",
		},
	}, nil)

	// Set up mock behavior for Algorithm
	mockAlgorithm.On("TotalScore", mock.AnythingOfType("*pkg.WeightedNode"), mock.AnythingOfType("*pkg.WeightedPod")).Return(1.0)

	// Act
	wNode, err := service.MatchWeightedPodToWeightedNode(&wPod)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, wNode, "Expected a weighted node to be found")
	assert.Equal(t, 4.0, wNode.AvailableCPU)

	// Assert expectations for all mocks
	mockCache.AssertExpectations(t)
	mockAlgorithm.AssertExpectations(t)
}
