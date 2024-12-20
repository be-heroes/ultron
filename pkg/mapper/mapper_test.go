package mapper_test

import (
	"fmt"
	"testing"

	ultron "github.com/be-heroes/ultron/pkg"
	mapper "github.com/be-heroes/ultron/pkg/mapper"
	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMapPodToWeightedPod_Success(t *testing.T) {
	mapper := mapper.NewMapper()

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			Annotations: map[string]string{
				ultron.AnnotationDiskType:    "HDD",
				ultron.AnnotationNetworkType: "4G",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("2Gi"),
						},
					},
				},
			},
		},
	}

	// Act
	weightedPod, err := mapper.MapPodToWeightedPod(pod)

	// Assert
	assert.NoError(t, err, "MapPodToWeightedPod should not return an error")
	assert.Equal(t, 0.5, weightedPod.Weights[ultron.WeightKeyCpuRequested], "Expected RequestedCPU to be 0.5")
	assert.Equal(t, float64(1*1024*1024*1024), weightedPod.Weights[ultron.WeightKeyMemoryRequested], "Expected RequestedMemory to be 1Gi")
	assert.Equal(t, "HDD", weightedPod.Annotations[ultron.AnnotationDiskType], "Expected DiskType to be HDD")
	assert.Equal(t, "4G", weightedPod.Annotations[ultron.AnnotationNetworkType], "Expected NetworkType to be 4G")
}

func TestMapPodToWeightedPod_MissingName(t *testing.T) {
	mapper := mapper.NewMapper()

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
		},
	}

	// Act
	_, err := mapper.MapPodToWeightedPod(pod)

	// Assert
	assert.Error(t, err, "Expected error for missing pod name")
	expectedErr := fmt.Sprintf("missing required field: %s", ultron.MetadataName)
	assert.EqualError(t, err, expectedErr, "Expected error message does not match")
}

func TestMapNodeToWeightedNode_Success(t *testing.T) {
	mapper := mapper.NewMapper()

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				ultron.LabelInstanceType: "m5.large",
				ultron.LabelHostName:     "node1",
			},
			Annotations: map[string]string{
				ultron.AnnotationDiskType:    "SSD",
				ultron.AnnotationNetworkType: "5G",
			},
		},
		Status: corev1.NodeStatus{
			Allocatable: corev1.ResourceList{
				corev1.ResourceCPU:              resource.MustParse("2000m"),
				corev1.ResourceMemory:           resource.MustParse("8Gi"),
				corev1.ResourceEphemeralStorage: resource.MustParse("100Gi"),
			},
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:              resource.MustParse("4000m"),
				corev1.ResourceMemory:           resource.MustParse("16Gi"),
				corev1.ResourceEphemeralStorage: resource.MustParse("200Gi"),
			},
		},
	}

	// Act
	weightedNode, err := mapper.MapNodeToWeightedNode(node)

	// Assert
	assert.NoError(t, err, "MapNodeToWeightedNode should not return an error")
	assert.Equal(t, 2.0, weightedNode.Weights[ultron.WeightKeyCpuAvailable], "Expected AvailableCPU to be 2")
	assert.Equal(t, 4.0, weightedNode.Weights[ultron.WeightKeyCpuTotal], "Expected TotalCPU to be 4")
	assert.Equal(t, float64(8), weightedNode.Weights[ultron.WeightKeyMemoryAvailable], "Expected AvailableMemory to be 8Gi")
	assert.Equal(t, float64(16), weightedNode.Weights[ultron.WeightKeyMemoryTotal], "Expected TotalMemory to be 16Gi")
	assert.Equal(t, "SSD", weightedNode.Annotations[ultron.AnnotationDiskType], "Expected DiskType to be SSD")
	assert.Equal(t, "5G", weightedNode.Annotations[ultron.AnnotationNetworkType], "Expected NetworkType to be 5G")
}

func TestMapNodeToWeightedNode_MissingInstanceType(t *testing.T) {
	mapper := mapper.NewMapper()

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{},
		},
		Status: corev1.NodeStatus{},
	}

	// Act
	_, err := mapper.MapNodeToWeightedNode(node)

	// Assert
	assert.Error(t, err, "Expected error for missing instance type")
	expectedErr := fmt.Sprintf("missing required label: %s or %s", ultron.LabelHostName, ultron.LabelInstanceType)
	assert.EqualError(t, err, expectedErr, "Expected error message does not match")
}

func TestGetAnnotationOrDefault(t *testing.T) {
	mapper := mapper.NewMapper()

	tests := []struct {
		annotations   map[string]string
		key           string
		defaultValue  string
		expectedValue string
	}{
		{map[string]string{"key1": "value1"}, "key1", "default", "value1"},
		{map[string]string{"key1": "value1"}, "key2", "default", "default"},
	}

	for _, test := range tests {
		result := mapper.GetAnnotationOrDefault(test.annotations, test.key, test.defaultValue)
		assert.Equal(t, test.expectedValue, result, fmt.Sprintf("Expected %v, got %v", test.expectedValue, result))
	}
}

func TestGetFloatAnnotationOrDefault(t *testing.T) {
	mapper := mapper.NewMapper()

	tests := []struct {
		annotations   map[string]string
		key           string
		defaultValue  float64
		expectedValue float64
	}{
		{map[string]string{"key1": "3.14"}, "key1", 1.23, 3.14},
		{map[string]string{"key1": "invalid"}, "key1", 1.23, 1.23},
		{map[string]string{}, "key2", 1.23, 1.23},
	}

	for _, test := range tests {
		result := mapper.GetFloatAnnotationOrDefault(test.annotations, test.key, test.defaultValue)
		assert.Equal(t, test.expectedValue, result, fmt.Sprintf("Expected %v, got %v", test.expectedValue, result))
	}
}

func TestGetPriorityFromAnnotation(t *testing.T) {
	mapper := mapper.NewMapper()

	tests := []struct {
		annotations   map[string]string
		expectedValue ultron.WorkloadPriorityEnum
	}{
		{map[string]string{ultron.AnnotationWorkloadPriority: ultron.WorkloadPriorityHighLabel}, ultron.WorkloadPriorityHigh},
		{map[string]string{ultron.AnnotationWorkloadPriority: ultron.WorkloadPriorityLowLabel}, ultron.WorkloadPriorityLow},
	}

	for _, test := range tests {
		result := mapper.GetPriorityFromAnnotation(test.annotations)
		assert.Equal(t, test.expectedValue, result, fmt.Sprintf("Expected %v, got %v", test.expectedValue, result))
	}
}
