package algorithm

import (
	ultron "github.com/be-heroes/ultron/pkg"
)

const (
	Alpha   = 1.0 // ResourceScore weight
	Beta    = 0.5 // StorageScore weight
	Gamma   = 0.5 // NetworkScore weight
	Delta   = 1.0 // PriceScore weight
	Epsilon = 1.0 // NodeScore weight
	Zeta    = 0.8 // PodScore weight
)

type IAlgorithm interface {
	ResourceScore(node ultron.WeightedNode, pod ultron.WeightedPod) float64
	StorageScore(node ultron.WeightedNode, pod ultron.WeightedPod) float64
	NetworkScore(node ultron.WeightedNode, pod ultron.WeightedPod) float64
	PriceScore(node ultron.WeightedNode) float64
	NodeScore(node ultron.WeightedNode) float64
	PodScore(pod ultron.WeightedPod) float64
	TotalScore(node ultron.WeightedNode, pod ultron.WeightedPod) float64
}

type Algorithm struct {
}

func NewAlgorithm() *Algorithm {
	return &Algorithm{}
}

func (a *Algorithm) ResourceScore(node ultron.WeightedNode, pod ultron.WeightedPod) float64 {
	var cpuScore, memScore float64

	if node.TotalCPU != 0 {
		cpuScore = (node.AvailableCPU - pod.RequestedCPU) / node.TotalCPU
	} else {
		cpuScore = 0.0
	}

	if node.TotalMemory != 0 {
		memScore = (node.AvailableMemory - pod.RequestedMemory) / node.TotalMemory
	} else {
		memScore = 0.0
	}

	return cpuScore + memScore
}

func (a *Algorithm) StorageScore(node ultron.WeightedNode, pod ultron.WeightedPod) float64 {
	if node.DiskType == pod.RequestedDiskType {
		return 1.0
	}

	return 0.0
}

func (a *Algorithm) NetworkScore(node ultron.WeightedNode, pod ultron.WeightedPod) float64 {
	if node.NetworkType == pod.RequestedNetworkType {
		return 1.0 - node.LatencyRate.Value
	}

	return 0.0
}

func (a *Algorithm) PriceScore(node ultron.WeightedNode) float64 {
	if node.Price == 0 {
		return 0.0
	}

	return 1.0 - (node.MedianPrice / node.Price)
}

func (a *Algorithm) NodeScore(node ultron.WeightedNode) float64 {
	if node.Price == 0 || node.MedianPrice == 0 {
		return 0.0
	}

	return node.InterruptionRate.Value * (node.Price / node.MedianPrice)
}

func (a *Algorithm) PodScore(pod ultron.WeightedPod) float64 {
	if pod.Priority == ultron.PriorityHigh {
		return 1.0
	}

	return 0.0
}

func (a *Algorithm) TotalScore(node ultron.WeightedNode, pod ultron.WeightedPod) float64 {
	resourceScore := Alpha * a.ResourceScore(node, pod)
	storageScore := Beta * a.StorageScore(node, pod)
	networkScore := Gamma * a.NetworkScore(node, pod)
	priceScore := Delta * a.PriceScore(node)
	nodeScore := Epsilon * a.NodeScore(node)
	podScore := Zeta * a.PodScore(pod)

	return resourceScore + storageScore + networkScore + priceScore - nodeScore + podScore
}
