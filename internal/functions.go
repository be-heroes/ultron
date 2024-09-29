package internal

func (p PriorityEnum) String() string {
	if p {
		return "PriorityHigh"
	}

	return "PriorityLow"
}
