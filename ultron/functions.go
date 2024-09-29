package ultron

import (
	"strconv"
)

func (p PriorityEnum) String() string {
	if p {
		return "PriorityHigh"
	}

	return "PriorityLow"
}

func getAnnotationOrDefault(annotations map[string]string, key, defaultValue string) string {
	if value, exists := annotations[key]; exists {
		return value
	}

	return defaultValue
}

func getFloatAnnotationOrDefault(annotations map[string]string, key string, defaultValue float64) float64 {
	if valueStr, exists := annotations[key]; exists {
		if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
			return value
		}
	}

	return defaultValue
}

func getPriorityFromAnnotation(annotations map[string]string) PriorityEnum {
	if value, exists := annotations[AnnotationPriority]; exists {
		switch value {
		case "PriorityHigh":
			return PriorityHigh
		case "PriorityLow":
			return PriorityLow
		}
	}

	return DefaultPriority
}
