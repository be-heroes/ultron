package ultron

const (
	Alpha   = 1.0 // ResourceFitScore weight
	Beta    = 0.5 // DiskTypeScore weight
	Gamma   = 0.5 // NetworkTypeScore weight
	Delta   = 1.0 // PriceScore weight
	Epsilon = 1.0 // NodeStabilityScore weight
	Zeta    = 0.8 // WorkloadPriorityScore weight
)

type Algorithm interface {
	ResourceFitScore(node WeightedNode, pod WeightedPod) float64
	DiskTypeScore(node WeightedNode, pod WeightedPod) float64
	NetworkTypeScore(node WeightedNode, pod WeightedPod) float64
	PriceScore(node WeightedNode) float64
	NodeStabilityScore(node WeightedNode) float64
	WorkloadPriorityScore(pod WeightedPod) float64
	Score(node WeightedNode, pod WeightedPod) float64
}

type IAlgorithm struct {
}

func NewIAlgorithm() *IAlgorithm {
	return &IAlgorithm{}
}

func (a *IAlgorithm) ResourceFitScore(node WeightedNode, pod WeightedPod) float64 {
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

func (a *IAlgorithm) DiskTypeScore(node WeightedNode, pod WeightedPod) float64 {
	if node.DiskType == pod.RequestedDiskType {
		return 1.0
	}

	return 0.0
}

func (a *IAlgorithm) NetworkTypeScore(node WeightedNode, pod WeightedPod) float64 {
	if node.NetworkType == pod.RequestedNetworkType {
		return 1.0
	}

	return 0.0
}

func (a *IAlgorithm) PriceScore(node WeightedNode) float64 {
	if node.Price == 0 {
		return 0.0
	}

	return 1.0 - (node.MedianPrice / node.Price)
}

func (a *IAlgorithm) NodeStabilityScore(node WeightedNode) float64 {
	if node.Price == 0 || node.MedianPrice == 0 {
		return 0.0
	}

	return node.InterruptionRate * (node.Price / node.MedianPrice)
}

func (a *IAlgorithm) WorkloadPriorityScore(pod WeightedPod) float64 {
	if pod.Priority == PriorityHigh {
		return 1.0
	}

	return 0.0
}

func (a *IAlgorithm) Score(node WeightedNode, pod WeightedPod) float64 {
	resourceFit := Alpha * a.ResourceFitScore(node, pod)
	diskType := Beta * a.DiskTypeScore(node, pod)
	networkType := Gamma * a.NetworkTypeScore(node, pod)
	price := Delta * a.PriceScore(node)
	stability := Epsilon * a.NodeStabilityScore(node)
	priority := Zeta * a.WorkloadPriorityScore(pod)

	return resourceFit + diskType + networkType + price - stability + priority
}
