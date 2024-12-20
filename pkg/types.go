package pkg

import (
	corev1 "k8s.io/api/core/v1"
)

type ComputeType string
type WorkloadPriorityEnum bool

func (p WorkloadPriorityEnum) String() string {
	if p {
		return "PriorityHigh"
	}

	return "PriorityLow"
}

type ComputeConfiguration struct {
	Identifier        *string      `json:"identifier,omitempty"`
	Provider          *string      `json:"provider,omitempty"`
	Location          *string      `json:"location,omitempty"`
	DataCenter        *string      `json:"dataCenter,omitempty"`
	OsType            *string      `json:"osType,omitempty"`
	OsVersion         *string      `json:"osVersion,omitempty"`
	CloudNetworkTypes []string     `json:"cloudNetworkTypes,omitempty"`
	VCpuType          *string      `json:"vCpuType,omitempty"`
	VCpu              *int64       `json:"vCpu,omitempty"`
	RamGb             *int64       `json:"ramGb,omitempty"`
	VolumeGb          *int64       `json:"volumeGb,omitempty"`
	VolumeType        *string      `json:"volumeType,omitempty"`
	Cost              *ComputeCost `json:"cost,omitempty"`
	ComputeType       ComputeType  `json:"computeType,omitempty"`
}

type ComputeCost struct {
	Unit         *string  `json:"unit,omitempty"`
	Currency     *string  `json:"currency,omitempty"`
	PricePerUnit *float64 `json:"pricePerUnit,omitempty"`
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
	Annotations      map[string]string
	Selector         map[string]string
	Weights          map[string]float64
	InterruptionRate WeightedInteruptionRate
	LatencyRate      WeightedLatencyRate
}

type WeightedPod struct {
	Annotations map[string]string
	Selector    map[string]string
	Weights     map[string]float64
}

type WeightedInteruptionRate struct {
	Selector map[string]string
	Weight   float64
}

type WeightedLatencyRate struct {
	Selector map[string]string
	Weight   float64
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
