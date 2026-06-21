package utils

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FlexibleBool can unmarshal JSON booleans or strings (case-insensitive "true"/"false")
type FlexibleBool bool

// UnmarshalJSON implements json.Unmarshaler
func (fb *FlexibleBool) UnmarshalJSON(data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch val := v.(type) {
	case bool:
		*fb = FlexibleBool(val)
	case string:
		*fb = FlexibleBool(strings.ToLower(val) == "true")
	default:
		return fmt.Errorf("invalid type for FlexibleBool: %T", v)
	}
	return nil
}

// FlexibleString can unmarshal JSON strings, numbers, or other types into a string
type FlexibleString string

// UnmarshalJSON implements json.Unmarshaler
func (fs *FlexibleString) UnmarshalJSON(data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch val := v.(type) {
	case string:
		*fs = FlexibleString(val)
	case float64:
		*fs = FlexibleString(fmt.Sprintf("%.0f", val))
	case bool:
		*fs = FlexibleString(fmt.Sprintf("%v", val))
	default:
		*fs = FlexibleString(fmt.Sprintf("%v", val))
	}
	return nil
}

