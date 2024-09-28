package ultron

import (
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
)

func MapPodToWeightedPod(k8sPod *corev1.Pod) (WeightedPod, error) {
	var totalCPURequest, totalMemoryRequest, totalCPULimit, totalMemoryLimit float64

	for _, container := range k8sPod.Spec.Containers {
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

	requestedDiskType := getAnnotationOrDefault(k8sPod.Annotations, AnnotationDiskType, DefaultDiskType)
	requestedNetworkType := getAnnotationOrDefault(k8sPod.Annotations, AnnotationNetworkType, DefaultNetworkType)
	requestedStorageSize := getFloatAnnotationOrDefault(k8sPod.Annotations, AnnotationStorageSize, DefaultStorageSizeGB)
	priority := getPriorityFromAnnotation(k8sPod.Annotations)

	return WeightedPod{
		Selector:             map[string]string{"metadata.name": k8sPod.Name},
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
