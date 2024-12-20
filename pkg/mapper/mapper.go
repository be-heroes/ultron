package mapper

import (
	"fmt"
	"strconv"

	ultron "github.com/be-heroes/ultron/pkg"

	corev1 "k8s.io/api/core/v1"
)

type IMapper interface {
	MapPodToWeightedPod(pod *corev1.Pod) (ultron.WeightedPod, error)
	MapNodeToWeightedNode(node *corev1.Node) (ultron.WeightedNode, error)
}

type Mapper struct{}

func NewMapper() *Mapper {
	return &Mapper{}
}

func (m *Mapper) MapPodToWeightedPod(pod *corev1.Pod) (ultron.WeightedPod, error) {
	if pod.Name == "" {
		return ultron.WeightedPod{}, fmt.Errorf("missing required field: %s", ultron.MetadataName)
	}

	var totalCPURequest, totalMemoryRequest, totalCPULimit, totalMemoryLimit float64

	for _, container := range pod.Spec.Containers {
		if container.Resources.Requests != nil {
			cpuRequest := container.Resources.Requests[corev1.ResourceCPU]
			cpuRequestFloat, err := strconv.ParseFloat(cpuRequest.AsDec().String(), 64)
			if err != nil {
				return ultron.WeightedPod{}, fmt.Errorf("failed to parse CPU request: %w", err)
			}

			totalCPURequest += cpuRequestFloat

			memRequest := container.Resources.Requests[corev1.ResourceMemory]
			memRequestFloat, err := strconv.ParseFloat(memRequest.AsDec().String(), 64)
			if err != nil {
				return ultron.WeightedPod{}, fmt.Errorf("failed to parse memory request: %w", err)
			}

			totalMemoryRequest += memRequestFloat
		}

		if container.Resources.Limits != nil {
			cpuLimit := container.Resources.Limits[corev1.ResourceCPU]
			cpuLimitFloat, err := strconv.ParseFloat(cpuLimit.AsDec().String(), 64)
			if err != nil {
				return ultron.WeightedPod{}, fmt.Errorf("failed to parse CPU limit: %w", err)
			}

			totalCPULimit += cpuLimitFloat

			memLimit := container.Resources.Limits[corev1.ResourceMemory]
			memLimitFloat, err := strconv.ParseFloat(memLimit.AsDec().String(), 64)
			if err != nil {
				return ultron.WeightedPod{}, fmt.Errorf("failed to parse memory limit: %w", err)
			}

			totalMemoryLimit += memLimitFloat
		}
	}

	requestedDiskType := m.GetAnnotationOrDefault(pod.Annotations, ultron.AnnotationDiskType, ultron.DefaultDiskType)
	requestedNetworkType := m.GetAnnotationOrDefault(pod.Annotations, ultron.AnnotationNetworkType, ultron.DefaultNetworkType)
	requestedStorageSize := m.GetFloatAnnotationOrDefault(pod.Annotations, ultron.AnnotationStorageSizeGb, ultron.DefaultStorageSizeGB)
	priority := m.GetPriorityFromAnnotation(pod.Annotations)

	return ultron.WeightedPod{
		Selector: map[string]string{ultron.MetadataName: pod.Name},
		Annotations: map[string]string{
			ultron.AnnotationDiskType:         requestedDiskType,
			ultron.AnnotationNetworkType:      requestedNetworkType,
			ultron.AnnotationWorkloadPriority: priority.String(),
			ultron.AnnotationStorageSizeGb:    strconv.FormatFloat(requestedStorageSize, 'f', -1, 64),
		},
		Weights: map[string]float64{
			ultron.WeightKeyCpuRequested:     totalCPURequest,
			ultron.WeightKeyCpuLimit:         totalCPULimit,
			ultron.WeightKeyMemoryRequested:  totalMemoryRequest,
			ultron.WeightKeyMemoryLimit:      totalMemoryLimit,
			ultron.WeightKeyStorageRequested: requestedStorageSize,
		},
	}, nil
}

func (m *Mapper) MapNodeToWeightedNode(node *corev1.Node) (ultron.WeightedNode, error) {
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
	hostname := node.Labels[ultron.LabelHostName]
	instanceType := node.Labels[ultron.LabelInstanceType]
	managed := node.Annotations[ultron.AnnotationManaged]

	if hostname == "" && instanceType == "" {
		return ultron.WeightedNode{}, fmt.Errorf("missing required label: %s or %s", ultron.LabelHostName, ultron.LabelInstanceType)
	}

	selector := map[string]string{}

	if instanceType != "" {
		selector[ultron.LabelInstanceType] = instanceType
	}

	if hostname != "" {
		selector[ultron.LabelHostName] = hostname
	}

	if managed != "" {
		selector[ultron.AnnotationManaged] = managed
	}

	return ultron.WeightedNode{
		Selector: selector,
		Annotations: map[string]string{
			ultron.AnnotationDiskType:     node.Annotations[ultron.AnnotationDiskType],
			ultron.AnnotationNetworkType:  node.Annotations[ultron.AnnotationNetworkType],
			ultron.AnnotationInstanceType: instanceType,
		},
		Weights: map[string]float64{
			ultron.WeightKeyCpuAvailable:     availableCPU,
			ultron.WeightKeyCpuTotal:         totalCPU,
			ultron.WeightKeyMemoryAvailable:  availableMemory,
			ultron.WeightKeyMemoryTotal:      totalMemory,
			ultron.WeightKeyStorageAvailable: availableStorage,
			ultron.WeightKeyStorageTotal:     totalStorage,
			ultron.WeightKeyPrice:            0,
			ultron.WeightKeyPriceMedian:      0,
		},
		InterruptionRate: ultron.WeightedInteruptionRate{Weight: 0},
	}, nil
}

func (m *Mapper) GetAnnotationOrDefault(annotations map[string]string, key, defaultValue string) string {
	if value, exists := annotations[key]; exists {
		return value
	}

	return defaultValue
}

func (m *Mapper) GetFloatAnnotationOrDefault(annotations map[string]string, key string, defaultValue float64) float64 {
	if valueStr, exists := annotations[key]; exists {
		if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
			return value
		}
	}

	return defaultValue
}

func (m *Mapper) GetPriorityFromAnnotation(annotations map[string]string) ultron.WorkloadPriorityEnum {
	if value, exists := annotations[ultron.AnnotationWorkloadPriority]; exists {
		switch value {
		case "PriorityHigh":
			return ultron.WorkloadPriorityHigh
		case "PriorityLow":
			return ultron.WorkloadPriorityLow
		}
	}

	return ultron.DefaultWorkloadPriority
}
