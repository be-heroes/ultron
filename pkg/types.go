package pkg

type ComputeType string
type PriorityEnum bool

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
