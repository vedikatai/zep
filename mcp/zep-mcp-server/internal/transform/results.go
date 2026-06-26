package transform

import (
	"encoding/json"
	"fmt"
)

// FormatJSON formats a value as pretty-printed JSON
func FormatJSON(v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(data), nil
}

// TODO: verify if dead code, consider removing
// ToJSONString converts a value to a compact JSON string
func ToJSONString(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return string(data)
}
