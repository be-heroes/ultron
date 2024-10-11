package services

import (
	"math"
	"slices"
	"sort"

	ultron "github.com/be-heroes/ultron/pkg"
	algorithm "github.com/be-heroes/ultron/pkg/algorithm"
	mapper "github.com/be-heroes/ultron/pkg/mapper"

	corev1 "k8s.io/api/core/v1"
)

type IComputeService interface {
	MatchPodSpec(pod *corev1.Pod) (*ultron.WeightedNode, error)
	MatchWeightedPodToComputeConfiguration(wPod *ultron.WeightedPod) (*ultron.ComputeConfiguration, error)
	MatchWeightedNodeToComputeConfiguration(wNode *ultron.WeightedNode) (*ultron.ComputeConfiguration, error)
	MatchWeightedPodToWeightedNode(wPod *ultron.WeightedPod) (*ultron.WeightedNode, error)
	CalculateWeightedNodeMedianPrice(wNode *ultron.WeightedNode) (float64, error)
	ComputeConfigurationMatchesWeightedNodeRequirements(computeConfiguration *ultron.ComputeConfiguration, wNode *ultron.WeightedNode) bool
	ComputeConfigurationMatchesWeightedPodRequirements(computeConfiguration *ultron.ComputeConfiguration, wPod *ultron.WeightedPod) bool
	GetInteruptionRateForWeightedNode(wNode *ultron.WeightedNode) (*ultron.WeightedInteruptionRate, error)
	GetLatencyRateForWeightedNode(wNode *ultron.WeightedNode) (*ultron.WeightedLatencyRate, error)
}

type ComputeService struct {
	algorithm    algorithm.IAlgorithm
	cacheService ICacheService
	mapper       mapper.IMapper
}

func NewComputeService(algorithm algorithm.IAlgorithm, cacheService ICacheService, mapper mapper.IMapper) *ComputeService {
	return &ComputeService{
		algorithm:    algorithm,
		cacheService: cacheService,
		mapper:       mapper,
	}
}

func (cs *ComputeService) MatchPodSpec(pod *corev1.Pod) (*ultron.WeightedNode, error) {
	wPod, err := cs.mapper.MapPodToWeightedPod(pod)
	if err != nil {
		return nil, err
	}

	wNode, err := cs.MatchWeightedPodToWeightedNode(&wPod)
	if err != nil {
		return nil, err
	}

	if wNode == nil {
		computeConfiguration, err := cs.MatchWeightedPodToComputeConfiguration(&wPod)
		if err != nil {
			return nil, err
		}

		if computeConfiguration != nil {
			var instanceType string

			if computeConfiguration.ComputeType == ultron.ComputeTypeDurable {
				instanceType = ultron.DefaultDurableInstanceType
			} else {
				instanceType = ultron.DefaultEphemeralInstanceType
			}

			wNode = &ultron.WeightedNode{
				Selector:         map[string]string{ultron.LabelInstanceType: instanceType, ultron.AnnotationManaged: "true"},
				AvailableCPU:     float64(*computeConfiguration.VCpu),
				TotalCPU:         float64(*computeConfiguration.VCpu),
				AvailableMemory:  float64(*computeConfiguration.RamGb),
				TotalMemory:      float64(*computeConfiguration.RamGb),
				AvailableStorage: float64(*computeConfiguration.VolumeGb),
				DiskType:         wPod.RequestedDiskType,
				NetworkType:      wPod.RequestedNetworkType,
				Price:            float64(*computeConfiguration.Cost.PricePerUnit),
				InstanceType:     instanceType,
			}

			interuptionRate, err := cs.GetInteruptionRateForWeightedNode(wNode)
			if err != nil {
				return nil, err
			}

			wNode.InterruptionRate = *interuptionRate

			latencyRate, err := cs.GetLatencyRateForWeightedNode(wNode)
			if err != nil {
				return nil, err
			}

			wNode.LatencyRate = *latencyRate
		}
	}

	return wNode, nil
}

func (cs *ComputeService) MatchWeightedPodToComputeConfiguration(wPod *ultron.WeightedPod) (*ultron.ComputeConfiguration, error) {
	var suitableConfigs []ultron.ComputeConfiguration
	computeConfigurations, err := cs.cacheService.GetAllComputeConfigurations()
	if err != nil {
		return nil, err
	}

	for _, computeConfig := range computeConfigurations {
		if cs.ComputeConfigurationMatchesWeightedPodRequirements(&computeConfig, wPod) {
			suitableConfigs = append(suitableConfigs, computeConfig)
		}
	}

	if len(suitableConfigs) == 0 {
		return nil, nil
	}

	sort.Slice(suitableConfigs, func(i, j int) bool {
		return (*suitableConfigs[i].Cost.PricePerUnit) < (*suitableConfigs[j].Cost.PricePerUnit)
	})

	return &suitableConfigs[0], nil
}

func (cs *ComputeService) MatchWeightedNodeToComputeConfiguration(wNode *ultron.WeightedNode) (*ultron.ComputeConfiguration, error) {
	var suitableConfigs []ultron.ComputeConfiguration
	computeConfigurations, err := cs.cacheService.GetAllComputeConfigurations()
	if err != nil {
		return nil, err
	}

	for _, computeConfig := range computeConfigurations {
		if cs.ComputeConfigurationMatchesWeightedNodeRequirements(&computeConfig, wNode) {
			suitableConfigs = append(suitableConfigs, computeConfig)
		}
	}

	if len(suitableConfigs) == 0 {
		return nil, nil
	}

	sort.Slice(suitableConfigs, func(i, j int) bool {
		return (*suitableConfigs[i].Cost.PricePerUnit) < (*suitableConfigs[j].Cost.PricePerUnit)
	})

	return &suitableConfigs[0], nil
}

func (cs *ComputeService) MatchWeightedPodToWeightedNode(pod *ultron.WeightedPod) (*ultron.WeightedNode, error) {
	wNodes, err := cs.cacheService.GetWeightedNodes()
	if err != nil {
		return nil, err
	}

	var match ultron.WeightedNode
	highestScore := math.Inf(-1)

	for _, wNode := range wNodes {
		if wNode.AvailableCPU < pod.RequestedCPU || wNode.AvailableMemory < pod.RequestedMemory {
			continue
		}

		score := cs.algorithm.TotalScore(&wNode, pod)

		if score > highestScore {
			highestScore = score
			match = wNode
		}
	}

	return &match, nil
}

func (cs *ComputeService) CalculateWeightedNodeMedianPrice(wNode *ultron.WeightedNode) (float64, error) {
	var totalCost float64
	var matchCount int32
	computeConfigurations, err := cs.cacheService.GetAllComputeConfigurations()
	if err != nil {
		return 0, err
	}

	for _, computeConfig := range computeConfigurations {
		if cs.ComputeConfigurationMatchesWeightedNodeRequirements(&computeConfig, wNode) && computeConfig.Cost != nil && computeConfig.Cost.PricePerUnit != nil {
			totalCost += float64(*computeConfig.Cost.PricePerUnit)
			matchCount++
		}
	}

	if matchCount > 0 {
		return totalCost / float64(matchCount), nil
	} else {
		return 0, nil
	}
}

func (cs *ComputeService) ComputeConfigurationMatchesWeightedNodeRequirements(computeConfiguration *ultron.ComputeConfiguration, wNode *ultron.WeightedNode) bool {
	if float64(*computeConfiguration.VCpu) > wNode.AvailableCPU {
		return false
	}

	if float64(*computeConfiguration.RamGb) > wNode.AvailableMemory {
		return false
	}

	if float64(*computeConfiguration.VolumeGb) > wNode.AvailableStorage {
		return false
	}

	if (*computeConfiguration.VolumeType) != wNode.DiskType {
		return false
	}

	if !slices.Contains(computeConfiguration.CloudNetworkTypes, wNode.NetworkType) {
		return false
	}

	return true
}

func (cs *ComputeService) ComputeConfigurationMatchesWeightedPodRequirements(computeConfiguration *ultron.ComputeConfiguration, wPod *ultron.WeightedPod) bool {
	if float64(*computeConfiguration.VCpu) < wPod.RequestedCPU {
		return false
	}

	if float64(*computeConfiguration.RamGb) < wPod.RequestedMemory {
		return false
	}

	if float64(*computeConfiguration.VolumeGb) < wPod.RequestedStorage {
		return false
	}

	if (*computeConfiguration.VolumeType) != wPod.RequestedDiskType {
		return false
	}

	if !slices.Contains(computeConfiguration.CloudNetworkTypes, wPod.RequestedNetworkType) {
		return false
	}

	return true
}

func (cs *ComputeService) GetInteruptionRateForWeightedNode(wNode *ultron.WeightedNode) (match *ultron.WeightedInteruptionRate, err error) {
	rates, err := cs.cacheService.GetWeightedInteruptionRates()
	if err != nil {
		return nil, err
	}

	for _, rate := range rates {
		if rate.Selector[ultron.LabelInstanceType] == wNode.InstanceType {
			match = &rate
			break
		}
	}

	return match, nil
}

func (cs *ComputeService) GetLatencyRateForWeightedNode(wNode *ultron.WeightedNode) (match *ultron.WeightedLatencyRate, err error) {
	rates, err := cs.cacheService.GetWeightedLatencyRates()
	if err != nil {
		return nil, err
	}

	for _, rate := range rates {
		if rate.Selector[ultron.LabelInstanceType] == wNode.InstanceType {
			match = &rate
			break
		}
	}

	return match, nil
}
