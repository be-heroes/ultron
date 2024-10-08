package pkg

const (
	AnnotationDiskType    = "ultron.io/disk-type"
	AnnotationNetworkType = "ultron.io/network-type"
	AnnotationStorageSize = "ultron.io/storage-size"
	AnnotationPriority    = "ultron.io/priority"

	CacheKeyWeightedNodes                       = "WEIGHTED_NODES"
	CacheKeyDurableVmConfigurations             = "DURABLE_VMCONFIGURATION"
	CacheKeyDurableVmConfigurationLatencyRates  = "DURABLE_VMCONFIGURATION_LATENCY_RATES"
	CacheKeySpotVmConfigurations                = "SPOT_VMCONFIGURATION"
	CacheKeySpotVmConfigurationInteruptionRates = "SPOT_VMCONFIGURATION_INTERUPTION_RATES"

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
