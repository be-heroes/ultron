package ultron

import (
	"math"
)

const (
	Alpha   = 1.0 // ResourceFitScore weight
	Beta    = 0.5 // DiskTypeScore weight
	Gamma   = 0.5 // NetworkTypeScore weight
	Delta   = 1.0 // PriceScore weight
	Epsilon = 1.0 // NodeStabilityScore weight
	Zeta    = 0.8 // WorkloadPriorityScore weight
)

func ResourceFitScore(node WeightedNode, pod WeightedPod) float64 {
	cpuScore := (node.AvailableCPU - pod.RequestedCPU) / node.TotalCPU
	memScore := (node.AvailableMemory - pod.RequestedMemory) / node.TotalMemory

	return cpuScore + memScore
}

func DiskTypeScore(node WeightedNode, pod WeightedPod) float64 {
	if node.DiskType == pod.RequestedDiskType {
		return 1.0
	}

	return 0.0
}

func NetworkTypeScore(node WeightedNode, pod WeightedPod) float64 {
	if node.NetworkType == pod.RequestedNetworkType {
		return 1.0
	}

	return 0.0
}

func PriceScore(node WeightedNode) float64 {
	return 1.0 - (node.MedianPrice / node.Price)
}

func NodeStabilityScore(node WeightedNode) float64 {
	return node.InterruptionRate * (node.Price / node.MedianPrice)
}

func WorkloadPriorityScore(pod WeightedPod) float64 {
	if pod.Priority == PriorityHigh {
		return 1.0
	}

	return 0.0
}

func TotalScore(node WeightedNode, pod WeightedPod) float64 {
	resourceFit := Alpha * ResourceFitScore(node, pod)
	diskType := Beta * DiskTypeScore(node, pod)
	networkType := Gamma * NetworkTypeScore(node, pod)
	price := Delta * PriceScore(node)
	stability := Epsilon * NodeStabilityScore(node)
	priority := Zeta * WorkloadPriorityScore(pod)

	return resourceFit + diskType + networkType + price - stability + priority
}

func FindBestNode(pod WeightedPod, nodes []WeightedNode) *WeightedNode {
	var bestNode WeightedNode
	highestScore := math.Inf(-1)

	for _, node := range nodes {
		if node.AvailableCPU < pod.RequestedCPU || node.AvailableMemory < pod.RequestedMemory {
			continue
		}

		score := TotalScore(node, pod)

		if score > highestScore {
			highestScore = score
			bestNode = node
		}
	}

	return &bestNode
}
