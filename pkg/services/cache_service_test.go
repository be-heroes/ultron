package services_test

import (
	"testing"

	ultron "github.com/be-heroes/ultron/pkg"
	services "github.com/be-heroes/ultron/pkg/services"

	goCache "github.com/patrickmn/go-cache"
)

func TestNewCache(t *testing.T) {
	// Arrange
	var iCache *services.CacheService = nil

	// Act
	iCache = services.NewCacheService(nil, nil)

	// Assert
	if iCache == nil {
		t.Errorf("NewICache should not return nil")
	}
}

func TestGetEphemeralComputeConfigurations(t *testing.T) {
	// Arrange
	iCache := services.NewCacheService(nil, nil)

	computeConfigs := []ultron.ComputeConfiguration{
		{Id: nil, ComputeType: ultron.ComputeTypeEphemeral, ProviderId: nil, ProviderName: nil, LocationId: nil, LocationName: nil, DataCenterId: nil, DataCenterName: nil, OsId: nil, OsType: nil, OsVersion: nil, CloudNetworkTypes: nil, VCpuType: nil, VCpu: nil, RamGb: nil, VolumeGb: nil, VolumeType: nil, Cost: nil},
		{Id: nil, ComputeType: ultron.ComputeTypeEphemeral, ProviderId: nil, ProviderName: nil, LocationId: nil, LocationName: nil, DataCenterId: nil, DataCenterName: nil, OsId: nil, OsType: nil, OsVersion: nil, CloudNetworkTypes: nil, VCpuType: nil, VCpu: nil, RamGb: nil, VolumeGb: nil, VolumeType: nil, Cost: nil},
	}

	// Act
	iCache.AddCacheItem(ultron.CacheKeySpotVmConfigurations, computeConfigs, goCache.DefaultExpiration)

	getComputeConfigs, err := iCache.GetEphemeralComputeConfigurations()

	// Assert
	if err != nil {
		t.Fatalf("GetEphemeralComputeConfigurations returned an error: %v", err)
	}

	if len(computeConfigs) != len(getComputeConfigs) {
		t.Errorf("Expected %d compute configurations, got %d", len(computeConfigs), len(getComputeConfigs))
	}

	for i, config := range getComputeConfigs {
		if config.ComputeType != ultron.ComputeTypeEphemeral {
			t.Errorf("Expected ComputeTypeEphemeral, got %s", config.ComputeType)
		}
		if config.Id != computeConfigs[i].Id {
			t.Errorf("Expected Id %d, got %d", computeConfigs[i].Id, config.Id)
		}
	}
}

func TestGetEphemeralComputeConfigurations_NotFound(t *testing.T) {
	// Arrange
	iCache := services.NewCacheService(nil, nil)

	// Act
	_, err := iCache.GetEphemeralComputeConfigurations()

	// Assert
	if err == nil {
		t.Errorf("Expected an error when spot configurations are not found in the cache")
	}
}

func TestGetDurableComputeConfigurations(t *testing.T) {
	// Arrange
	iCache := services.NewCacheService(nil, nil)

	computeConfigs := []ultron.ComputeConfiguration{
		{Id: nil, ComputeType: ultron.ComputeTypeDurable, ProviderId: nil, ProviderName: nil, LocationId: nil, LocationName: nil, DataCenterId: nil, DataCenterName: nil, OsId: nil, OsType: nil, OsVersion: nil, CloudNetworkTypes: nil, VCpuType: nil, VCpu: nil, RamGb: nil, VolumeGb: nil, VolumeType: nil, Cost: nil},
		{Id: nil, ComputeType: ultron.ComputeTypeDurable, ProviderId: nil, ProviderName: nil, LocationId: nil, LocationName: nil, DataCenterId: nil, DataCenterName: nil, OsId: nil, OsType: nil, OsVersion: nil, CloudNetworkTypes: nil, VCpuType: nil, VCpu: nil, RamGb: nil, VolumeGb: nil, VolumeType: nil, Cost: nil},
	}

	// Act
	iCache.AddCacheItem(ultron.CacheKeyDurableVmConfigurations, computeConfigs, goCache.DefaultExpiration)

	getComputeConfigs, err := iCache.GetDurableComputeConfigurations()

	// Assert
	if err != nil {
		t.Fatalf("GetDurableComputeConfigurations returned an error: %v", err)
	}

	if len(computeConfigs) != len(getComputeConfigs) {
		t.Errorf("Expected %d compute configurations, got %d", len(computeConfigs), len(getComputeConfigs))
	}

	for i, config := range getComputeConfigs {
		if config.ComputeType != ultron.ComputeTypeDurable {
			t.Errorf("Expected ComputeTypeDurable, got %s", config.ComputeType)
		}
		if config.Id != computeConfigs[i].Id {
			t.Errorf("Expected Id %d, got %d", computeConfigs[i].Id, config.Id)
		}
	}
}

func TestGetDurableComputeConfigurations_NotFound(t *testing.T) {
	// Arrange
	iCache := services.NewCacheService(nil, nil)

	// Act
	_, err := iCache.GetDurableComputeConfigurations()

	// Assert
	if err == nil {
		t.Errorf("Expected an error when spot configurations are not found in the cache")
	}
}
