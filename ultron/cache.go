package ultron

import (
	"context"
	"fmt"
	"io"
	"net/http"
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
	Initialize(credentials emma.Credentials, kubernetesClient *IKubernetesClient) error
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
	if innerCache != nil {
		innerCache = cache.New(cache.NoExpiration, cache.NoExpiration)
	}

	return &ICache{
		memCache: innerCache,
	}
}

func (s *ICache) Initialize(credentials emma.Credentials, kubernetesClient *IKubernetesClient) error {
	apiClient := emma.NewAPIClient(emma.NewConfiguration())
	token, resp, err := apiClient.AuthenticationAPI.IssueToken(context.Background()).Credentials(credentials).Execute()
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		_, err := io.ReadAll(resp.Body)
		return err
	}

	auth := context.WithValue(context.Background(), emma.ContextAccessToken, token.GetAccessToken())
	durableConfigs, resp, err := apiClient.ComputeInstancesConfigurationsAPI.GetVmConfigs(auth).Execute()
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		_, err := io.ReadAll(resp.Body)
		return err
	}

	ephemeralConfigs, resp, err := apiClient.ComputeInstancesConfigurationsAPI.GetSpotConfigs(auth).Execute()
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		_, err := io.ReadAll(resp.Body)
		return err
	}

	wNodes, err := kubernetesClient.GetWeightedNodes()
	if err != nil {
		return err
	}

	s.memCache.Set(CacheKeyWeightedNodes, wNodes, cache.DefaultExpiration)
	s.memCache.Set(CacheKeyDurableVmConfigurations, durableConfigs.Content, cache.DefaultExpiration)
	s.memCache.Set(CacheKeySpotVmConfigurations, ephemeralConfigs.Content, cache.DefaultExpiration)

	return nil
}

func (s *ICache) AddCacheItem(key string, value interface{}, d time.Duration) {
	s.memCache.Set(key, value, cache.DefaultExpiration)
}

func (s *ICache) GetAllComputeConfigurations() ([]ComputeConfiguration, error) {
	durableConfigurations, err := s.GetDurableComputeConfigurations()
	if err != nil {
		return nil, err
	}

	ephemeralConfigurations, err := s.GetEphemeralComputeConfigurations()
	if err != nil {
		return nil, err
	}

	return append(durableConfigurations, ephemeralConfigurations...), nil
}

func (s *ICache) GetEphemeralComputeConfigurations() ([]ComputeConfiguration, error) {
	ephemeralConfigsInterface, found := s.memCache.Get(CacheKeySpotVmConfigurations)
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

func (s *ICache) GetDurableComputeConfigurations() ([]ComputeConfiguration, error) {
	durableConfigsInterface, found := s.memCache.Get(CacheKeyDurableVmConfigurations)
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

func (s *ICache) GetWeightedNodes() ([]WeightedNode, error) {
	weightedNodesInterface, found := s.memCache.Get(CacheKeyWeightedNodes)
	if !found {
		return nil, fmt.Errorf("failed to get weighted nodes from cache")
	}

	return weightedNodesInterface.([]WeightedNode), nil
}
