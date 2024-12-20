package pkg

const (
	AnnotationDiskType         = "ultron.io/disk-type"
	AnnotationInstanceType     = "ultron.io/instance-type"
	AnnotationManaged          = "ultron.io/managed"
	AnnotationNetworkType      = "ultron.io/network-type"
	AnnotationStorageSizeGb    = "ultron.io/storage-size-gb"
	AnnotationWorkloadPriority = "ultron.io/workload-priority"

	BlockTypeCertificate   = "CERTIFICATE"
	BlockTypeRsaPrivateKey = "RSA PRIVATE KEY"

	CacheKeyWeightedNodes                                 = "ULTRON_WEIGHTED_NODES"
	CacheKeyDurableComputeConfigurations                  = "ULTRON_DURABLE_CONFIGURATION"
	CacheKeyDurableComputeConfigurationLatencyRates       = "ULTRON_DURABLE_COMPUTECONFIGURATION_LATENCY_RATES"
	CacheKeyEphemeralComputeConfigurations                = "ULTRON_EPHEMERAL_COMPUTECONFIGURATION"
	CacheKeyEphemeralComputeConfigurationInteruptionRates = "ULTRON_EPHEMERAL_COMPUTECONFIGURATION_INTERUPTION_RATES"

	ComputeTypeDurable   ComputeType = "durable"
	ComputeTypeEphemeral ComputeType = "ephemeral"

	DefaultDiskType              = "SSD"
	DefaultNetworkType           = "isolated"
	DefaultStorageSizeGB         = 10.0
	DefaultDurableInstanceType   = "ultron.durable"
	DefaultEphemeralInstanceType = "ultron.ephemeral"
	DefaultWorkloadPriority      = WorkloadPriorityLow

	EnvServerAddress                 = "ULTRON_SERVER_ADDRESS"
	EnvServerCertificateOrganization = "ULTRON_SERVER_CERTIFICATE_ORGANIZATION"
	EnvServerCertificateCommonName   = "ULTRON_SERVER_CERTIFICATE_COMMON_NAME"
	EnvServerCertificateDnsNames     = "ULTRON_SERVER_CERTIFICATE_DNS_NAMES"
	EnvServerCertificateIpAddresses  = "ULTRON_SERVER_CERTIFICATE_IP_ADDRESSES"
	EnvServerCertificateExportPath   = "ULTRON_SERVER_CERTIFICATE_EXPORT_PATH"
	EnvRedisServerAddress            = "ULTRON_SERVER_REDIS_ADDRESS"
	EnvRedisServerPassword           = "ULTRON_SERVER_REDIS_PASSWORD"
	EnvRedisServerDatabase           = "ULTRON_SERVER_REDIS_DATABASE"
	EnvKubernetesConfig              = "KUBECONFIG"
	EnvKubernetesServiceHost         = "KUBERNETES_SERVICE_HOST"
	EnvKubernetesServicePort         = "KUBERNETES_SERVICE_PORT"

	LabelHostName     = "kubernetes.io/hostname"
	LabelInstanceType = "node.kubernetes.io/instance-type"

	MetadataName = "metadata.name"

	TopicNodeObserve = "ULTRON_TOPIC_NODE_OBSERVE"
	TopicPodObserve  = "ULTRON_TOPIC_POD_OBSERVE"

	WeightKeyCpuAvailable     = "cpu_available"
	WeightKeyCpuLimit         = "cpu_limit"
	WeightKeyCpuRequested     = "cpu_requested"
	WeightKeyCpuTotal         = "cpu_total"
	WeightKeyCpuUsage         = "cpu_usage"
	WeightKeyInterruptionRate = "interruption_rate"
	WeightKeyLatencyRate      = "latency_rate"
	WeightKeyMemoryAvailable  = "memory_available"
	WeightKeyMemoryLimit      = "memory_limit"
	WeightKeyMemoryRequested  = "memory_requested"
	WeightKeyMemoryTotal      = "memory_total"
	WeightKeyMemoryUsage      = "memory_usage"
	WeightKeyStorageAvailable = "storage_available"
	WeightKeyStorageRequested = "storage_requested"
	WeightKeyStorageTotal     = "storage_total"
	WeightKeyStorageUsage     = "storage_usage"
	WeightKeyPrice            = "price"
	WeightKeyPriceMedian      = "price_median"

	WorkloadPriorityLow       WorkloadPriorityEnum = false
	WorkloadPriorityHigh      WorkloadPriorityEnum = true
	WorkloadPriorityHighLabel                      = "PriorityHigh"
	WorkloadPriorityLowLabel                       = "PriorityLow"
)
