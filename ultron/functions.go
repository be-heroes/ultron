package ultron

import (
	"fmt"
	"slices"
	"sort"
	"strconv"

	emmaSdk "github.com/emma-community/emma-go-sdk"
	"github.com/patrickmn/go-cache"
	corev1 "k8s.io/api/core/v1"
)

const (
	AnnotationDiskType    = "ultron.io/disk-type"
	AnnotationNetworkType = "ultron.io/network-type"
	AnnotationStorageSize = "ultron.io/storage-size"
	AnnotationPriority    = "ultron.io/priority"

	DefaultDiskType            = "SSD"
	DefaultNetworkType         = "isolated"
	DefaultStorageSizeGB       = 10.0
	DefaultPriority            = PriorityLow
	DefaultDurableInstanceType = "emma.durable"
	DefaultSpotInstanceType    = "emma.spot"
)

var Cache = cache.New(cache.NoExpiration, cache.NoExpiration)

func ComputePodSpec(pod *corev1.Pod) (*WeightedNode, error) {
	weightedNodesInterface, found := Cache.Get(CacheKeyWeightedNodes)

	if !found {
		return nil, fmt.Errorf("failed to get weighted nodes from cache")
	}

	wPod, err := MapK8sPodToWeightedPod(pod)
	if err != nil {
		return nil, err
	}

	bestNode := FindBestNode(wPod, weightedNodesInterface.([]WeightedNode))

	if bestNode == nil {
		durableConfigsInterface, found := Cache.Get(CacheKeyDurableVmConfigurations)

		if !found {
			return nil, fmt.Errorf("failed to get durable VmConfiguration list from cache")
		}

		spotConfigsInterface, found := Cache.Get(CacheKeySpotVmConfigurations)

		if !found {
			return nil, fmt.Errorf("failed to get spot VmConfiguration list from cache")
		}

		durableConfigs := durableConfigsInterface.([]emmaSdk.VmConfiguration)
		spotConfigs := spotConfigsInterface.([]emmaSdk.VmConfiguration)
		bestVmConfig := findBestVmConfiguration(wPod, append(durableConfigs, spotConfigs...))

		if bestVmConfig == nil {
			return nil, fmt.Errorf("no suitable VmConfiguration found")
		}

		var instanceType = DefaultDurableInstanceType

		for _, config := range spotConfigs {
			if config.Id == bestVmConfig.Id {
				instanceType = DefaultSpotInstanceType

				break
			}
		}

		bestNode = &WeightedNode{
			Selector:         map[string]string{"node.kubernetes.io/instance-type": instanceType},
			AvailableCPU:     float64(*bestVmConfig.VCpu),
			TotalCPU:         float64(*bestVmConfig.VCpu),
			AvailableMemory:  float64(*bestVmConfig.RamGb),
			TotalMemory:      float64(*bestVmConfig.RamGb),
			AvailableStorage: float64(*bestVmConfig.VolumeGb),
			DiskType:         wPod.RequestedDiskType,
			NetworkType:      wPod.RequestedNetworkType,
			Price:            float64(*bestVmConfig.Cost.PricePerUnit),
			InstanceType:     instanceType,
			InterruptionRate: 0, // TODO: Talk with External API team about possibility of calculating this value based on historic metrics in the backend and expose it in the VmConfiguration struct
		}
	}

	return bestNode, nil
}

func findBestVmConfiguration(wPod WeightedPod, configs []emmaSdk.VmConfiguration) *emmaSdk.VmConfiguration {
	var suitableConfigs []emmaSdk.VmConfiguration

	for _, config := range configs {
		if vmConfigurationMatchesPodRequirements(config, wPod) {
			suitableConfigs = append(suitableConfigs, config)
		}
	}

	if len(suitableConfigs) == 0 {
		return nil
	}

	sort.Slice(suitableConfigs, func(i, j int) bool {
		return (*suitableConfigs[i].Cost.PricePerUnit) < (*suitableConfigs[j].Cost.PricePerUnit)
	})

	return &suitableConfigs[0]
}

func vmConfigurationMatchesPodRequirements(config emmaSdk.VmConfiguration, wPod WeightedPod) bool {
	if float64(*config.VCpu) < wPod.RequestedCPU {
		return false
	}

	if float64(*config.RamGb) < wPod.RequestedMemory {
		return false
	}

	if float64(*config.VolumeGb) < wPod.RequestedStorage {
		return false
	}

	if (*config.VolumeType) != wPod.RequestedDiskType {
		return false
	}

	if !slices.Contains(config.CloudNetworkTypes, wPod.RequestedNetworkType) {
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
