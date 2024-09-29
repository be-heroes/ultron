package ultron

import (
	"math"
	"slices"
	"sort"

	corev1 "k8s.io/api/core/v1"
)

const (
	DefaultDiskType              = "SSD"
	DefaultNetworkType           = "isolated"
	DefaultStorageSizeGB         = 10.0
	DefaultPriority              = PriorityLow
	DefaultDurableInstanceType   = "ultron.durable"
	DefaultEphemeralInstanceType = "ultron.ephemeral"
)

type ComputeService interface {
	ComputePodSpec(pod *corev1.Pod) (*WeightedNode, error)
	MatchWeightedPodToComputeConfiguration(wPod WeightedPod) (*ComputeConfiguration, error)
	MatchWeightedNodeToComputeConfiguration(wNode WeightedNode) (*ComputeConfiguration, error)
	MatchWeightedPodToWeightedNode(wPod WeightedPod) (*WeightedNode, error)
	CalculateWeightedNodeMedianPrice(wNode WeightedNode) (float64, error)
	ComputeConfigurationMatchesWeightedNodeRequirements(computeConfiguration ComputeConfiguration, wNode WeightedNode) bool
	ComputeConfigurationMatchesWeightedPodRequirements(computeConfiguration ComputeConfiguration, wPod WeightedPod) bool
}

type IComputeService struct {
	algorithm Algorithm
	cache     Cache
	mapper    Mapper
}

func NewIComputeService(algorithm Algorithm, cache Cache, mapper Mapper) *IComputeService {
	return &IComputeService{
		algorithm: algorithm,
		cache:     cache,
		mapper:    mapper,
	}
}

func (cs IComputeService) ComputePodSpec(pod *corev1.Pod) (*WeightedNode, error) {
	wPod, err := cs.mapper.MapPodToWeightedPod(pod)
	if err != nil {
		return nil, err
	}

	wNode, err := cs.MatchWeightedPodToWeightedNode(wPod)
	if err != nil {
		return nil, err
	}

	if wNode == nil {
		computeConfiguration, err := cs.MatchWeightedPodToComputeConfiguration(wPod)
		if err != nil {
			return nil, err
		}

		var instanceType string

		if computeConfiguration.ComputeType == ComputeTypeDurable {
			instanceType = DefaultDurableInstanceType
		} else {
			instanceType = DefaultEphemeralInstanceType
		}

		wNode = &WeightedNode{
			Selector:         map[string]string{LabelInstanceType: instanceType},
			AvailableCPU:     float64(*computeConfiguration.VCpu),
			TotalCPU:         float64(*computeConfiguration.VCpu),
			AvailableMemory:  float64(*computeConfiguration.RamGb),
			TotalMemory:      float64(*computeConfiguration.RamGb),
			AvailableStorage: float64(*computeConfiguration.VolumeGb),
			DiskType:         wPod.RequestedDiskType,
			NetworkType:      wPod.RequestedNetworkType,
			Price:            float64(*computeConfiguration.Cost.PricePerUnit),
			InstanceType:     instanceType,
			// InterruptionRate: *float64(*computeConfiguration.InterruptionRate),
			InterruptionRate: 0, // TODO: Replace with above once External API is updated
		}
	}

	return wNode, nil
}

func (cs IComputeService) MatchWeightedPodToComputeConfiguration(wPod WeightedPod) (*ComputeConfiguration, error) {
	var suitableConfigs []ComputeConfiguration
	computeConfigurations, err := cs.cache.GetAllComputeConfigurations()
	if err != nil {
		return nil, err
	}

	for _, computeConfig := range computeConfigurations {
		if cs.ComputeConfigurationMatchesWeightedPodRequirements(computeConfig, wPod) {
			suitableConfigs = append(suitableConfigs, computeConfig)
		}
	}

	if len(suitableConfigs) == 0 {
		return nil, nil
	}

	sort.Slice(suitableConfigs, func(i, j int) bool {
		return (*suitableConfigs[i].Cost.PricePerUnit) < (*suitableConfigs[j].Cost.PricePerUnit)
	})

	return &suitableConfigs[0], nil
}

func (cs IComputeService) MatchWeightedNodeToComputeConfiguration(wNode WeightedNode) (*ComputeConfiguration, error) {
	var suitableConfigs []ComputeConfiguration
	computeConfigurations, err := cs.cache.GetAllComputeConfigurations()
	if err != nil {
		return nil, err
	}

	for _, computeConfig := range computeConfigurations {
		if cs.ComputeConfigurationMatchesWeightedNodeRequirements(computeConfig, wNode) {
			suitableConfigs = append(suitableConfigs, computeConfig)
		}
	}

	if len(suitableConfigs) == 0 {
		return nil, nil
	}

	sort.Slice(suitableConfigs, func(i, j int) bool {
		return (*suitableConfigs[i].Cost.PricePerUnit) < (*suitableConfigs[j].Cost.PricePerUnit)
	})

	return &suitableConfigs[0], nil
}

func (cs IComputeService) MatchWeightedPodToWeightedNode(pod WeightedPod) (*WeightedNode, error) {
	wNodes, err := cs.cache.GetWeightedNodes()
	if err != nil {
		return nil, err
	}

	var match WeightedNode
	highestScore := math.Inf(-1)

	for _, wNode := range wNodes {
		if wNode.AvailableCPU < pod.RequestedCPU || wNode.AvailableMemory < pod.RequestedMemory {
			continue
		}

		score := cs.algorithm.Score(wNode, pod)

		if score > highestScore {
			highestScore = score
			match = wNode
		}
	}

	return &match, nil
}

func (cs IComputeService) CalculateWeightedNodeMedianPrice(wNode WeightedNode) (float64, error) {
	var totalCost float64
	var matchCount int32
	computeConfigurations, err := cs.cache.GetAllComputeConfigurations()
	if err != nil {
		return 0, err
	}

	for _, computeConfig := range computeConfigurations {
		if cs.ComputeConfigurationMatchesWeightedNodeRequirements(computeConfig, wNode) && computeConfig.Cost != nil && computeConfig.Cost.PricePerUnit != nil {
			totalCost += float64(*computeConfig.Cost.PricePerUnit)
			matchCount++
		}
	}

	return totalCost / float64(matchCount), nil
}

func (cs IComputeService) ComputeConfigurationMatchesWeightedNodeRequirements(computeConfiguration ComputeConfiguration, wNode WeightedNode) bool {
	if float64(*computeConfiguration.VCpu) < wNode.AvailableCPU {
		return false
	}

	if float64(*computeConfiguration.RamGb) < wNode.AvailableMemory {
		return false
	}

	if float64(*computeConfiguration.VolumeGb) < wNode.AvailableStorage {
		return false
	}

	if (*computeConfiguration.VolumeType) != wNode.DiskType {
		return false
	}

	if !slices.Contains(computeConfiguration.CloudNetworkTypes, wNode.NetworkType) {
		return false
	}

	return true
}

func (cs IComputeService) ComputeConfigurationMatchesWeightedPodRequirements(computeConfiguration ComputeConfiguration, wPod WeightedPod) bool {
	if float64(*computeConfiguration.VCpu) < wPod.RequestedCPU {
		return false
	}

	if float64(*computeConfiguration.RamGb) < wPod.RequestedMemory {
		return false
	}

	if float64(*computeConfiguration.VolumeGb) < wPod.RequestedStorage {
		return false
	}

	if (*computeConfiguration.VolumeType) != wPod.RequestedDiskType {
		return false
	}

	if !slices.Contains(computeConfiguration.CloudNetworkTypes, wPod.RequestedNetworkType) {
		return false
	}

	return true
}
