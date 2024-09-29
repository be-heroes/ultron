package internal

import (
	emma "github.com/emma-community/emma-go-sdk"
)

type ComputeType string
type PriorityEnum bool

type WeightedNode struct {
	AvailableCPU     float64
	TotalCPU         float64
	AvailableMemory  float64
	TotalMemory      float64
	AvailableStorage float64
	TotalStorage     float64
	DiskType         string
	NetworkType      string
	Price            float64
	MedianPrice      float64
	InstanceType     string
	InterruptionRate float64
	Selector         map[string]string
}

type WeightedPod struct {
	RequestedCPU         float64
	RequestedMemory      float64
	RequestedStorage     float64
	RequestedDiskType    string
	RequestedNetworkType string
	LimitCPU             float64
	LimitMemory          float64
	Priority             PriorityEnum
	Selector             map[string]string
}

type ComputeConfiguration struct {
	emma.VmConfiguration
	ComputeType ComputeType
}
