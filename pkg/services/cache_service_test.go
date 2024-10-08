package services_test

import (
	"testing"

	ultron "github.com/be-heroes/ultron/pkg"
	services "github.com/be-heroes/ultron/pkg/services"

	emma "github.com/emma-community/emma-go-sdk"
	goCache "github.com/patrickmn/go-cache"
)

func TestNewICache(t *testing.T) {
	// Arrange
	var iCache *services.ICacheService = nil

	// Act
	iCache = services.NewICacheService(nil, nil)

	// Assert
	if iCache == nil {
		t.Errorf("NewICache should not return nil")
	}
}

func TestGetEphemeralComputeConfigurations(t *testing.T) {
	// Arrange
	iCache := services.NewICacheService(nil, nil)

	vmConfigs := []emma.VmConfiguration{
		{Id: nil, ProviderId: nil, ProviderName: nil, LocationId: nil, LocationName: nil, DataCenterId: nil, DataCenterName: nil, OsId: nil, OsType: nil, OsVersion: nil, CloudNetworkTypes: nil, VCpuType: nil, VCpu: nil, RamGb: nil, VolumeGb: nil, VolumeType: nil, Cost: nil},
		{Id: nil, ProviderId: nil, ProviderName: nil, LocationId: nil, LocationName: nil, DataCenterId: nil, DataCenterName: nil, OsId: nil, OsType: nil, OsVersion: nil, CloudNetworkTypes: nil, VCpuType: nil, VCpu: nil, RamGb: nil, VolumeGb: nil, VolumeType: nil, Cost: nil},
	}

	// Act
	iCache.AddCacheItem(ultron.CacheKeySpotVmConfigurations, vmConfigs, goCache.DefaultExpiration)

	computeConfigs, err := iCache.GetEphemeralComputeConfigurations()

	// Assert
	if err != nil {
		t.Fatalf("GetEphemeralComputeConfigurations returned an error: %v", err)
	}

	if len(computeConfigs) != len(vmConfigs) {
		t.Errorf("Expected %d compute configurations, got %d", len(vmConfigs), len(computeConfigs))
	}

	for i, config := range computeConfigs {
		if config.ComputeType != ultron.ComputeTypeEphemeral {
			t.Errorf("Expected ComputeTypeEphemeral, got %s", config.ComputeType)
		}
		if config.VmConfiguration.Id != vmConfigs[i].Id {
			t.Errorf("Expected Id %d, got %d", vmConfigs[i].Id, config.VmConfiguration.Id)
		}
	}
}

func TestGetEphemeralComputeConfigurations_NotFound(t *testing.T) {
	// Arrange
	iCache := services.NewICacheService(nil, nil)

	// Act
	_, err := iCache.GetEphemeralComputeConfigurations()

	// Assert
	if err == nil {
		t.Errorf("Expected an error when spot configurations are not found in the cache")
	}
}

func TestGetDurableComputeConfigurations(t *testing.T) {
	// Arrange
	iCache := services.NewICacheService(nil, nil)

	vmConfigs := []emma.VmConfiguration{
		{Id: nil, ProviderId: nil, ProviderName: nil, LocationId: nil, LocationName: nil, DataCenterId: nil, DataCenterName: nil, OsId: nil, OsType: nil, OsVersion: nil, CloudNetworkTypes: nil, VCpuType: nil, VCpu: nil, RamGb: nil, VolumeGb: nil, VolumeType: nil, Cost: nil},
		{Id: nil, ProviderId: nil, ProviderName: nil, LocationId: nil, LocationName: nil, DataCenterId: nil, DataCenterName: nil, OsId: nil, OsType: nil, OsVersion: nil, CloudNetworkTypes: nil, VCpuType: nil, VCpu: nil, RamGb: nil, VolumeGb: nil, VolumeType: nil, Cost: nil},
	}

	// Act
	iCache.AddCacheItem(ultron.CacheKeyDurableVmConfigurations, vmConfigs, goCache.DefaultExpiration)

	computeConfigs, err := iCache.GetDurableComputeConfigurations()

	// Assert
	if err != nil {
		t.Fatalf("GetDurableComputeConfigurations returned an error: %v", err)
	}

	if len(computeConfigs) != len(vmConfigs) {
		t.Errorf("Expected %d compute configurations, got %d", len(vmConfigs), len(computeConfigs))
	}

	for i, config := range computeConfigs {
		if config.ComputeType != ultron.ComputeTypeDurable {
			t.Errorf("Expected ComputeTypeDurable, got %s", config.ComputeType)
		}
		if config.VmConfiguration.Id != vmConfigs[i].Id {
			t.Errorf("Expected Id %d, got %d", vmConfigs[i].Id, config.VmConfiguration.Id)
		}
	}
}

func TestGetDurableComputeConfigurations_NotFound(t *testing.T) {
	// Arrange
	iCache := services.NewICacheService(nil, nil)

	// Act
	_, err := iCache.GetDurableComputeConfigurations()

	// Assert
	if err == nil {
		t.Errorf("Expected an error when spot configurations are not found in the cache")
	}
}
