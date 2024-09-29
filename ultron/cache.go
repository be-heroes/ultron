package ultron

import (
	"fmt"
	"time"

	emma "github.com/emma-community/emma-go-sdk"
	"github.com/patrickmn/go-cache"
)

const (
	CacheKeyWeightedNodes           = "WEIGHTED_NODES"
	CacheKeyDurableVmConfigurations = "DURABLE_VMCONFIGURATION"
	CacheKeySpotVmConfigurations    = "SPOT_VMCONFIGURATION"
)

type Cache interface {
	AddCacheItem(key string, value interface{}, d time.Duration)
	GetAllComputeConfigurations() ([]ComputeConfiguration, error)
	GetEphemeralComputeConfigurations() ([]ComputeConfiguration, error)
	GetDurableComputeConfigurations() ([]ComputeConfiguration, error)
	GetWeightedNodes() ([]WeightedNode, error)
}

type ICache struct {
	memCache *cache.Cache
}

func NewICache(innerCache *cache.Cache) *ICache {
	if innerCache == nil {
		innerCache = cache.New(cache.NoExpiration, cache.NoExpiration)
	}

	return &ICache{
		memCache: innerCache,
	}
}

func (c *ICache) AddCacheItem(key string, value interface{}, d time.Duration) {
	c.memCache.Set(key, value, cache.DefaultExpiration)
}

func (c *ICache) GetAllComputeConfigurations() ([]ComputeConfiguration, error) {
	durableConfigurations, err := c.GetDurableComputeConfigurations()
	if err != nil {
		return nil, err
	}

	ephemeralConfigurations, err := c.GetEphemeralComputeConfigurations()
	if err != nil {
		return nil, err
	}

	return append(durableConfigurations, ephemeralConfigurations...), nil
}

func (c *ICache) GetEphemeralComputeConfigurations() ([]ComputeConfiguration, error) {
	ephemeralConfigsInterface, found := c.memCache.Get(CacheKeySpotVmConfigurations)
	if !found {
		return nil, fmt.Errorf("failed to get spot configurations from cache")
	}

	ephemeralConfigurations := ephemeralConfigsInterface.([]emma.VmConfiguration)
	var convertedConfigurations []ComputeConfiguration

	for i := range ephemeralConfigurations {
		convertedConfiguration := ComputeConfiguration{
			VmConfiguration: ephemeralConfigurations[i],
			ComputeType:     ComputeTypeEphemeral,
		}
		convertedConfigurations = append(convertedConfigurations, convertedConfiguration)
	}

	return convertedConfigurations, nil
}

func (c *ICache) GetDurableComputeConfigurations() ([]ComputeConfiguration, error) {
	durableConfigsInterface, found := c.memCache.Get(CacheKeyDurableVmConfigurations)
	if !found {
		return nil, fmt.Errorf("failed to get durable configurations from cache")
	}

	durableConfigurations := durableConfigsInterface.([]emma.VmConfiguration)
	var convertedConfigurations []ComputeConfiguration

	for i := range durableConfigurations {
		convertedConfiguration := ComputeConfiguration{
			VmConfiguration: durableConfigurations[i],
			ComputeType:     ComputeTypeDurable,
		}
		convertedConfigurations = append(convertedConfigurations, convertedConfiguration)
	}

	return convertedConfigurations, nil
}

func (c *ICache) GetWeightedNodes() ([]WeightedNode, error) {
	weightedNodesInterface, found := c.memCache.Get(CacheKeyWeightedNodes)
	if !found {
		return nil, fmt.Errorf("failed to get weighted nodes from cache")
	}

	return weightedNodesInterface.([]WeightedNode), nil
}
