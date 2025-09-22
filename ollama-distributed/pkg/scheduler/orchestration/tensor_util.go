package orchestration

import (
	"strconv"
)

// ConvertToInt converts various types to int
func ConvertToInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int32:
		return int(val)
	case int64:
		return int(val)
	case float32:
		return int(val)
	case float64:
		return int(val)
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return int(f)
		}
	}
	return 0
}

// GetIntFromMetadata retrieves an integer value from metadata
func GetIntFromMetadata(m map[string]interface{}, key string) int {
	if m == nil {
		return 0
	}
	if v, ok := m[key]; ok {
		return ConvertToInt(v)
	}
	return 0
}

// ToFloat32 converts various types to float32
func ToFloat32(w interface{}) (float32, bool) {
	switch val := w.(type) {
	case float32:
		return val, true
	case float64:
		return float32(val), true
	case int:
		return float32(val), true
	case int32:
		return float32(val), true
	case int64:
		return float32(val), true
	case string:
		if f, err := strconv.ParseFloat(val, 32); err == nil {
			return float32(f), true
		}
	}
	return 0, false
}