package services

import (
	"context"
	"fmt"
	"strconv"

	k8s "github.com/be-heroes/ultron/internal/clients/kubernetes"
	ultron "github.com/be-heroes/ultron/pkg"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

type ICoreClient interface {
	ListPods(ctx context.Context, namespace string, opts metav1.ListOptions) (*corev1.PodList, error)
	ListNodes(ctx context.Context, opts metav1.ListOptions) (*corev1.NodeList, error)
	ListNamespaces(ctx context.Context, opts metav1.ListOptions) (*corev1.NamespaceList, error)
}

type IMetricsClient interface {
	ListNodeMetrics(ctx context.Context, opts metav1.ListOptions) (*ultron.MetricsNodeList, error)
	ListPodMetrics(ctx context.Context, namespace string, opts metav1.ListOptions) (*ultron.MetricsPodList, error)
}

type IKubernetesService interface {
	GetPods(ctx context.Context, options metav1.ListOptions) ([]corev1.Pod, error)
	GetNodes(ctx context.Context, options metav1.ListOptions) ([]corev1.Node, error)
	GetNodeMetrics(ctx context.Context, options metav1.ListOptions) (map[string]map[string]string, error)
	GetPodMetrics(ctx context.Context, options metav1.ListOptions) (map[string]map[string]string, error)
}

type KubernetesService struct {
	config        *rest.Config
	K8sClient     ICoreClient
	MetricsClient IMetricsClient
}

func NewKubernetesService(kubernetesMasterUrl string, kubernetesConfigPath string, insecure bool) (*KubernetesService, error) {
	if kubernetesMasterUrl == "https://:" {
		kubernetesMasterUrl = ""
	}

	config, err := clientcmd.BuildConfigFromFlags(kubernetesMasterUrl, kubernetesConfigPath)
	if err != nil {
		return nil, err
	}

	config.Insecure = insecure

	k8sClientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	metricsClientset, err := metricsclient.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &KubernetesService{
		config:        config,
		K8sClient:     &k8s.RealKubernetesClient{ClientSet: k8sClientset},
		MetricsClient: &k8s.RealMetricsClient{ClientSet: metricsClientset},
	}, nil
}

func (ks *KubernetesService) GetNodes(ctx context.Context, options metav1.ListOptions) ([]corev1.Node, error) {
	nodesList, err := ks.K8sClient.ListNodes(ctx, options)
	if err != nil {
		return nil, err
	}

	return nodesList.Items, nil
}

func (ks *KubernetesService) GetPods(ctx context.Context, options metav1.ListOptions) ([]corev1.Pod, error) {
	namespacesList, err := ks.K8sClient.ListNamespaces(ctx, options)
	if err != nil {
		return nil, err
	}

	var pods []corev1.Pod

	for _, namespace := range namespacesList.Items {
		podList, err := ks.K8sClient.ListPods(ctx, namespace.Name, options)
		if err != nil {
			return nil, err
		}

		pods = append(pods, podList.Items...)
	}

	return pods, nil
}

func (ks *KubernetesService) GetNodeMetrics(ctx context.Context, options metav1.ListOptions) (map[string]map[string]string, error) {
	metricsNodeList, err := ks.MetricsClient.ListNodeMetrics(ctx, options)
	if err != nil {
		return nil, err
	}

	metrics := make(map[string]map[string]string)

	for _, nodeMetric := range metricsNodeList.Items {
		cpuUsage := nodeMetric.Usage["cpu"]
		memoryUsage := nodeMetric.Usage["memory"]

		metrics[nodeMetric.Name] = map[string]string{
			ultron.WeightKeyCpuUsage:    cpuUsage.AsDec().String(),
			ultron.WeightKeyMemoryUsage: memoryUsage.AsDec().String(),
		}
	}

	return metrics, nil
}

func (ks *KubernetesService) GetPodMetrics(ctx context.Context, options metav1.ListOptions) (map[string]map[string]string, error) {
	namespacesList, err := ks.K8sClient.ListNamespaces(ctx, options)
	if err != nil {
		return nil, err
	}

	metrics := make(map[string]map[string]string)

	for _, namespace := range namespacesList.Items {
		podMetricsList, err := ks.MetricsClient.ListPodMetrics(ctx, namespace.Name, options)
		if err != nil {
			return nil, err
		}

		for _, podMetric := range podMetricsList.Items {
			cpuTotal := int64(0)
			memoryTotal := int64(0)

			for _, container := range podMetric.Containers {
				cpuUsage := container.Usage[corev1.ResourceCPU]
				memUsage := container.Usage[corev1.ResourceMemory]

				cpuTotal += cpuUsage.MilliValue()
				memoryTotal += memUsage.Value()
			}

			metricsKey := fmt.Sprintf("%s/%s", podMetric.Namespace, podMetric.Name)
			metrics[metricsKey] = map[string]string{
				ultron.WeightKeyCpuTotal:    strconv.FormatInt(cpuTotal, 10),
				ultron.WeightKeyMemoryTotal: strconv.FormatInt(memoryTotal, 10),
			}
		}
	}

	return metrics, nil
}
