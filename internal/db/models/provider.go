package models

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
	ProviderDO       ProviderID = "do"
	ProviderScaleway ProviderID = "scw"
	ProviderVultr    ProviderID = "vultr"
	ProviderLinode   ProviderID = "linode"
	ProviderHetzner  ProviderID = "hetzner"
	ProviderOVH      ProviderID = "ovh"
)
