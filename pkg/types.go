package pkg

import (
	"context"

	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

type ComputeType string
type PriorityEnum bool

func (p PriorityEnum) String() string {
	if p {
		return "PriorityHigh"
	}

	return "PriorityLow"
}

type Config struct {
	RedisServerAddress        string
	RedisServerPassword       string
	RedisServerDatabase       int
	ServerAddress             string
	CertificateOrganization   string
	CertificateCommonName     string
	CertificateDnsNamesCSV    string
	CertificateIpAddressesCSV string
	CertificateExportPath     string
}

type WeightedNode struct {
	Selector         map[string]string
	AvailableCPU     float64
	TotalCPU         float64
	AvailableMemory  float64
	TotalMemory      float64
	AvailableStorage float64
	TotalStorage     float64
	DiskType         string
	NetworkType      string
	Price            float64
	MedianPrice      float64
	InstanceType     string
	InterruptionRate WeightedInteruptionRate
	LatencyRate      WeightedLatencyRate
}

type WeightedPod struct {
	Selector             map[string]string
	RequestedCPU         float64
	RequestedMemory      float64
	RequestedStorage     float64
	RequestedDiskType    string
	RequestedNetworkType string
	LimitCPU             float64
	LimitMemory          float64
	Priority             PriorityEnum
}

type WeightedInteruptionRate struct {
	Selector map[string]string
	Value    float64
}

type WeightedLatencyRate struct {
	Selector map[string]string
	Value    float64
}

type ComputeCost struct {
	Unit         *string  `json:"unit,omitempty"`
	Currency     *string  `json:"currency,omitempty"`
	PricePerUnit *float32 `json:"pricePerUnit,omitempty"`
}

type ComputeConfiguration struct {
	Id                *int32       `json:"id,omitempty"`
	ProviderId        *int32       `json:"providerId,omitempty"`
	ProviderName      *string      `json:"providerName,omitempty"`
	LocationId        *int32       `json:"locationId,omitempty"`
	LocationName      *string      `json:"locationName,omitempty"`
	DataCenterId      *string      `json:"dataCenterId,omitempty"`
	DataCenterName    *string      `json:"dataCenterName,omitempty"`
	OsId              *int32       `json:"osId,omitempty"`
	OsType            *string      `json:"osType,omitempty"`
	OsVersion         *string      `json:"osVersion,omitempty"`
	CloudNetworkTypes []string     `json:"cloudNetworkTypes,omitempty"`
	VCpuType          *string      `json:"vCpuType,omitempty"`
	VCpu              *int32       `json:"vCpu,omitempty"`
	RamGb             *int32       `json:"ramGb,omitempty"`
	VolumeGb          *int32       `json:"volumeGb,omitempty"`
	VolumeType        *string      `json:"volumeType,omitempty"`
	Cost              *ComputeCost `json:"cost,omitempty"`
	ComputeType       ComputeType  `json:"computeType,omitempty"`
}

type KubernetesClientWrapper struct {
	ClientSet *kubernetes.Clientset
}

func (k *KubernetesClientWrapper) CoreV1() v1.CoreV1Interface {
	return k.ClientSet.CoreV1()
}

type MetricsClientWrapper struct {
	ClientSet *metricsclient.Clientset
}

func (m *MetricsClientWrapper) MetricsV1beta1() metricsv1beta1.MetricsV1beta1Interface {
	return m.ClientSet.MetricsV1beta1()
}

type MetricsNodeList struct {
	Items []MetricsNode
}

type MetricsNode struct {
	Name  string
	Usage corev1.ResourceList
}

type MetricsPodList struct {
	Items []MetricsPod
}

type MetricsPod struct {
	Name       string
	Namespace  string
	Containers []MetricsContainer
}

type MetricsContainer struct {
	Name  string
	Usage corev1.ResourceList
}

type RealKubernetesClient struct {
	ClientSet *kubernetes.Clientset
}

func (r *RealKubernetesClient) ListNodes(ctx context.Context, opts metav1.ListOptions) (*corev1.NodeList, error) {
	return r.ClientSet.CoreV1().Nodes().List(ctx, opts)
}

func (r *RealKubernetesClient) ListNamespaces(ctx context.Context, opts metav1.ListOptions) (*corev1.NamespaceList, error) {
	return r.ClientSet.CoreV1().Namespaces().List(ctx, opts)
}

type RealMetricsClient struct {
	ClientSet *metricsclient.Clientset
}

func (r *RealMetricsClient) ListNodeMetrics(ctx context.Context, opts metav1.ListOptions) (*MetricsNodeList, error) {
	nodeMetricsList, err := r.ClientSet.MetricsV1beta1().NodeMetricses().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var metricsNodeList MetricsNodeList
	for _, item := range nodeMetricsList.Items {
		metricsNodeList.Items = append(metricsNodeList.Items, MetricsNode{
			Name:  item.Name,
			Usage: item.Usage,
		})
	}

	return &metricsNodeList, nil
}

func (r *RealMetricsClient) ListPodMetrics(ctx context.Context, namespace string, opts metav1.ListOptions) (*MetricsPodList, error) {
	podMetricsList, err := r.ClientSet.MetricsV1beta1().PodMetricses(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var metricsPodList MetricsPodList
	for _, item := range podMetricsList.Items {
		var containers []MetricsContainer
		for _, c := range item.Containers {
			containers = append(containers, MetricsContainer{
				Name:  c.Name,
				Usage: c.Usage,
			})
		}
		metricsPodList.Items = append(metricsPodList.Items, MetricsPod{
			Name:       item.Name,
			Namespace:  item.Namespace,
			Containers: containers,
		})
	}

	return &metricsPodList, nil
}
