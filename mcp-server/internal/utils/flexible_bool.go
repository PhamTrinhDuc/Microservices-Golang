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
