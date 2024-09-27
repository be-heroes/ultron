package ultron

import (
	"fmt"
	"log"
	"strconv"
	"time"

	emmaSdk "github.com/emma-community/emma-go-sdk"
	"github.com/patrickmn/go-cache"
	corev1 "k8s.io/api/core/v1"
)

// TODO: Add logic to repopulate cache when items expire
var Cache = cache.New(10*time.Minute, 20*time.Minute)

func ComputePodSpec(pod corev1.Pod) *WeightedNode {
	weightedNodesInterface, found := Cache.Get("weightedNodes")

	if !found {
		log.Fatalf("Failed to get weighted nodes from cache")
	}

	wPod, err := mapK8sPodToWeightedPod(pod)
	if err != nil {
		log.Fatalf("Error mapping pod: %v", err)
	}

	bestNode := FindBestNode(wPod, weightedNodesInterface.([]WeightedNode))

	if bestNode == nil {
		durableConfigsInterface, found := Cache.Get("durableConfigs")

		if !found {
			log.Fatalf("Failed to get durable VmConfiguration list from cache")
		}

		_ = durableConfigsInterface.([]emmaSdk.VmConfiguration)

		spotConfigsInterface, found := Cache.Get("spotConfigs")

		if !found {
			log.Fatalf("Failed to get spot VmConfiguration listfrom cache")
		}

		_ = spotConfigsInterface.([]emmaSdk.VmConfiguration)

		//TODO: Implement logic to match a node type from a known VmConfiguration
		bestNode = &WeightedNode{}
	}

	return bestNode
}

func mapK8sPodToWeightedPod(k8sPod corev1.Pod) (WeightedPod, error) {
	// TODO: Implement logic for mapping multiple containers in a pod
	cpuRequest := k8sPod.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU]
	memRequest := k8sPod.Spec.Containers[0].Resources.Requests[corev1.ResourceMemory]
	cpuLimit := k8sPod.Spec.Containers[0].Resources.Limits[corev1.ResourceCPU]
	memLimit := k8sPod.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory]

	cpuRequestFloat, err := strconv.ParseFloat(cpuRequest.AsDec().String(), 64)
	if err != nil {
		return WeightedPod{}, fmt.Errorf("failed to parse CPU request: %v", err)
	}

	memRequestFloat, err := strconv.ParseFloat(memRequest.AsDec().String(), 64)
	if err != nil {
		return WeightedPod{}, fmt.Errorf("failed to parse memory request: %v", err)
	}

	cpuLimitFloat, err := strconv.ParseFloat(cpuLimit.AsDec().String(), 64)
	if err != nil {
		return WeightedPod{}, fmt.Errorf("failed to parse CPU limit: %v", err)
	}

	memLimitFloat, err := strconv.ParseFloat(memLimit.AsDec().String(), 64)
	if err != nil {
		return WeightedPod{}, fmt.Errorf("failed to parse memory limit: %v", err)
	}

	// TODO: Implement logic to determine disk type, network type, storage size and priority
	requestedDiskType := "SSD"
	requestedNetworkType := "isolated"
	requestedStorageSize := 10.0
	priority := LowPriority

	return WeightedPod{
		Selector:             fmt.Sprintf("metadata.name=%s", k8sPod.Name),
		RequestedCPU:         cpuRequestFloat,
		RequestedMemory:      memRequestFloat,
		RequestedStorage:     requestedStorageSize,
		RequestedDiskType:    requestedDiskType,
		RequestedNetworkType: requestedNetworkType,
		LimitCPU:             cpuLimitFloat,
		LimitMemory:          memLimitFloat,
		Priority:             priority,
	}, nil
}
