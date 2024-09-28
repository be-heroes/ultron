package ultron

import (
	"context"
	"io"
	"net/http"

	emmaSdk "github.com/emma-community/emma-go-sdk"
	"github.com/patrickmn/go-cache"
)

const (
	CacheKeyWeightedNodes           = "WEIGHTED_NODES"
	CacheKeyDurableVmConfigurations = "DURABLE_VMCONFIGURATION"
	CacheKeySpotVmConfigurations    = "SPOT_VMCONFIGURATION"
)

func InitializeCache(credentials emmaSdk.Credentials, kubernetesMasterUrl string, kubernetesConfigPath string) error {
	apiClient := emmaSdk.NewAPIClient(emmaSdk.NewConfiguration())
	token, resp, err := apiClient.AuthenticationAPI.IssueToken(context.Background()).Credentials(credentials).Execute()
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		_, err := io.ReadAll(resp.Body)

		return err
	}

	auth := context.WithValue(context.Background(), emmaSdk.ContextAccessToken, token.GetAccessToken())
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
