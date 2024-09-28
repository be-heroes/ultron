package ultron

import (
	"math"
	"slices"
	"sort"
	"strconv"

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

var ComputePodSpec = func(pod *corev1.Pod) (*WeightedNode, error) {
	wPod, err := MapPodToWeightedPod(pod)
	if err != nil {
		return nil, err
	}

	wNode, err := MatchWeightedPodToWeightedNode(wPod)
	if err != nil {
		return nil, err
	}

	if wNode == nil {
		computeConfiguration, err := MatchWeightedPodToComputeConfiguration(wPod)
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
			InterruptionRate: 0, // TODO: Talk with External API team about possibility of calculating this value based on historic metrics in the backend and expose it in the VmConfiguration struct
		}
	}

	return wNode, nil
}

var MatchWeightedPodToComputeConfiguration = func(wPod WeightedPod) (*ComputeConfiguration, error) {
	var suitableConfigs []ComputeConfiguration
	computeConfigurations, err := GetAllComputeConfigurationsFromCache()
	if err != nil {
		return nil, err
	}

	for _, computeConfig := range computeConfigurations {
		if computeConfigurationMatchesWeightedPodRequirements(computeConfig, wPod) {
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

var MatchWeightedNodeToComputeConfiguration = func(wNode WeightedNode) (*ComputeConfiguration, error) {
	var suitableConfigs []ComputeConfiguration
	computeConfigurations, err := GetAllComputeConfigurationsFromCache()
	if err != nil {
		return nil, err
	}

	for _, computeConfig := range computeConfigurations {
		if computeConfigurationMatchesWeightedNodeRequirements(computeConfig, wNode) {
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

var MatchWeightedPodToWeightedNode = func(pod WeightedPod) (*WeightedNode, error) {
	wNodes, err := GetWeightedNodesFromCache()
	if err != nil {
		return nil, err
	}

	var match WeightedNode
	highestScore := math.Inf(-1)

	for _, wNode := range wNodes {
		if wNode.AvailableCPU < pod.RequestedCPU || wNode.AvailableMemory < pod.RequestedMemory {
			continue
		}

		score := Score(wNode, pod)

		if score > highestScore {
			highestScore = score
			match = wNode
		}
	}

	return &match, nil
}

var CalculateWeightedNodeMedianPrice = func(wNode WeightedNode) (float64, error) {
	var totalCost float64
	var matchCount int32
	computeConfigurations, err := GetAllComputeConfigurationsFromCache()
	if err != nil {
		return 0, err
	}

	for _, computeConfig := range computeConfigurations {
		if computeConfigurationMatchesWeightedNodeRequirements(computeConfig, wNode) && computeConfig.Cost != nil && computeConfig.Cost.PricePerUnit != nil {
			totalCost += float64(*computeConfig.Cost.PricePerUnit)
			matchCount++
		}
	}

	return totalCost / float64(matchCount), nil
}

func computeConfigurationMatchesWeightedNodeRequirements(computeConfiguration ComputeConfiguration, wNode WeightedNode) bool {
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

func computeConfigurationMatchesWeightedPodRequirements(computeConfiguration ComputeConfiguration, wPod WeightedPod) bool {
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

func getAnnotationOrDefault(annotations map[string]string, key, defaultValue string) string {
	if value, exists := annotations[key]; exists {
		return value
	}

	return defaultValue
}

func getFloatAnnotationOrDefault(annotations map[string]string, key string, defaultValue float64) float64 {
	if valueStr, exists := annotations[key]; exists {
		if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
			return value
		}
	}

	return defaultValue
}

func getPriorityFromAnnotation(annotations map[string]string) PriorityEnum {
	if value, exists := annotations[AnnotationPriority]; exists {
		switch value {
		case "true":
			return PriorityHigh
		case "false":
			return PriorityLow
		}
	}

	return DefaultPriority
}
