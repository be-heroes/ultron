package kubernetes

import (
	"context"

	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RealKubernetesClient struct {
	ClientSet *kubernetes.Clientset
}

func (r *RealKubernetesClient) ListNodes(ctx context.Context, opts metav1.ListOptions) (*corev1.NodeList, error) {
	return r.ClientSet.CoreV1().Nodes().List(ctx, opts)
}

func (r *RealKubernetesClient) ListPods(ctx context.Context, namespace string, opts metav1.ListOptions) (*corev1.PodList, error) {
	return r.ClientSet.CoreV1().Pods(namespace).List(ctx, opts)
}

func (r *RealKubernetesClient) ListNamespaces(ctx context.Context, opts metav1.ListOptions) (*corev1.NamespaceList, error) {
	return r.ClientSet.CoreV1().Namespaces().List(ctx, opts)
}
