package ultron

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	AnnotationDiskType    = "ultron.io/disk-type"
	AnnotationNetworkType = "ultron.io/network-type"
	AnnotationStorageSize = "ultron.io/storage-size"
	AnnotationPriority    = "ultron.io/priority"
	LabelHostName         = "kubernetes.io/hostname"
	LabelInstanceType     = "node.kubernetes.io/instance-type"
	MetadataName          = "metadata.name"
)

type KubernetesClient interface {
	GetWeightedNodes() ([]WeightedNode, error)
}

type IKubernetesClient struct {
	mapper               Mapper
	computeService       ComputeService
	kubernetesMasterUrl  string
	kubernetesConfigPath string
}

func NewIKubernetesClient(kubernetesMasterUrl string, kubernetesConfigPath string, mapper Mapper, computeService ComputeService) *IKubernetesClient {
	return &IKubernetesClient{
		kubernetesMasterUrl:  kubernetesMasterUrl,
		kubernetesConfigPath: kubernetesConfigPath,
		computeService:       computeService,
		mapper:               mapper,
	}
}

func (kc IKubernetesClient) GetWeightedNodes() ([]WeightedNode, error) {
	var err error

	if kc.kubernetesMasterUrl == "tcp://:" {
		kc.kubernetesMasterUrl = ""
	}

	config, err := clientcmd.BuildConfigFromFlags(kc.kubernetesMasterUrl, kc.kubernetesConfigPath)
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

	var wNodes []WeightedNode
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, node := range nodes.Items {
		wNode, err := kc.mapper.MapNodeToWeightedNode(&node)
		if err != nil {
			return nil, err
		}

		computeConfiguration, err := kc.computeService.MatchWeightedNodeToComputeConfiguration(wNode)
		if err != nil {
			return nil, err
		}

		if computeConfiguration != nil && computeConfiguration.Cost != nil && computeConfiguration.Cost.PricePerUnit != nil {
			wNode.Price = float64(*computeConfiguration.Cost.PricePerUnit)
		}

		medianPrice, err := kc.computeService.CalculateWeightedNodeMedianPrice(wNode)
		if err != nil {
			return nil, err
		}

		wNode.MedianPrice = medianPrice
		// TODO: Uncomment once External API is updated
		// wNode.InterruptionRate = vmConfiguration.InterruptionRate

		wNodes = append(wNodes, wNode)
	}

	return wNodes, nil
}
