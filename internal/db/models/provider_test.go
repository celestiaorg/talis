package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderID(t *testing.T) {
	tests := []struct {
		name          string
		provider      ProviderID
		stringValue   string
		jsonValue     string
		validForParse bool
		validForJSON  bool
	}{
		{
			name:          "AWS provider",
			provider:      ProviderAWS,
			stringValue:   "aws",
			jsonValue:     `"aws"`,
			validForParse: true,
			validForJSON:  true,
		},
		{
			name:          "GCP provider",
			provider:      ProviderGCP,
			stringValue:   "gcp",
			jsonValue:     `"gcp"`,
			validForParse: true,
			validForJSON:  true,
		},
		{
			name:          "Azure provider",
			provider:      ProviderAzure,
			stringValue:   "azure",
			jsonValue:     `"azure"`,
			validForParse: true,
			validForJSON:  true,
		},
		{
			name:          "DigitalOcean provider",
			provider:      ProviderDO,
			stringValue:   "do",
			jsonValue:     `"do"`,
			validForParse: true,
			validForJSON:  true,
		},
		{
			name:          "Scaleway provider",
			provider:      ProviderScaleway,
			stringValue:   "scw",
			jsonValue:     `"scw"`,
			validForParse: true,
			validForJSON:  true,
		},
		{
			name:          "Vultr provider",
			provider:      ProviderVultr,
			stringValue:   "vultr",
			jsonValue:     `"vultr"`,
			validForParse: true,
			validForJSON:  true,
		},
		{
			name:          "Linode provider",
			provider:      ProviderLinode,
			stringValue:   "linode",
			jsonValue:     `"linode"`,
			validForParse: true,
			validForJSON:  true,
		},
		{
			name:          "Hetzner provider",
			provider:      ProviderHetzner,
			stringValue:   "hetzner",
			jsonValue:     `"hetzner"`,
			validForParse: true,
			validForJSON:  true,
		},
		{
			name:          "OVH provider",
			provider:      ProviderOVH,
			stringValue:   "ovh",
			jsonValue:     `"ovh"`,
			validForParse: true,
			validForJSON:  true,
		},
		{
			name:          "Empty provider",
			provider:      "",
			stringValue:   "",
			jsonValue:     `""`,
			validForParse: true,
			validForJSON:  true,
		},
		{
			name:          "Invalid JSON",
			jsonValue:     `invalid`,
			validForParse: false,
			validForJSON:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test String() method
			assert.Equal(t, tt.stringValue, tt.provider.String(), "String() method failed")

			// Test JSON marshaling
			if tt.validForJSON {
				bytes, err := tt.provider.MarshalJSON()
				assert.NoError(t, err, "Marshal should not return error")
				assert.Equal(t, tt.jsonValue, string(bytes), "Marshal produced incorrect JSON")
			}

			// Test JSON unmarshaling
			var unmarshaledProvider ProviderID
			err := unmarshaledProvider.UnmarshalJSON([]byte(tt.jsonValue))
			if tt.validForJSON {
				assert.NoError(t, err, "Unmarshal should not return error")
				assert.Equal(t, tt.provider, unmarshaledProvider, "Unmarshal produced incorrect provider")
			} else {
				assert.Error(t, err, "Unmarshal should return error for invalid JSON")
			}

			// Test direct string conversion (since ProviderID is a string type)
			assert.Equal(t, string(tt.provider), tt.stringValue, "Direct string conversion failed")
		})
	}
}

func TestProviderID_Validation(t *testing.T) {
	t.Run("Provider constants are unique", func(t *testing.T) {
		// Create a map to track unique values
		seen := make(map[string]bool)
		providers := []ProviderID{
			ProviderAWS,
			ProviderGCP,
			ProviderAzure,
			ProviderDO,
			ProviderScaleway,
			ProviderVultr,
			ProviderLinode,
			ProviderHetzner,
			ProviderOVH,
		}

		// Check that each provider has a unique value
		for _, provider := range providers {
			value := string(provider)
			if seen[value] {
				t.Errorf("Duplicate provider value found: %s", value)
			}
			seen[value] = true
		}
	})

	t.Run("Provider values are lowercase", func(t *testing.T) {
		providers := []ProviderID{
			ProviderAWS,
			ProviderGCP,
			ProviderAzure,
			ProviderDO,
			ProviderScaleway,
			ProviderVultr,
			ProviderLinode,
			ProviderHetzner,
			ProviderOVH,
		}

		for _, provider := range providers {
			value := string(provider)
			assert.Equal(t, value, provider.String(), "Provider string value should match direct string conversion")
		}
	})
}
