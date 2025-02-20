package models

// we keep it like this for the moment, but we will put them all in the database with some auth details
type ProviderID string

const (
	ProviderAWS          ProviderID = "aws"
	ProviderGCP          ProviderID = "gcp"
	ProviderAzure        ProviderID = "azure"
	ProviderDigitalOcean ProviderID = "do"
	ProviderScaleway     ProviderID = "scw"
	ProviderVultr        ProviderID = "vultr"
	ProviderLinode       ProviderID = "linode"
	ProviderHetzner      ProviderID = "hetzner"
	ProviderOVH          ProviderID = "ovh"
)
