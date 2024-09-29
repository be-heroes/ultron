package cache

import (
	"fmt"
	"time"

	ultron "ultron/internal"

	emma "github.com/emma-community/emma-go-sdk"
	"github.com/patrickmn/go-cache"
)

type Cache interface {
	AddCacheItem(key string, value interface{}, d time.Duration)
	GetAllComputeConfigurations() ([]ultron.ComputeConfiguration, error)
	GetEphemeralComputeConfigurations() ([]ultron.ComputeConfiguration, error)
	GetDurableComputeConfigurations() ([]ultron.ComputeConfiguration, error)
	GetWeightedNodes() ([]ultron.WeightedNode, error)
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

func (c *ICache) GetAllComputeConfigurations() ([]ultron.ComputeConfiguration, error) {
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

func (c *ICache) GetEphemeralComputeConfigurations() ([]ultron.ComputeConfiguration, error) {
	ephemeralConfigsInterface, found := c.memCache.Get(ultron.CacheKeySpotVmConfigurations)
	if !found {
		return nil, fmt.Errorf("failed to get spot configurations from cache")
	}

	ephemeralConfigurations := ephemeralConfigsInterface.([]emma.VmConfiguration)
	var convertedConfigurations []ultron.ComputeConfiguration

	for i := range ephemeralConfigurations {
		convertedConfiguration := ultron.ComputeConfiguration{
			VmConfiguration: ephemeralConfigurations[i],
			ComputeType:     ultron.ComputeTypeEphemeral,
		}
		convertedConfigurations = append(convertedConfigurations, convertedConfiguration)
	}

	return convertedConfigurations, nil
}

func (c *ICache) GetDurableComputeConfigurations() ([]ultron.ComputeConfiguration, error) {
	durableConfigsInterface, found := c.memCache.Get(ultron.CacheKeyDurableVmConfigurations)
	if !found {
		return nil, fmt.Errorf("failed to get durable configurations from cache")
	}

	durableConfigurations := durableConfigsInterface.([]emma.VmConfiguration)
	var convertedConfigurations []ultron.ComputeConfiguration

	for i := range durableConfigurations {
		convertedConfiguration := ultron.ComputeConfiguration{
			VmConfiguration: durableConfigurations[i],
			ComputeType:     ultron.ComputeTypeDurable,
		}
		convertedConfigurations = append(convertedConfigurations, convertedConfiguration)
	}

	return convertedConfigurations, nil
}

func (c *ICache) GetWeightedNodes() ([]ultron.WeightedNode, error) {
	weightedNodesInterface, found := c.memCache.Get(ultron.CacheKeyWeightedNodes)
	if !found {
		return nil, fmt.Errorf("failed to get weighted nodes from cache")
	}

	return weightedNodesInterface.([]ultron.WeightedNode), nil
}
