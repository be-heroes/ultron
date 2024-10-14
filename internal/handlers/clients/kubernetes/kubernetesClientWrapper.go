package kubernetes

import (
	"k8s.io/client-go/kubernetes"

	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type KubernetesClientWrapper struct {
	ClientSet *kubernetes.Clientset
}

func (k *KubernetesClientWrapper) CoreV1() v1.CoreV1Interface {
	return k.ClientSet.CoreV1()
}
