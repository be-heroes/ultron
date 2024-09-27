package ultron

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"

	emmaSdk "github.com/emma-community/emma-go-sdk"
	"github.com/patrickmn/go-cache"
)

const (
	CacheKeyWNodes                  = "WNODES"
	CacheKeyDurableVmConfigurations = "DURABLE_VMCONFIGURATION"
	CacheKeySpotVmConfigurations    = "SPOT_VMCONFIGURATION"
)

func PopulateCache() {
	apiClient := emmaSdk.NewAPIClient(emmaSdk.NewConfiguration())
	credentials := emmaSdk.Credentials{ClientId: os.Getenv("EMMA_CLIENT_ID"), ClientSecret: os.Getenv("EMMA_CLIENT_SECRET")}
	token, resp, err := apiClient.AuthenticationAPI.IssueToken(context.Background()).Credentials(credentials).Execute()
	if err != nil {
		log.Fatalf("Failed to fetch token: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		log.Fatalf("Failed to fetch token: %v", string(body))
	}

	auth := context.WithValue(context.Background(), emmaSdk.ContextAccessToken, token.GetAccessToken())
	durableConfigs, resp, err := apiClient.ComputeInstancesConfigurationsAPI.GetVmConfigs(auth).Execute()
	if err != nil {
		log.Fatalf("Failed to fetch durable configs: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		log.Fatalf("Failed to fetch durable configs: %v", string(body))
	}

	spotConfigs, resp, err := apiClient.ComputeInstancesConfigurationsAPI.GetSpotConfigs(auth).Execute()
	if err != nil {
		log.Fatalf("Failed to fetch spot configs: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		log.Fatalf("Failed to fetch spot configs: %v", string(body))
	}

	wNodes, err := GetWeightedNodes()
	if err != nil {
		log.Fatalf("Failed to fetch nodes: %v", err)
	}

	Cache.Set(CacheKeyWNodes, wNodes, cache.DefaultExpiration)
	Cache.Set(CacheKeyDurableVmConfigurations, durableConfigs.Content, cache.DefaultExpiration)
	Cache.Set(CacheKeySpotVmConfigurations, spotConfigs.Content, cache.DefaultExpiration)
}
