package ultron

import (
	"math"
)

// Node represents a Kubernetes Node with relevant attributes.
type Node struct {
	Name             string
	AvailableCPU     float64
	TotalCPU         float64
	AvailableMemory  float64
	TotalMemory      float64
	AvailableStorage float64
	DiskType         string
	NetworkType      string
	Price            float64
	MedianPrice      float64
	Type             string
	InterruptionRate float64
}

// Pod represents a Kubernetes Pod with its resource requirements and other attributes.
type Pod struct {
	Name                 string
	RequestedCPU         float64
	RequestedMemory      float64
	RequestedStorage     float64
	RequestedDiskType    string
	RequestedNetworkType string
	LimitCPU             float64
	LimitMemory          float64
	Priority             string // Can be "HighPriority" or "LowPriority"
}

// Weights for the different scoring components.
const (
	Alpha   = 1.0 // ResourceFitScore weight
	Beta    = 0.5 // DiskTypeScore weight
	Gamma   = 0.5 // NetworkTypeScore weight
	Delta   = 1.0 // PriceScore weight
	Epsilon = 1.0 // NodeStabilityScore weight
	Zeta    = 0.8 // WorkloadPriorityScore weight
)

// Calculate ResourceFitScore for a given Node and Pod.
func ResourceFitScore(node Node, pod Pod) float64 {
	cpuScore := (node.AvailableCPU - pod.RequestedCPU) / node.TotalCPU
	memScore := (node.AvailableMemory - pod.RequestedMemory) / node.TotalMemory
	return cpuScore + memScore
}

// Calculate DiskTypeScore for a given Node and Pod.
func DiskTypeScore(node Node, pod Pod) float64 {
	if node.DiskType == pod.RequestedDiskType {
		return 1.0
	}
	return 0.0
}

// Calculate NetworkTypeScore for a given Node and Pod.
func NetworkTypeScore(node Node, pod Pod) float64 {
	if node.NetworkType == pod.RequestedNetworkType {
		return 1.0
	}
	return 0.0
}

// Calculate PriceScore for a given Node.
func PriceScore(node Node) float64 {
	return 1.0 - (node.MedianPrice / node.Price)
}

// Calculate NodeStabilityScore for a given Node.
func NodeStabilityScore(node Node) float64 {
	return node.InterruptionRate * (node.Price / node.MedianPrice)
}

// Calculate WorkloadPriorityScore for a given Pod.
func WorkloadPriorityScore(pod Pod) float64 {
	if pod.Priority == "HighPriority" {
		return 1.0
	}
	return 0.0
}

// Calculate the total score for a Node given a Pod.
func TotalScore(node Node, pod Pod) float64 {
	resourceFit := Alpha * ResourceFitScore(node, pod)
	diskType := Beta * DiskTypeScore(node, pod)
	networkType := Gamma * NetworkTypeScore(node, pod)
	price := Delta * PriceScore(node)
	stability := Epsilon * NodeStabilityScore(node)
	priority := Zeta * WorkloadPriorityScore(pod)

	// Total score
	return resourceFit + diskType + networkType + price - stability + priority
}

// Find the best Node for a Pod from a list of Nodes.
func FindBestNode(pod Pod, nodes []Node) *Node {
	var bestNode *Node
	highestScore := math.Inf(-1)

	for _, node := range nodes {
		// Skip nodes that can't meet the basic requirements
		if node.AvailableCPU < pod.RequestedCPU || node.AvailableMemory < pod.RequestedMemory {
			continue
		}

		// Calculate the score for this Node
		score := TotalScore(node, pod)

		// Update the best Node if this score is higher
		if score > highestScore {
			highestScore = score
			bestNode = &node
		}
	}

	return bestNode
}
