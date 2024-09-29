package ultron

// import (
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	corev1 "k8s.io/api/core/v1"
// 	"k8s.io/apimachinery/pkg/api/resource"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// )

// func TestMapPodToWeightedPod_Success(t *testing.T) {
// 	// Arrange
// 	pod := &corev1.Pod{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name: "test-pod",
// 			Annotations: map[string]string{
// 				AnnotationDiskType:    "ssd",
// 				AnnotationNetworkType: "isolated",
// 				AnnotationStorageSize: "100",
// 				AnnotationPriority:    "PriorityHigh",
// 			},
// 		},
// 		Spec: corev1.PodSpec{
// 			Containers: []corev1.Container{
// 				{
// 					Resources: corev1.ResourceRequirements{
// 						Requests: corev1.ResourceList{
// 							corev1.ResourceCPU:    resource.MustParse("100m"),
// 							corev1.ResourceMemory: resource.MustParse("128Mi"),
// 						},
// 						Limits: corev1.ResourceList{
// 							corev1.ResourceCPU:    resource.MustParse("200m"),
// 							corev1.ResourceMemory: resource.MustParse("256Mi"),
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	// Act
// 	weightedPod, err := MapPodToWeightedPod(pod)

// 	// Assert
// 	assert.NoError(t, err)
// 	assert.Equal(t, "test-pod", weightedPod.Selector[MetadataName])
// 	assert.Equal(t, 0.1, weightedPod.RequestedCPU)
// 	assert.Equal(t, float64(134217728), weightedPod.RequestedMemory)
// 	assert.Equal(t, 0.2, weightedPod.LimitCPU)
// 	assert.Equal(t, float64(268435456), weightedPod.LimitMemory)
// 	assert.Equal(t, "ssd", weightedPod.RequestedDiskType)
// 	assert.Equal(t, "isolated", weightedPod.RequestedNetworkType)
// 	assert.Equal(t, 100.0, weightedPod.RequestedStorage)
// 	assert.Equal(t, PriorityHigh, weightedPod.Priority)
// }

// func TestMapPodToWeightedPod_Failure(t *testing.T) {
// 	// Arrange
// 	pod := &corev1.Pod{}

// 	// Act
// 	_, err := MapPodToWeightedPod(pod)

// 	// Assert
// 	assert.Error(t, err)
// }

// func TestMapNodeToWeightedNode_Success(t *testing.T) {
// 	// Arrange
// 	node := &corev1.Node{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name: "test-node",
// 			Labels: map[string]string{
// 				LabelInstanceType: "t2.medium",
// 				LabelHostName:     "node1",
// 			},
// 			Annotations: map[string]string{
// 				AnnotationDiskType:    "ssd",
// 				AnnotationNetworkType: "fast-network",
// 			},
// 		},
// 		Status: corev1.NodeStatus{
// 			Allocatable: corev1.ResourceList{
// 				corev1.ResourceCPU:              resource.MustParse("4"),
// 				corev1.ResourceMemory:           resource.MustParse("8Gi"),
// 				corev1.ResourceEphemeralStorage: resource.MustParse("500Gi"),
// 			},
// 			Capacity: corev1.ResourceList{
// 				corev1.ResourceCPU:              resource.MustParse("8"),
// 				corev1.ResourceMemory:           resource.MustParse("16Gi"),
// 				corev1.ResourceEphemeralStorage: resource.MustParse("1000Gi"),
// 			},
// 		},
// 	}

// 	// Act
// 	weightedNode, err := MapNodeToWeightedNode(node)

// 	// Assert
// 	assert.NoError(t, err)
// 	assert.Equal(t, "t2.medium", weightedNode.InstanceType)
// 	assert.Equal(t, "node1", weightedNode.Selector[LabelHostName])
// 	assert.Equal(t, 4.0, weightedNode.AvailableCPU)
// 	assert.Equal(t, 8.0, weightedNode.TotalCPU)
// 	assert.Equal(t, 8.0, weightedNode.AvailableMemory)
// 	assert.Equal(t, 16.0, weightedNode.TotalMemory)
// 	assert.Equal(t, 500.0, weightedNode.AvailableStorage)
// 	assert.Equal(t, 1000.0, weightedNode.TotalStorage)
// 	assert.Equal(t, "ssd", weightedNode.DiskType)
// 	assert.Equal(t, "fast-network", weightedNode.NetworkType)
// }

// func TestMapNodeToWeightedNode_MissingInstanceType(t *testing.T) {
// 	// Arrange
// 	node := &corev1.Node{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name: "test-node",
// 			Labels: map[string]string{
// 				LabelHostName: "node1",
// 			},
// 		},
// 		Status: corev1.NodeStatus{},
// 	}

// 	//Act
// 	_, err := MapNodeToWeightedNode(node)

// 	// Assert
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "missing required label: node.kubernetes.io/instance-type")
// }
