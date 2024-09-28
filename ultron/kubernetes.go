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
	MetaDataName          = "metadata.name"
)

func GetWeightedNodes(kubernetesMasterUrl string, kubernetesConfigPath string) ([]WeightedNode, error) {
	var err error

	if kubernetesMasterUrl == "tcp://:" {
		kubernetesMasterUrl = ""
	}

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

	var wNodes []WeightedNode
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, node := range nodes.Items {
		wNode, err := MapNodeToWeightedNode(&node)
		if err != nil {
			return nil, err
		}

		vmConfiguration, err := MatchWeightedNodeToVmConfiguration(wNode)
		if err != nil {
			return nil, err
		}

		if vmConfiguration.Cost != nil && vmConfiguration.Cost.PricePerUnit != nil {
			wNode.Price = float64(*vmConfiguration.Cost.PricePerUnit)
		}

		// TODO: Figure out strategy for calculating median prices and interruption rates
		wNode.MedianPrice = 0.25
		wNode.InterruptionRate = 0.05

		wNodes = append(wNodes, wNode)
	}

	return wNodes, nil
}
