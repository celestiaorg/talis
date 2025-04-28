package models

import "encoding/json"

// ProviderID represents a unique identifier for a cloud provider
type ProviderID string

// Provider constants define the supported cloud providers
const (
	// ProviderVirtFusion represents VirtFusion provider
	ProviderVirtFusion ProviderID = "virtfusion"

	// Mock Providers
	// ProviderMock represents a mock provider for testing
	ProviderMock ProviderID = "mock"
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
	case ProviderVirtFusion:
		return true
	case ProviderMock: // mocked provider
		return true
	default:
		return false
	}
}
