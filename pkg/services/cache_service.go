package services

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"

	ultron "ultron/pkg"

	emma "github.com/emma-community/emma-go-sdk"
	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
)

type CacheService interface {
	AddCacheItem(key string, value interface{}, d time.Duration) error
	GetCacheItem(key string) (interface{}, error)
	GetAllComputeConfigurations() ([]ultron.ComputeConfiguration, error)
	GetEphemeralComputeConfigurations() ([]ultron.ComputeConfiguration, error)
	GetDurableComputeConfigurations() ([]ultron.ComputeConfiguration, error)
	GetWeightedNodes() ([]ultron.WeightedNode, error)
	GetWeightedInteruptionRates() ([]ultron.WeightedInteruptionRate, error)
	GetWeightedLatencyRates() ([]ultron.WeightedLatencyRate, error)
}

type ICacheService struct {
	memCache    *cache.Cache
	redisClient *redis.Client
}

func NewICacheService(innerCache *cache.Cache, redisClient *redis.Client) *ICacheService {
	if innerCache == nil && redisClient == nil {
		innerCache = cache.New(cache.NoExpiration, cache.NoExpiration)
	}

	return &ICacheService{
		memCache:    innerCache,
		redisClient: redisClient,
	}
}

func (c *ICacheService) AddCacheItem(key string, value interface{}, d time.Duration) error {

	if c.memCache != nil {
		c.memCache.Set(key, value, cache.DefaultExpiration)
	} else if c.redisClient != nil {
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(value)
		if err != nil {
			return err
		}

		c.redisClient.Set(context.Background(), key, buf.Bytes(), cache.DefaultExpiration)
	} else {
		return fmt.Errorf("both memCache and redisClient are nil")
	}

	return nil
}

func (c *ICacheService) GetCacheItem(key string) (interface{}, error) {
	var data []byte
	var err error
	var found bool
	var returnValue interface{}

	if c.memCache != nil {
		returnValue, found = c.memCache.Get(key)
		if !found {
			return nil, fmt.Errorf("key not found")
		}
	} else if c.redisClient != nil {
		data, err = c.redisClient.Get(context.Background(), key).Bytes()
		if err != nil {
			return nil, err
		}

		buf := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		err = dec.Decode(returnValue)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("both memCache and redisClient are nil")
	}

	return returnValue, nil
}

func (c *ICacheService) GetAllComputeConfigurations() ([]ultron.ComputeConfiguration, error) {
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

func (c *ICacheService) GetEphemeralComputeConfigurations() ([]ultron.ComputeConfiguration, error) {
	ephemeralConfigurationsRaw, err := c.GetCacheItem(ultron.CacheKeySpotVmConfigurations)
	if err != nil {
		return nil, err
	}

	ephemeralConfigurations := ephemeralConfigurationsRaw.([]emma.VmConfiguration)
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

func (c *ICacheService) GetDurableComputeConfigurations() ([]ultron.ComputeConfiguration, error) {
	durableConfigurationsRaw, err := c.GetCacheItem(ultron.CacheKeyDurableVmConfigurations)
	if err != nil {
		return nil, err
	}

	durableConfigurations := durableConfigurationsRaw.([]emma.VmConfiguration)
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

func (c *ICacheService) GetWeightedNodes() ([]ultron.WeightedNode, error) {
	weightedNodesInterface, err := c.GetCacheItem(ultron.CacheKeyWeightedNodes)
	if err != nil {
		return nil, err
	}

	return weightedNodesInterface.([]ultron.WeightedNode), nil
}

func (c *ICacheService) GetWeightedInteruptionRates() ([]ultron.WeightedInteruptionRate, error) {
	weightedInteruptionRatesInterface, err := c.GetCacheItem(ultron.CacheKeySpotVmConfigurationInteruptionRates)
	if err != nil {
		return nil, err
	}

	return weightedInteruptionRatesInterface.([]ultron.WeightedInteruptionRate), nil
}

func (c *ICacheService) GetWeightedLatencyRates() ([]ultron.WeightedLatencyRate, error) {
	weightedLatencyRatesInterface, err := c.GetCacheItem(ultron.CacheKeyDurableVmConfigurationLatencyRates)
	if err != nil {
		return nil, err
	}

	return weightedLatencyRatesInterface.([]ultron.WeightedLatencyRate), nil
}
