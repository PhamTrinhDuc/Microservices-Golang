package redis

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/adk/session"
)

func TestExtractStateDeltas(t *testing.T) {
	countFuncTest += 1
	tests := []struct {
		name                 string
		state                map[string]any
		expectedAppDelta     map[string]any
		expectedUserDelta    map[string]any
		expectedSessionDelta map[string]any
	}{
		{
			name:                 getName("nil state"),
			state:                nil,
			expectedAppDelta:     nil,
			expectedUserDelta:    nil,
			expectedSessionDelta: nil,
		},
		{
			name:                 getName("empty state"),
			state:                map[string]any{},
			expectedAppDelta:     map[string]any{},
			expectedUserDelta:    map[string]any{},
			expectedSessionDelta: map[string]any{},
		},
		{
			name: getName("mixed state"),
			state: map[string]any{
				session.KeyPrefixApp + "theme": "dark",
				session.KeyPrefixUser + "name": "Jiyuu",
				"counter":                      123,
			},
			expectedAppDelta: map[string]any{
				"theme": "dark",
			},
			expectedUserDelta: map[string]any{
				"name": "Jiyuu",
			},
			expectedSessionDelta: map[string]any{
				"counter": 123,
			},
		},
		{
			name: getName("only app state"),
			state: map[string]any{
				session.KeyPrefixApp + "lang": "en",
			},
			expectedAppDelta: map[string]any{
				"lang": "en",
			},
			expectedUserDelta:    map[string]any{},
			expectedSessionDelta: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appDelta, userDelta, sessionDelta := extractStateDeltas(tt.state)
			assert.Equal(t, tt.expectedAppDelta, appDelta)
			assert.Equal(t, tt.expectedUserDelta, userDelta)
			assert.Equal(t, tt.expectedSessionDelta, sessionDelta)
		})
	}
}

func TestUnmarshalHashFields(t *testing.T) {
	countFuncTest += 1
	tests := []struct {
		name     string
		data     map[string]string
		expected map[string]any
	}{
		{
			name:     getName("empty data"),
			data:     map[string]string{},
			expected: map[string]any{},
		},
		{
			name: getName("various types"),
			data: map[string]string{
				"string": "\"hello\"",
				"int":    "123",
				"bool":   "true",
				"float":  "12.34",
				"map":    "{\"key\":\"value\"}",
				"array":  "[1, 2, 3]",
			},
			expected: map[string]any{
				"string": "hello",
				"int":    float64(123), // JSON unmarshal into any results in float64 for numbers
				"bool":   true,
				"float":  12.34,
				"map":    map[string]any{"key": "value"},
				"array":  []any{float64(1), float64(2), float64(3)},
			},
		},
		{
			name: getName("invalid json"),
			data: map[string]string{
				"invalid": "{not-json}",
			},
			expected: map[string]any{
				"invalid": "{not-json}",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unmarshalHashFields(tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMarshalHashFields(t *testing.T) {
	countFuncTest += 1
	tests := []struct {
		name     string
		data     map[string]any
		expected map[string]string
	}{
		{
			name:     getName("empty data"),
			data:     map[string]any{},
			expected: map[string]string{},
		},
		{
			name: getName("various types"),
			data: map[string]any{
				"string": "hello",
				"int":    123,
				"bool":   true,
				"float":  12.34,
				"map":    map[string]any{"key": "value"},
				"array":  []int{1, 2, 3},
			},
			expected: map[string]string{
				"string": "\"hello\"",
				"int":    "123",
				"bool":   "true",
				"float":  "12.34",
				"map":    "{\"key\":\"value\"}",
				"array":  "[1,2,3]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := marshalHashFields(tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}
