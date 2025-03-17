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
