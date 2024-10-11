package services_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/be-heroes/ultron/mocks"
	ultron "github.com/be-heroes/ultron/pkg"
	services "github.com/be-heroes/ultron/pkg/services"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetNodes(t *testing.T) {
	mockK8sClient := new(mocks.ICoreClient)
	mockK8sClient.On("ListNodes", mock.Anything, metav1.ListOptions{}).Return(&corev1.NodeList{
		Items: []corev1.Node{
			{ObjectMeta: metav1.ObjectMeta{Name: "node1"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "node2"}},
		},
	}, nil)

	mockMetricsClient := new(mocks.IMetricsClient)

	service := &services.KubernetesService{
		K8sClient:     mockK8sClient,
		MetricsClient: mockMetricsClient,
	}

	nodes, err := service.GetNodes()

	assert.NoError(t, err)
	assert.Equal(t, 2, len(nodes))
	assert.Equal(t, "node1", nodes[0].Name)
	assert.Equal(t, "node2", nodes[1].Name)

	mockK8sClient.AssertExpectations(t)
}

func TestGetNodeMetrics(t *testing.T) {
	mockK8sClient := new(mocks.ICoreClient)
	mockMetricsClient := new(mocks.IMetricsClient)

	mockMetricsClient.On("ListNodeMetrics", mock.Anything, metav1.ListOptions{}).Return(&ultron.MetricsNodeList{
		Items: []ultron.MetricsNode{
			{
				Name: "node1",
				Usage: map[corev1.ResourceName]resource.Quantity{
					"cpu":    *resource.NewMilliQuantity(500000, resource.DecimalSI),
					"memory": *resource.NewQuantity(1024*1024*1024, resource.BinarySI),
				},
			},
			{
				Name: "node2",
				Usage: map[corev1.ResourceName]resource.Quantity{
					"cpu":    *resource.NewMilliQuantity(300000, resource.DecimalSI),
					"memory": *resource.NewQuantity(512*1024*1024, resource.BinarySI),
				},
			},
		},
	}, nil)

	service := &services.KubernetesService{
		K8sClient:     mockK8sClient,
		MetricsClient: mockMetricsClient,
	}

	nodeMetrics, err := service.GetNodeMetrics()

	assert.NoError(t, err)
	assert.Equal(t, 2, len(nodeMetrics))
	assert.Equal(t, "500.000", nodeMetrics["node1"]["cpuUsage"])
	assert.Equal(t, "1073741824", nodeMetrics["node1"]["memoryUsage"])
	assert.Equal(t, "300.000", nodeMetrics["node2"]["cpuUsage"])
	assert.Equal(t, "536870912", nodeMetrics["node2"]["memoryUsage"])

	mockMetricsClient.AssertExpectations(t)
}

func TestGetPodMetrics(t *testing.T) {
	mockK8sClient := new(mocks.ICoreClient)
	mockK8sClient.On("ListNamespaces", mock.Anything, metav1.ListOptions{}).Return(&corev1.NamespaceList{
		Items: []corev1.Namespace{
			{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
		},
	}, nil)

	mockMetricsClient := new(mocks.IMetricsClient)
	mockMetricsClient.On("ListPodMetrics", mock.Anything, "default", metav1.ListOptions{}).Return(&ultron.MetricsPodList{
		Items: []ultron.MetricsPod{
			{
				Name:      "pod1",
				Namespace: "default",
				Containers: []ultron.MetricsContainer{
					{
						Name: "container1",
						Usage: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
					{
						Name: "container2",
						Usage: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("256Mi"),
						},
					},
				},
			},
		},
	}, nil)

	service := &services.KubernetesService{
		K8sClient:     mockK8sClient,
		MetricsClient: mockMetricsClient,
	}

	metrics, err := service.GetPodMetrics()

	assert.NoError(t, err)
	assert.Contains(t, metrics, "default/pod1")
	assert.Equal(t, "300", metrics["default/pod1"]["cpuTotal"])
	assert.Equal(t, "402653184", metrics["default/pod1"]["memoryTotal"])

	mockK8sClient.AssertExpectations(t)
	mockMetricsClient.AssertExpectations(t)
}
