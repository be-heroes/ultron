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
				Selector: map[string]string{ultron.LabelInstanceType: instanceType, ultron.AnnotationManaged: "true"},
				Weights: map[string]float64{
					ultron.WeightKeyCpuAvailable:     float64(*computeConfiguration.VCpu),
					ultron.WeightKeyCpuTotal:         float64(*computeConfiguration.VCpu),
					ultron.WeightKeyMemoryAvailable:  float64(*computeConfiguration.RamGb),
					ultron.WeightKeyMemoryTotal:      float64(*computeConfiguration.RamGb),
					ultron.WeightKeyStorageAvailable: float64(*computeConfiguration.VolumeGb),
					ultron.WeightKeyPrice:            float64(*computeConfiguration.Cost.PricePerUnit),
				},
				Annotations: map[string]string{
					ultron.AnnotationInstanceType: instanceType,
					ultron.AnnotationDiskType:     wPod.Annotations[ultron.AnnotationDiskType],
					ultron.AnnotationNetworkType:  wPod.Annotations[ultron.AnnotationNetworkType],
				},
			}

			interuptionRate, err := cs.GetInteruptionRateForWeightedNode(wNode)
			if err != nil {
				return nil, err
			}

			if interuptionRate != nil {
				wNode.InterruptionRate = *interuptionRate
			} else {
				wNode.InterruptionRate = ultron.WeightedInteruptionRate{Weight: -1}
			}

			latencyRate, err := cs.GetLatencyRateForWeightedNode(wNode)
			if err != nil {
				return nil, err
			}

			if interuptionRate != nil {
				wNode.LatencyRate = *latencyRate
			} else {
				wNode.LatencyRate = ultron.WeightedLatencyRate{Weight: -1}
			}
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
		if wNode.Weights[ultron.WeightKeyCpuAvailable] < pod.Weights[ultron.WeightKeyCpuRequested] || wNode.Weights[ultron.WeightKeyMemoryAvailable] < pod.Weights[ultron.WeightKeyMemoryRequested] {
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
	if float64(*computeConfiguration.VCpu) > wNode.Weights[ultron.WeightKeyCpuAvailable] {
		return false
	}

	if float64(*computeConfiguration.RamGb) > wNode.Weights[ultron.WeightKeyMemoryAvailable] {
		return false
	}

	if float64(*computeConfiguration.VolumeGb) > wNode.Weights[ultron.WeightKeyStorageAvailable] {
		return false
	}

	if (*computeConfiguration.VolumeType) != wNode.Annotations[ultron.AnnotationDiskType] {
		return false
	}

	if !slices.Contains(computeConfiguration.CloudNetworkTypes, wNode.Annotations[ultron.AnnotationNetworkType]) {
		return false
	}

	return true
}

func (cs *ComputeService) ComputeConfigurationMatchesWeightedPodRequirements(computeConfiguration *ultron.ComputeConfiguration, wPod *ultron.WeightedPod) bool {
	if float64(*computeConfiguration.VCpu) < wPod.Weights[ultron.WeightKeyCpuRequested] {
		return false
	}

	if float64(*computeConfiguration.RamGb) < wPod.Weights[ultron.WeightKeyMemoryRequested] {
		return false
	}

	if float64(*computeConfiguration.VolumeGb) < wPod.Weights[ultron.WeightKeyStorageRequested] {
		return false
	}

	if (*computeConfiguration.VolumeType) != wPod.Annotations[ultron.AnnotationDiskType] {
		return false
	}

	if !slices.Contains(computeConfiguration.CloudNetworkTypes, wPod.Annotations[ultron.AnnotationNetworkType]) {
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
		if rate.Selector[ultron.LabelInstanceType] == wNode.Annotations[ultron.AnnotationInstanceType] {
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
		if rate.Selector[ultron.LabelInstanceType] == wNode.Annotations[ultron.AnnotationInstanceType] {
			match = &rate

			break
		}
	}

	return match, nil
}
