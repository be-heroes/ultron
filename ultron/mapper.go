package ultron

import (
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
)

var MapPodToWeightedPod = func(pod *corev1.Pod) (WeightedPod, error) {
	if pod.Name == "" {
		return WeightedPod{}, fmt.Errorf("missing required field: %s", MetadataName)
	}

	var totalCPURequest, totalMemoryRequest, totalCPULimit, totalMemoryLimit float64

	for _, container := range pod.Spec.Containers {
		cpuRequest := container.Resources.Requests[corev1.ResourceCPU]
		memRequest := container.Resources.Requests[corev1.ResourceMemory]
		cpuLimit := container.Resources.Limits[corev1.ResourceCPU]
		memLimit := container.Resources.Limits[corev1.ResourceMemory]

		cpuRequestFloat, err := strconv.ParseFloat(cpuRequest.AsDec().String(), 64)
		if err != nil {
			return WeightedPod{}, fmt.Errorf("failed to parse CPU request: %w", err)
		}

		totalCPURequest += cpuRequestFloat
		memRequestFloat, err := strconv.ParseFloat(memRequest.AsDec().String(), 64)
		if err != nil {
			return WeightedPod{}, fmt.Errorf("failed to parse memory request: %w", err)
		}

		totalMemoryRequest += memRequestFloat
		cpuLimitFloat, err := strconv.ParseFloat(cpuLimit.AsDec().String(), 64)
		if err != nil {
			return WeightedPod{}, fmt.Errorf("failed to parse CPU limit: %w", err)
		}

		totalCPULimit += cpuLimitFloat
		memLimitFloat, err := strconv.ParseFloat(memLimit.AsDec().String(), 64)
		if err != nil {
			return WeightedPod{}, fmt.Errorf("failed to parse memory limit: %w", err)
		}

		totalMemoryLimit += memLimitFloat
	}

	requestedDiskType := getAnnotationOrDefault(pod.Annotations, AnnotationDiskType, DefaultDiskType)
	requestedNetworkType := getAnnotationOrDefault(pod.Annotations, AnnotationNetworkType, DefaultNetworkType)
	requestedStorageSize := getFloatAnnotationOrDefault(pod.Annotations, AnnotationStorageSize, DefaultStorageSizeGB)
	priority := getPriorityFromAnnotation(pod.Annotations)

	return WeightedPod{
		Selector:             map[string]string{MetadataName: pod.Name},
		RequestedCPU:         totalCPURequest,
		RequestedMemory:      totalMemoryRequest,
		RequestedStorage:     requestedStorageSize,
		RequestedDiskType:    requestedDiskType,
		RequestedNetworkType: requestedNetworkType,
		LimitCPU:             totalCPULimit,
		LimitMemory:          totalMemoryLimit,
		Priority:             priority,
	}, nil
}

var MapNodeToWeightedNode = func(node *corev1.Node) (WeightedNode, error) {
	const bytesInGiB = 1024 * 1024 * 1024

	cpuAllocatable := node.Status.Allocatable[corev1.ResourceCPU]
	memAllocatable := node.Status.Allocatable[corev1.ResourceMemory]
	storageAllocatable := node.Status.Allocatable[corev1.ResourceEphemeralStorage]
	cpuCapacity := node.Status.Capacity[corev1.ResourceCPU]
	memCapacity := node.Status.Capacity[corev1.ResourceMemory]
	storageCapacity := node.Status.Capacity[corev1.ResourceEphemeralStorage]
	availableCPU := cpuAllocatable.AsApproximateFloat64()
	availableMemory := float64(memAllocatable.Value()) / float64(bytesInGiB)
	availableStorage := float64(storageAllocatable.Value()) / float64(bytesInGiB)
	totalCPU := cpuCapacity.AsApproximateFloat64()
	totalMemory := float64(memCapacity.Value()) / float64(bytesInGiB)
	totalStorage := float64(storageCapacity.Value()) / float64(bytesInGiB)
	hostname := node.Labels[LabelHostName]

	instanceType, ok := node.Labels[LabelInstanceType]
	if !ok || instanceType == "" {
		return WeightedNode{}, fmt.Errorf("missing required label: %s", LabelInstanceType)
	}

	selector := map[string]string{LabelInstanceType: instanceType}

	if hostname != "" {
		selector[LabelHostName] = hostname
	}

	return WeightedNode{
		Selector:         selector,
		AvailableCPU:     availableCPU,
		TotalCPU:         totalCPU,
		AvailableMemory:  availableMemory,
		TotalMemory:      totalMemory,
		AvailableStorage: availableStorage,
		TotalStorage:     totalStorage,
		DiskType:         node.Annotations[AnnotationDiskType],
		NetworkType:      node.Annotations[AnnotationNetworkType],
		Price:            0,
		MedianPrice:      0,
		InstanceType:     instanceType,
		InterruptionRate: 0,
	}, nil
}
