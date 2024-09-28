package ultron

import (
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
)

func MapPodToWeightedPod(pod *corev1.Pod) (WeightedPod, error) {
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
		Selector:             map[string]string{MetaDataName: pod.Name},
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

func MapNodeToWeightedNode(node *corev1.Node) (WeightedNode, error) {
	const bytesInGiB = 1024 * 1024 * 1024

	cpuAllocatable := node.Status.Allocatable[v1.ResourceCPU]
	memAllocatable := node.Status.Allocatable[v1.ResourceMemory]
	storageAllocatable := node.Status.Allocatable[v1.ResourceEphemeralStorage]

	availableCPU := cpuAllocatable.AsApproximateFloat64()
	availableMemory := float64(memAllocatable.Value()) / float64(bytesInGiB)
	availableStorage := float64(storageAllocatable.Value()) / float64(bytesInGiB)

	hostname, ok := node.Labels[LabelHostName]
	if !ok || hostname == "" {
		return WeightedNode{}, fmt.Errorf("missing required label: %s", LabelHostName)
	}

	instanceType, ok := node.Labels[LabelInstanceType]
	if !ok || instanceType == "" {
		return WeightedNode{}, fmt.Errorf("missing required label: %s", LabelInstanceType)
	}

	diskType := node.Annotations[AnnotationDiskType]
	networkType := node.Annotations[AnnotationNetworkType]

	weightedNode := WeightedNode{
		Selector:         map[string]string{LabelHostName: hostname, LabelInstanceType: instanceType},
		AvailableCPU:     availableCPU,
		TotalCPU:         availableCPU,
		AvailableMemory:  availableMemory,
		TotalMemory:      availableMemory,
		AvailableStorage: availableStorage,
		DiskType:         diskType,
		NetworkType:      networkType,
		Price:            0,
		MedianPrice:      0,
		InstanceType:     instanceType,
		InterruptionRate: 0,
	}

	return weightedNode, nil
}
