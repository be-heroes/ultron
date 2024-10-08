package pkg

const (
	EnvServerAddress                 = "ULTRON_SERVER_ADDRESS"
	EnvServerCertificateOrganization = "ULTRON_SERVER_CERTIFICATE_ORGANIZATION"
	EnvServerCertificateCommonName   = "ULTRON_SERVER_CERTIFICATE_COMMON_NAME"
	EnvServerCertificateDnsNames     = "ULTRON_SERVER_CERTIFICATE_DNS_NAMES"
	EnvServerCertificateIpAddresses  = "ULTRON_SERVER_CERTIFICATE_IP_ADDRESSES"
	EnvServerCertificateExportPath   = "ULTRON_SERVER_CERTIFICATE_EXPORT_PATH"
	EnvRedisServerAddress            = "ULTRON_REDIS_SERVER_ADDRESS"
	EnvRedisServerPassword           = "ULTRON_REDIS_SERVER_PASSWORD"
	EnvRedisServerDatabase           = "ULTRON_REDIS_SERVER_DATABASE"

	AnnotationDiskType    = "ultron.io/disk-type"
	AnnotationNetworkType = "ultron.io/network-type"
	AnnotationStorageSize = "ultron.io/storage-size"
	AnnotationPriority    = "ultron.io/priority"

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

	LabelHostName     = "kubernetes.io/hostname"
	LabelInstanceType = "node.kubernetes.io/instance-type"

	MetadataName = "metadata.name"

	PriorityLow  PriorityEnum = false
	PriorityHigh PriorityEnum = true
)
