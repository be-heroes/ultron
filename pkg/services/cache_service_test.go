package services_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

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
	assert.NotNil(t, iCache, "NewCacheService should not return nil")
}

func TestGetEphemeralComputeConfigurations(t *testing.T) {
	// Arrange
	iCache := services.NewCacheService(nil, nil)

	computeConfigs := []ultron.ComputeConfiguration{
		{Identifier: nil, ComputeType: ultron.ComputeTypeEphemeral, Provider: nil, Location: nil, DataCenter: nil, OsType: nil, OsVersion: nil, CloudNetworkTypes: nil, VCpuType: nil, VCpu: nil, RamGb: nil, VolumeGb: nil, VolumeType: nil, Cost: nil},
		{Identifier: nil, ComputeType: ultron.ComputeTypeEphemeral, Provider: nil, Location: nil, DataCenter: nil, OsType: nil, OsVersion: nil, CloudNetworkTypes: nil, VCpuType: nil, VCpu: nil, RamGb: nil, VolumeGb: nil, VolumeType: nil, Cost: nil},
	}

	// Act
	iCache.AddCacheItem(ultron.CacheKeyEphemeralComputeConfigurations, computeConfigs, goCache.DefaultExpiration)

	getComputeConfigs, err := iCache.GetEphemeralComputeConfigurations()

	// Assert
	assert.NoError(t, err, "GetEphemeralComputeConfigurations should not return an error")
	assert.Equal(t, len(computeConfigs), len(getComputeConfigs), "Expected number of compute configurations does not match")

	for i, config := range getComputeConfigs {
		assert.Equal(t, ultron.ComputeTypeEphemeral, config.ComputeType, "Expected ComputeTypeEphemeral")
		assert.Equal(t, computeConfigs[i].Identifier, config.Identifier, "Expected matching Id")
	}
}

func TestGetEphemeralComputeConfigurations_NotFound(t *testing.T) {
	// Arrange
	iCache := services.NewCacheService(nil, nil)

	// Act
	_, err := iCache.GetEphemeralComputeConfigurations()

	// Assert
	assert.Error(t, err, "Expected an error when spot configurations are not found in the cache")
}

func TestGetDurableComputeConfigurations(t *testing.T) {
	// Arrange
	iCache := services.NewCacheService(nil, nil)

	computeConfigs := []ultron.ComputeConfiguration{
		{Identifier: nil, ComputeType: ultron.ComputeTypeDurable, Provider: nil, Location: nil, DataCenter: nil, OsType: nil, OsVersion: nil, CloudNetworkTypes: nil, VCpuType: nil, VCpu: nil, RamGb: nil, VolumeGb: nil, VolumeType: nil, Cost: nil},
		{Identifier: nil, ComputeType: ultron.ComputeTypeDurable, Provider: nil, Location: nil, DataCenter: nil, OsType: nil, OsVersion: nil, CloudNetworkTypes: nil, VCpuType: nil, VCpu: nil, RamGb: nil, VolumeGb: nil, VolumeType: nil, Cost: nil},
	}

	// Act
	iCache.AddCacheItem(ultron.CacheKeyDurableComputeConfigurations, computeConfigs, goCache.DefaultExpiration)

	getComputeConfigs, err := iCache.GetDurableComputeConfigurations()

	// Assert
	assert.NoError(t, err, "GetDurableComputeConfigurations should not return an error")
	assert.Equal(t, len(computeConfigs), len(getComputeConfigs), "Expected number of compute configurations does not match")

	for i, config := range getComputeConfigs {
		assert.Equal(t, ultron.ComputeTypeDurable, config.ComputeType, "Expected ComputeTypeDurable")
		assert.Equal(t, computeConfigs[i].Identifier, config.Identifier, "Expected matching Id")
	}
}

func TestGetDurableComputeConfigurations_NotFound(t *testing.T) {
	// Arrange
	iCache := services.NewCacheService(nil, nil)

	// Act
	_, err := iCache.GetDurableComputeConfigurations()

	// Assert
	assert.Error(t, err, "Expected an error when durable configurations are not found in the cache")
}
