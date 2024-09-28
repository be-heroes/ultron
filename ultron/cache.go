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

	Cache.Set(CacheKeyWeightedNodes, wNodes, cache.DefaultExpiration)
	Cache.Set(CacheKeyDurableVmConfigurations, durableConfigs.Content, cache.DefaultExpiration)
	Cache.Set(CacheKeySpotVmConfigurations, spotConfigs.Content, cache.DefaultExpiration)

	return nil
}

func GetAllVmConfigurationsFromCache() ([]emma.VmConfiguration, error) {
	dureableConfigurations, err := GetDurableConfigurationsFromCache()
	if err != nil {
		return nil, err
	}

	spotConfigurations, err := GetSpotConfigurationsFromCache()
	if err != nil {
		return nil, err
	}

	return append(dureableConfigurations, spotConfigurations...), nil
}

func GetSpotConfigurationsFromCache() ([]emma.VmConfiguration, error) {
	spotConfigsInterface, found := Cache.Get(CacheKeySpotVmConfigurations)

	if !found {
		return nil, fmt.Errorf("failed to get spot VmConfiguration list from cache")
	}

	return spotConfigsInterface.([]emma.VmConfiguration), nil
}

func GetDurableConfigurationsFromCache() ([]emma.VmConfiguration, error) {
	durableConfigsInterface, found := Cache.Get(CacheKeyDurableVmConfigurations)

	if !found {
		return nil, fmt.Errorf("failed to get durable VmConfiguration list from cache")
	}

	return durableConfigsInterface.([]emma.VmConfiguration), nil
}

func GetWeightedNodesFromCache() ([]WeightedNode, error) {
	weightedNodesInterface, found := Cache.Get(CacheKeyWeightedNodes)

	if !found {
		return nil, fmt.Errorf("failed to get weighted nodes from cache")
	}

	return weightedNodesInterface.([]WeightedNode), nil
}
