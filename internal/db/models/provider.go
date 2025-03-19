package models

import "encoding/json"

// ProviderID represents a unique identifier for a cloud provider
type ProviderID string

// Provider constants define the supported cloud providers
const (
	// ProviderMock represents a mock provider for testing purposes
	ProviderMock ProviderID = "mock"
	// ProviderAWS represents Amazon Web Services provider
	ProviderAWS ProviderID = "aws"
	// ProviderGCP represents Google Cloud Platform provider
	ProviderGCP ProviderID = "gcp"
	// ProviderAzure represents Microsoft Azure provider
	ProviderAzure ProviderID = "azure"
	// ProviderDO represents DigitalOcean provider
	ProviderDO ProviderID = "do"
	// ProviderScaleway represents Scaleway provider
	ProviderScaleway ProviderID = "scw"
	// ProviderVultr represents Vultr provider
	ProviderVultr ProviderID = "vultr"
	// ProviderLinode represents Linode provider
	ProviderLinode ProviderID = "linode"
	// ProviderHetzner represents Hetzner provider
	ProviderHetzner ProviderID = "hetzner"
	// ProviderOVH represents OVH provider
	ProviderOVH ProviderID = "ovh"
)

// String implements the fmt.Stringer interface
func (p ProviderID) String() string {
	return string(p)
}

// MarshalJSON implements the json.Marshaler interface
func (p ProviderID) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (p *ProviderID) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	*p = ProviderID(str)
	return nil
}
