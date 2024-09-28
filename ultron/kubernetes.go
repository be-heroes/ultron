package ultron

import (
	"context"
	"fmt"

	emma "github.com/emma-community/emma-go-sdk"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetWeightedNodes(kubernetesMasterUrl string, kubernetesConfigPath string) ([]WeightedNode, error) {
	var err error

	config, err := clientcmd.BuildConfigFromFlags(kubernetesMasterUrl, kubernetesConfigPath)
	if err != nil {
		fmt.Println("Falling back to docker Kubernetes API at  https://kubernetes.docker.internal:6443")

		config = &rest.Config{
			Host: "https://kubernetes.docker.internal:6443",
			TLSClientConfig: rest.TLSClientConfig{
				Insecure: true,
			},
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	var weightedNodes []WeightedNode
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		// TODO: Ensure that we have RBAC in place to allow the service account to list nodes. For now just silence error and return empty array
		return weightedNodes, nil
	}

	for _, node := range nodes.Items {
		vmConfiguration, err := matchVmConfiguration(&node)
		if err != nil {
			return nil, err
		}

		cpuAllocatable := node.Status.Allocatable[v1.ResourceCPU]
		memAllocatable := node.Status.Allocatable[v1.ResourceMemory]
		storageAllocatable := node.Status.Allocatable[v1.ResourceEphemeralStorage]
		availableCPU := cpuAllocatable.AsApproximateFloat64()
		availableMemory := (float64)(memAllocatable.Value() / (1024 * 1024 * 1024))
		availableStorage := (float64)(storageAllocatable.Value() / (1024 * 1024 * 1024))
		hostname := node.Labels["kubernetes.io/hostname"]
		instanceType := node.Labels["node.kubernetes.io/instance-type"]
		diskType := node.Annotations[AnnotationDiskType]
		networkType := node.Annotations[AnnotationNetworkType]

		var nodePrice float64
		if vmConfiguration.Cost != nil && vmConfiguration.Cost.PricePerUnit != nil {
			nodePrice = float64(*vmConfiguration.Cost.PricePerUnit)
		} else {
			nodePrice = 0
		}

		// TODO: Calculate median price and interruption rate
		nodeMedianPrice := 0.25
		nodeInteruptionRate := 0.05

		weightedNode := WeightedNode{
			Selector:         map[string]string{"kubernetes.io/hostname": hostname, "node.kubernetes.io/instance-type": instanceType},
			AvailableCPU:     availableCPU,
			TotalCPU:         availableCPU,
			AvailableMemory:  availableMemory,
			TotalMemory:      availableMemory,
			AvailableStorage: availableStorage,
			DiskType:         diskType,
			NetworkType:      networkType,
			Price:            nodePrice,
			MedianPrice:      nodeMedianPrice,
			InstanceType:     instanceType,
			InterruptionRate: nodeInteruptionRate,
		}

		weightedNodes = append(weightedNodes, weightedNode)
	}

	return weightedNodes, nil
}

func matchVmConfiguration(node *v1.Node) (*emma.VmConfiguration, error) {

	return &emma.VmConfiguration{}, nil
}
