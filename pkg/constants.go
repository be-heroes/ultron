package pkg

const (
	AnnotationDiskType    = "ultron.io/disk-type"
	AnnotationNetworkType = "ultron.io/network-type"
	AnnotationStorageSize = "ultron.io/storage-size"
	AnnotationPriority    = "ultron.io/priority"
	AnnotationManaged     = "ultron.io/managed"

	CacheKeyWeightedNodes                       = "ULTRON_WEIGHTED_NODES"
	CacheKeyDurableVmConfigurations             = "ULTRON_DURABLE_VMCONFIGURATION"
	CacheKeyDurableVmConfigurationLatencyRates  = "ULTRON_DURABLE_VMCONFIGURATION_LATENCY_RATES"
	CacheKeySpotVmConfigurations                = "ULTRON_SPOT_VMCONFIGURATION"
	CacheKeySpotVmConfigurationInteruptionRates = "ULTRON_SPOT_VMCONFIGURATION_INTERUPTION_RATES"

	ComputeTypeDurable   ComputeType = "durable"
	ComputeTypeEphemeral ComputeType = "ephemeral"

	DefaultDiskType              = "SSD"
	DefaultNetworkType           = "isolated"
	DefaultStorageSizeGB         = 10.0
	DefaultPriority              = PriorityLow
	DefaultDurableInstanceType   = "ultron.durable"
	DefaultEphemeralInstanceType = "ultron.ephemeral"

	EnvServerAddress                 = "ULTRON_SERVER_ADDRESS"
	EnvServerCertificateOrganization = "ULTRON_SERVER_CERTIFICATE_ORGANIZATION"
	EnvServerCertificateCommonName   = "ULTRON_SERVER_CERTIFICATE_COMMON_NAME"
	EnvServerCertificateDnsNames     = "ULTRON_SERVER_CERTIFICATE_DNS_NAMES"
	EnvServerCertificateIpAddresses  = "ULTRON_SERVER_CERTIFICATE_IP_ADDRESSES"
	EnvServerCertificateExportPath   = "ULTRON_SERVER_CERTIFICATE_EXPORT_PATH"
	EnvRedisServerAddress            = "ULTRON_SERVER_REDIS_ADDRESS"
	EnvRedisServerPassword           = "ULTRON_SERVER_REDIS_PASSWORD"
	EnvRedisServerDatabase           = "ULTRON_SERVER_REDIS_DATABASE"

	LabelHostName     = "kubernetes.io/hostname"
	LabelInstanceType = "node.kubernetes.io/instance-type"

	MetadataName = "metadata.name"

	PriorityLow  PriorityEnum = false
	PriorityHigh PriorityEnum = true

	TopicNodeObserve = "ULTRON_TOPIC_NODE_OBSERVE"
	TopicPodObserve  = "ULTRON_TOPIC_POD_OBSERVE"
)
