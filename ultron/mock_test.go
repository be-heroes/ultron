package ultron_test

import (
	"time"

	ultron "ultron/ultron"

	emma "github.com/emma-community/emma-go-sdk"
	corev1 "k8s.io/api/core/v1"
)

func intPtr(i int32) *int32         { return &i }
func float32Ptr(f float32) *float32 { return &f }
func stringPtr(s string) *string    { return &s }

type MockComputeService struct{}

func (mcs *MockComputeService) ComputePodSpec(pod *corev1.Pod) (*ultron.WeightedNode, error) {
	return &ultron.WeightedNode{
		Selector: map[string]string{
			"node-type": "mock-node",
		},
	}, nil
}

func (mcs *MockComputeService) CalculateWeightedNodeMedianPrice(node ultron.WeightedNode) (float64, error) {
	return 0, nil
}

func (mcs *MockComputeService) ComputeConfigurationMatchesWeightedNodeRequirements(configuration ultron.ComputeConfiguration, node ultron.WeightedNode) bool {
	return true
}

func (mcs *MockComputeService) ComputeConfigurationMatchesWeightedPodRequirements(configuration ultron.ComputeConfiguration, pod ultron.WeightedPod) bool {
	return true
}

func (mcs *MockComputeService) MatchWeightedNodeToComputeConfiguration(node ultron.WeightedNode) (*ultron.ComputeConfiguration, error) {
	return &ultron.ComputeConfiguration{}, nil
}

func (mcs *MockComputeService) MatchWeightedPodToComputeConfiguration(node ultron.WeightedPod) (*ultron.ComputeConfiguration, error) {
	return &ultron.ComputeConfiguration{}, nil
}

func (mcs *MockComputeService) MatchWeightedPodToWeightedNode(pod ultron.WeightedPod) (*ultron.WeightedNode, error) {
	return nil, nil
}

type MockAlgorithm struct{}

func (ma *MockAlgorithm) DiskTypeScore(wNode ultron.WeightedNode, wPod ultron.WeightedPod) float64 {
	return 0
}

func (ma *MockAlgorithm) NetworkTypeScore(wNode ultron.WeightedNode, wPod ultron.WeightedPod) float64 {
	return 0
}

func (ma *MockAlgorithm) NodeStabilityScore(wNode ultron.WeightedNode) float64 {
	return 0
}

func (ma *MockAlgorithm) PriceScore(wNode ultron.WeightedNode) float64 {
	return 0.2
}

func (ma *MockAlgorithm) ResourceFitScore(wNode ultron.WeightedNode, wPod ultron.WeightedPod) float64 {
	return wNode.AvailableCPU - wPod.RequestedCPU
}

func (ma *MockAlgorithm) WorkloadPriorityScore(wPod ultron.WeightedPod) float64 {
	return 0
}

func (ma *MockAlgorithm) Score(wNode ultron.WeightedNode, wPod ultron.WeightedPod) float64 {
	return wNode.AvailableCPU - wPod.RequestedCPU
}

type MockCache struct{}

func (mc *MockCache) AddCacheItem(key string, value interface{}, time time.Duration) {
}

func (mc *MockCache) GetAllComputeConfigurations() ([]ultron.ComputeConfiguration, error) {
	return []ultron.ComputeConfiguration{
		{
			ComputeType: ultron.ComputeTypeDurable,
			VmConfiguration: emma.VmConfiguration{
				VCpu:     intPtr(2),
				RamGb:    intPtr(8),
				VolumeGb: intPtr(50),
				Cost: &emma.VmConfigurationCost{
					PricePerUnit: float32Ptr(0.2),
				},
				VolumeType:        stringPtr("SSD"),
				CloudNetworkTypes: []string{"isolated", "public"},
			},
		},
	}, nil
}

func (mc *MockCache) GetEphemeralComputeConfigurations() ([]ultron.ComputeConfiguration, error) {
	return nil, nil
}

func (mc *MockCache) GetDurableComputeConfigurations() ([]ultron.ComputeConfiguration, error) {
	return nil, nil
}

func (mc *MockCache) GetWeightedNodes() ([]ultron.WeightedNode, error) {
	return []ultron.WeightedNode{
		{
			AvailableCPU:     4,
			TotalCPU:         4,
			AvailableMemory:  8,
			TotalMemory:      8,
			AvailableStorage: 50,
			DiskType:         "SSD",
			NetworkType:      "isolated",
		},
	}, nil
}

type MockMapper struct{}

func (mm *MockMapper) MapPodToWeightedPod(pod *corev1.Pod) (ultron.WeightedPod, error) {
	return ultron.WeightedPod{
		RequestedCPU:         2,
		RequestedMemory:      4,
		RequestedStorage:     50,
		RequestedDiskType:    "SSD",
		RequestedNetworkType: "isolated",
	}, nil
}

func (mm *MockMapper) MapNodeToWeightedNode(node *corev1.Node) (ultron.WeightedNode, error) {
	return ultron.WeightedNode{
		AvailableCPU:     2,
		TotalCPU:         4,
		AvailableMemory:  8,
		TotalMemory:      16,
		AvailableStorage: 100,
		TotalStorage:     200,
		DiskType:         "SSD",
		NetworkType:      "5G",
	}, nil
}
