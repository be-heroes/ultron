package mapper_test

import (
	"fmt"
	"testing"

	ultron "github.com/be-heroes/ultron/pkg"
	mapper "github.com/be-heroes/ultron/pkg/mapper"

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
				ultron.AnnotationStorageSize: "20",
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

	weightedPod, err := mapper.MapPodToWeightedPod(pod)
	if err != nil {
		t.Fatalf("MapPodToWeightedPod returned an error: %v", err)
	}

	if weightedPod.RequestedCPU != 0.5 {
		t.Errorf("Expected RequestedCPU to be 0.5, got %f", weightedPod.RequestedCPU)
	}
	if weightedPod.RequestedMemory != 1*1024*1024*1024 {
		t.Errorf("Expected RequestedMemory to be 1Gi, got %f", weightedPod.RequestedMemory)
	}
	if weightedPod.RequestedDiskType != "HDD" {
		t.Errorf("Expected RequestedDiskType to be HDD, got %s", weightedPod.RequestedDiskType)
	}
	if weightedPod.RequestedNetworkType != "4G" {
		t.Errorf("Expected RequestedNetworkType to be 4G, got %s", weightedPod.RequestedNetworkType)
	}
	if weightedPod.RequestedStorage != 20 {
		t.Errorf("Expected RequestedStorage to be 20GB, got %f", weightedPod.RequestedStorage)
	}
}

func TestMapPodToWeightedPod_MissingName(t *testing.T) {
	mapper := mapper.NewMapper()

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
		},
	}

	_, err := mapper.MapPodToWeightedPod(pod)
	if err == nil {
		t.Fatalf("Expected error for missing pod name, but got none")
	}

	expectedErr := fmt.Sprintf("missing required field: %s", ultron.MetadataName)
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', but got '%v'", expectedErr, err)
	}
}

func TestMapNodeToWeightedNode_Success(t *testing.T) {
	// Arrange
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
	if err != nil {
		t.Fatalf("MapNodeToWeightedNode returned an error: %v", err)
	}

	if weightedNode.AvailableCPU != 2 {
		t.Errorf("Expected AvailableCPU to be 2, got %f", weightedNode.AvailableCPU)
	}
	if weightedNode.TotalCPU != 4 {
		t.Errorf("Expected TotalCPU to be 4, got %f", weightedNode.TotalCPU)
	}
	if weightedNode.AvailableMemory != 8 {
		t.Errorf("Expected AvailableMemory to be 8Gi, got %f", weightedNode.AvailableMemory)
	}
	if weightedNode.TotalMemory != 16 {
		t.Errorf("Expected TotalMemory to be 16Gi, got %f", weightedNode.TotalMemory)
	}
	if weightedNode.DiskType != "SSD" {
		t.Errorf("Expected DiskType to be SSD, got %s", weightedNode.DiskType)
	}
	if weightedNode.NetworkType != "5G" {
		t.Errorf("Expected NetworkType to be 5G, got %s", weightedNode.NetworkType)
	}
}

func TestMapNodeToWeightedNode_MissingInstanceType(t *testing.T) {
	// Arrange
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
	if err == nil {
		t.Fatalf("Expected error for missing instance type, but got none")
	}

	expectedErr := fmt.Sprintf("missing required label: %s or %s", ultron.LabelHostName, ultron.LabelInstanceType)
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', but got '%v'", expectedErr, err)
	}
}

func TestGetAnnotationOrDefault(t *testing.T) {
	// Arrange
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

	// Act & Assert
	for _, test := range tests {
		result := mapper.GetAnnotationOrDefault(test.annotations, test.key, test.defaultValue)
		if result != test.expectedValue {
			t.Errorf("expected %v, got %v", test.expectedValue, result)
		}
	}
}

func TestGetFloatAnnotationOrDefault(t *testing.T) {
	// Arrange
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

	// Act & Assert
	for _, test := range tests {
		result := mapper.GetFloatAnnotationOrDefault(test.annotations, test.key, test.defaultValue)
		if result != test.expectedValue {
			t.Errorf("expected %v, got %v", test.expectedValue, result)
		}
	}
}

func TestGetPriorityFromAnnotation(t *testing.T) {
	// Arrange
	mapper := mapper.NewMapper()
	tests := []struct {
		annotations   map[string]string
		expectedValue ultron.PriorityEnum
	}{
		{map[string]string{ultron.AnnotationPriority: "PriorityHigh"}, ultron.PriorityHigh},
		{map[string]string{ultron.AnnotationPriority: "PriorityLow"}, ultron.PriorityLow},
	}

	// Act & Assert
	for _, test := range tests {
		result := mapper.GetPriorityFromAnnotation(test.annotations)
		if result != test.expectedValue {
			t.Errorf("expected %v, got %v", test.expectedValue, result)
		}
	}
}
