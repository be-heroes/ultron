package ultron

const (
	Alpha   = 1.0 // ResourceFitScore weight
	Beta    = 0.5 // DiskTypeScore weight
	Gamma   = 0.5 // NetworkTypeScore weight
	Delta   = 1.0 // PriceScore weight
	Epsilon = 1.0 // NodeStabilityScore weight
	Zeta    = 0.8 // WorkloadPriorityScore weight
)

var ResourceFitScore = func(node WeightedNode, pod WeightedPod) float64 {
	var cpuScore, memScore float64

	if node.TotalCPU != 0 {
		cpuScore = (node.AvailableCPU - pod.RequestedCPU) / node.TotalCPU
	} else {
		cpuScore = 0.0 // Avoid division by zero, assume no fit for CPU if TotalCPU is 0
	}

	if node.TotalMemory != 0 {
		memScore = (node.AvailableMemory - pod.RequestedMemory) / node.TotalMemory
	} else {
		memScore = 0.0 // Avoid division by zero, assume no fit for Memory if TotalMemory is 0
	}

	return cpuScore + memScore
}

var DiskTypeScore = func(node WeightedNode, pod WeightedPod) float64 {
	if node.DiskType == pod.RequestedDiskType {
		return 1.0
	}

	return 0.0
}

var NetworkTypeScore = func(node WeightedNode, pod WeightedPod) float64 {
	if node.NetworkType == pod.RequestedNetworkType {
		return 1.0
	}

	return 0.0
}

var PriceScore = func(node WeightedNode) float64 {
	if node.Price == 0 {
		return 0.0 // Avoid division by zero, assume no price advantage if node.Price is 0
	}

	return 1.0 - (node.MedianPrice / node.Price)
}

var NodeStabilityScore = func(node WeightedNode) float64 {
	if node.Price == 0 || node.MedianPrice == 0 {
		return 0.0 // Avoid division by zero, assume no stability impact if Price or MedianPrice is 0
	}

	return node.InterruptionRate * (node.Price / node.MedianPrice)
}

var WorkloadPriorityScore = func(pod WeightedPod) float64 {
	if pod.Priority == PriorityHigh {
		return 1.0
	}

	return 0.0
}

var Score = func(node WeightedNode, pod WeightedPod) float64 {
	resourceFit := Alpha * ResourceFitScore(node, pod)
	diskType := Beta * DiskTypeScore(node, pod)
	networkType := Gamma * NetworkTypeScore(node, pod)
	price := Delta * PriceScore(node)
	stability := Epsilon * NodeStabilityScore(node)
	priority := Zeta * WorkloadPriorityScore(pod)

	return resourceFit + diskType + networkType + price - stability + priority
}
