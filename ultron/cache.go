package ultron

import (
	"context"
	"fmt"
	"io"
	"net/http"

	emma "github.com/emma-community/emma-go-sdk"
	"github.com/patrickmn/go-cache"
)

const (
	CacheKeyWeightedNodes           = "WEIGHTED_NODES"
	CacheKeyDurableVmConfigurations = "DURABLE_VMCONFIGURATION"
	CacheKeySpotVmConfigurations    = "SPOT_VMCONFIGURATION"
)

var memCache = cache.New(cache.NoExpiration, cache.NoExpiration)

func InitializeCache(credentials emma.Credentials, kubernetesMasterUrl string, kubernetesConfigPath string) error {
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

	spotConfigs, resp, err := apiClient.ComputeInstancesConfigurationsAPI.GetSpotConfigs(auth).Execute()
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		_, err := io.ReadAll(resp.Body)

		return err
	}

	wNodes, err := GetWeightedNodes(kubernetesMasterUrl, kubernetesConfigPath)
	if err != nil {
		return err
	}

	memCache.Set(CacheKeyWeightedNodes, wNodes, cache.DefaultExpiration)
	memCache.Set(CacheKeyDurableVmConfigurations, durableConfigs.Content, cache.DefaultExpiration)
	memCache.Set(CacheKeySpotVmConfigurations, spotConfigs.Content, cache.DefaultExpiration)

	return nil
}

func GetAllComputeConfigurationsFromCache() ([]ComputeConfiguration, error) {
	dureableConfigurations, err := GetDurableComputeConfigurationsFromCache()
	if err != nil {
		return nil, err
	}

	ephemeralConfigurations, err := GetEphemeralComputeConfigurationsFromCache()
	if err != nil {
		return nil, err
	}

	return append(dureableConfigurations, ephemeralConfigurations...), nil
}

func GetEphemeralComputeConfigurationsFromCache() ([]ComputeConfiguration, error) {
	ephemeralConfigsInterface, found := memCache.Get(CacheKeySpotVmConfigurations)

	if !found {
		return nil, fmt.Errorf("failed to get spot configurations from cache")
	}

	ephemeralConfigurations := ephemeralConfigsInterface.([]ComputeConfiguration)

	for i := range ephemeralConfigurations {
		ephemeralConfigurations[i].ComputeType = ComputeTypeEphemeral
	}

	return ephemeralConfigurations, nil
}

func GetDurableComputeConfigurationsFromCache() ([]ComputeConfiguration, error) {
	durableConfigsInterface, found := memCache.Get(CacheKeyDurableVmConfigurations)

	if !found {
		return nil, fmt.Errorf("failed to get durable configurations from cache")
	}

	durableConfigurations := durableConfigsInterface.([]ComputeConfiguration)

	for i := range durableConfigurations {
		durableConfigurations[i].ComputeType = ComputeTypeDurable
	}

	return durableConfigurations, nil
}

func GetWeightedNodesFromCache() ([]WeightedNode, error) {
	weightedNodesInterface, found := memCache.Get(CacheKeyWeightedNodes)

	if !found {
		return nil, fmt.Errorf("failed to get weighted nodes from cache")
	}

	return weightedNodesInterface.([]WeightedNode), nil
}
