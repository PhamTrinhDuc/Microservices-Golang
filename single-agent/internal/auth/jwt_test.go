package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractTokenFromHeader(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		want       string
		wantErr    bool
	}{
		{
			name:       "valid header",
			authHeader: "Bearer eyJhbG...",
			want:       "eyJhbG...",
			wantErr:    false,
		},
		{
			name:       "empty header",
			authHeader: "",
			wantErr:    true,
		},
		{
			name:       "missing bearer",
			authHeader: "eyJhbG...",
			wantErr:    true,
		},
		{
			name:       "invalid format",
			authHeader: "Bearer ",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractTokenFromHeader(tt.authHeader)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
