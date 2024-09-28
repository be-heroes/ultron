package ultron

import (
	"math"
	"slices"
	"sort"
	"strconv"

	emma "github.com/emma-community/emma-go-sdk"
	corev1 "k8s.io/api/core/v1"
)

const (
	DefaultDiskType            = "SSD"
	DefaultNetworkType         = "isolated"
	DefaultStorageSizeGB       = 10.0
	DefaultPriority            = PriorityLow
	DefaultDurableInstanceType = "emma.durable"
	DefaultSpotInstanceType    = "emma.spot"
)

func ComputePodSpec(pod *corev1.Pod) (*WeightedNode, error) {
	wPod, err := MapPodToWeightedPod(pod)
	if err != nil {
		return nil, err
	}

	wNode, err := MatchWeightedPodToWeightedNode(wPod)
	if err != nil {
		return nil, err
	}

	if wNode == nil {
		vmConfiguration, err := MatchWeightedPodToVmConfiguration(wPod)
		if err != nil {
			return nil, err
		}

		instanceType := DefaultDurableInstanceType
		spotConfigurations, err := GetSpotVmConfigurationsFromCache()
		if err != nil {
			return nil, err
		}

		for _, config := range spotConfigurations {
			if config.Id == vmConfiguration.Id {
				instanceType = DefaultSpotInstanceType

				break
			}
		}

		wNode = &WeightedNode{
			Selector:         map[string]string{"node.kubernetes.io/instance-type": instanceType},
			AvailableCPU:     float64(*vmConfiguration.VCpu),
			TotalCPU:         float64(*vmConfiguration.VCpu),
			AvailableMemory:  float64(*vmConfiguration.RamGb),
			TotalMemory:      float64(*vmConfiguration.RamGb),
			AvailableStorage: float64(*vmConfiguration.VolumeGb),
			DiskType:         wPod.RequestedDiskType,
			NetworkType:      wPod.RequestedNetworkType,
			Price:            float64(*vmConfiguration.Cost.PricePerUnit),
			InstanceType:     instanceType,
			InterruptionRate: 0, // TODO: Talk with External API team about possibility of calculating this value based on historic metrics in the backend and expose it in the VmConfiguration struct
		}
	}

	return wNode, nil
}

func MatchWeightedPodToVmConfiguration(wPod WeightedPod) (*emma.VmConfiguration, error) {
	var suitableConfigs []emma.VmConfiguration
	vmConfigurations, err := GetAllVmConfigurationsFromCache()
	if err != nil {
		return nil, err
	}

	for _, config := range vmConfigurations {
		if vmConfigurationMatchesWeightedPodRequirements(config, wPod) {
			suitableConfigs = append(suitableConfigs, config)
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

func MatchWeightedNodeToVmConfiguration(wNode WeightedNode) (*emma.VmConfiguration, error) {
	var suitableConfigs []emma.VmConfiguration
	vmConfigurations, err := GetAllVmConfigurationsFromCache()
	if err != nil {
		return nil, err
	}

	for _, config := range vmConfigurations {
		if vmConfigurationMatchesWeightedNodeRequirements(config, wNode) {
			suitableConfigs = append(suitableConfigs, config)
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

func MatchWeightedPodToWeightedNode(pod WeightedPod) (*WeightedNode, error) {
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

		score := TotalScore(wNode, pod)

		if score > highestScore {
			highestScore = score
			match = wNode
		}
	}

	return &match, nil
}

func CalculateWeightedNodeMedianPrice(wNode WeightedNode) (float64, error) {
	var totalCost float64
	var matchCount int32
	allVmConfigurations, err := GetAllVmConfigurationsFromCache()
	if err != nil {
		return 0, err
	}

	for _, vmConfiguration := range allVmConfigurations {
		if vmConfigurationMatchesWeightedNodeRequirements(vmConfiguration, wNode) && vmConfiguration.Cost != nil && vmConfiguration.Cost.PricePerUnit != nil {
			totalCost += float64(*vmConfiguration.Cost.PricePerUnit)
			matchCount++
		}
	}

	return totalCost / float64(matchCount), nil
}

func vmConfigurationMatchesWeightedNodeRequirements(vmConfiguration emma.VmConfiguration, wNode WeightedNode) bool {
	if float64(*vmConfiguration.VCpu) < wNode.AvailableCPU {
		return false
	}

	if float64(*vmConfiguration.RamGb) < wNode.AvailableMemory {
		return false
	}

	if float64(*vmConfiguration.VolumeGb) < wNode.AvailableStorage {
		return false
	}

	if (*vmConfiguration.VolumeType) != wNode.DiskType {
		return false
	}

	if !slices.Contains(vmConfiguration.CloudNetworkTypes, wNode.NetworkType) {
		return false
	}

	return true
}

func vmConfigurationMatchesWeightedPodRequirements(vmConfiguration emma.VmConfiguration, wPod WeightedPod) bool {
	if float64(*vmConfiguration.VCpu) < wPod.RequestedCPU {
		return false
	}

	if float64(*vmConfiguration.RamGb) < wPod.RequestedMemory {
		return false
	}

	if float64(*vmConfiguration.VolumeGb) < wPod.RequestedStorage {
		return false
	}

	if (*vmConfiguration.VolumeType) != wPod.RequestedDiskType {
		return false
	}

	if !slices.Contains(vmConfiguration.CloudNetworkTypes, wPod.RequestedNetworkType) {
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
