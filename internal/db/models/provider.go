package models

import "encoding/json"

// ProviderID represents a unique identifier for a cloud provider
type ProviderID string

// Provider constants define the supported cloud providers
const (
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
	// ProviderXimera represents Ximera provider
	ProviderXimera ProviderID = "ximera"

	// Mock Providers
	// ProviderDOMock1 represents DigitalOcean provider mock 1
	ProviderDOMock1 ProviderID = "do-mock"
	// ProviderDOMock2 represents DigitalOcean provider mock 2
	ProviderDOMock2 ProviderID = "digitalocean-mock"
	// ProviderDOMock3 represents a mock provider
	ProviderMock3 ProviderID = "mock"
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

// FromString creates a ProviderID from a string
func FromString(s string) ProviderID {
	return ProviderID(s)
}

// ToProviderID converts a string to a ProviderID
func ToProviderID(s string) ProviderID {
	return ProviderID(s)
}

// IsValid checks if the provider ID is a valid supported provider
func (p ProviderID) IsValid() bool {
	switch p {
	case ProviderAWS, ProviderGCP, ProviderAzure, ProviderDO,
		ProviderScaleway, ProviderVultr, ProviderLinode, ProviderHetzner, ProviderOVH,
		ProviderXimera:
		return true
	case ProviderDOMock1, ProviderDOMock2, ProviderMock3: // mocked providers
		return true
	default:
		return false
	}
}
