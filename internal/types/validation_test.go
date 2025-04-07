package types

import (
	"strings"
	"testing"
)

func Test_validateHostname(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid hostname",
			hostname: "valid-hostname-123",
			wantErr:  false,
		},
		{
			name:     "valid hostname with numbers",
			hostname: "web1",
			wantErr:  false,
		},
		{
			name:     "valid hostname with hyphens",
			hostname: "my-web-server-01",
			wantErr:  false,
		},
		{
			name:     "empty hostname",
			hostname: "",
			wantErr:  true,
			errMsg:   "hostname cannot be empty",
		},
		{
			name:     "hostname too long",
			hostname: "this-hostname-is-way-too-long-and-should-fail-because-it-exceeds-sixty-three-chars",
			wantErr:  true,
			errMsg:   "hostname length must be less than or equal to 63 characters",
		},
		{
			name:     "hostname with invalid characters",
			hostname: "invalid_hostname$123",
			wantErr:  true,
			errMsg:   "invalid hostname format",
		},
		{
			name:     "hostname starting with hyphen",
			hostname: "-invalid-start",
			wantErr:  true,
			errMsg:   "invalid hostname format",
		},
		{
			name:     "hostname ending with hyphen",
			hostname: "invalid-end-",
			wantErr:  true,
			errMsg:   "invalid hostname format",
		},
		{
			name:     "hostname with uppercase letters",
			hostname: "UPPERCASE-HOST",
			wantErr:  false, // Should pass because we convert to lowercase
		},
		{
			name:     "hostname with spaces",
			hostname: "invalid hostname",
			wantErr:  true,
			errMsg:   "invalid hostname format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHostname(tt.hostname)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateHostname() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateHostname() error message = %v, want to contain %v", err, tt.errMsg)
			}
		})
	}
}
