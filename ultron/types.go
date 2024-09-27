package ultron

type WeightedNode struct {
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
	Selector         string
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
	Selector             string
}

type PriorityEnum bool

const (
	LowPriority  PriorityEnum = false
	HighPriority PriorityEnum = true
)

func (p PriorityEnum) String() string {
	if p == true {
		return "HighPriority"
	}

	return "LowPriority"
}
