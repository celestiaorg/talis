package mocks

import (
	"fmt"
	"time"

	"github.com/digitalocean/godo"
)

// Default test values for droplets
var (
	DefaultDropletID1    = 12345
	DefaultDropletID2    = 12346
	DefaultDropletName1  = "test-droplet-1"
	DefaultDropletName2  = "test-droplet-2"
	DefaultDropletIP1    = "192.0.2.1"
	DefaultDropletIP2    = "192.0.2.2"
	DefaultDropletRegion = "nyc1"
	DefaultDropletSize   = "s-1vcpu-1gb"
	DefaultDropletStatus = "active"

	DefaultDropletList = []struct {
		ID   int
		Name string
		IP   string
	}{
		{DefaultDropletID1, DefaultDropletName1, DefaultDropletIP1},
		{DefaultDropletID2, DefaultDropletName2, DefaultDropletIP2},
	}
)

// Default test values for SSH keys
var (
	DefaultKeyID1        = 67890
	DefaultKeyID2        = 67891
	DefaultKeyName1      = "test-key-1"
	DefaultKeyName2      = "test-key-2"
	DefaultKeyPublicKey1 = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC..."
	DefaultKeyPublicKey2 = "ssh-rsa BBBBB3NzaC1yc2EAAAADAQABAAABAQC..."

	DefaultKeyList = []struct {
		ID        int
		Name      string
		PublicKey string
	}{
		{DefaultKeyID1, DefaultKeyName1, DefaultKeyPublicKey1},
		{DefaultKeyID2, DefaultKeyName2, DefaultKeyPublicKey2},
	}
)

// Error messages
var (
	ErrDropletNotFound = fmt.Errorf("DO API: droplet not found")
	ErrKeyNotFound     = fmt.Errorf("DO API: SSH key not found")
	ErrRateLimit       = fmt.Errorf("DO API: rate limit exceeded")
	ErrAuthentication  = fmt.Errorf("DO API: authentication failed")
)

// StandardResponses contains all standard mock responses
type StandardResponses struct {
	Droplets StandardDropletResponses
	Keys     StandardKeyResponses
}

// StandardDropletResponses contains all standard mock responses for droplets
type StandardDropletResponses struct {
	// Single droplet responses
	DefaultDroplet *godo.Droplet

	// Multiple droplet responses
	DefaultDropletList []godo.Droplet

	// Error responses
	NotFoundError       error
	RateLimitError      error
	AuthenticationError error
}

// StandardKeyResponses contains all standard mock responses for keys
type StandardKeyResponses struct {
	// Success responses
	DefaultKey     *godo.Key
	DefaultKeyList []godo.Key

	// Error responses
	NotFoundError       error
	RateLimitError      error
	AuthenticationError error
}

// newStandardResponses creates a new set of standard responses
func newStandardResponses() *StandardResponses {
	return &StandardResponses{
		Droplets: StandardDropletResponses{
			DefaultDroplet: &godo.Droplet{
				ID:   DefaultDropletID1,
				Name: DefaultDropletName1,
				Networks: &godo.Networks{
					V4: []godo.NetworkV4{
						{
							Type:      "public",
							IPAddress: DefaultDropletIP1,
						},
					},
				},
				Region: &godo.Region{
					Slug: DefaultDropletRegion,
				},
				Size: &godo.Size{
					Slug: DefaultDropletSize,
				},
				Status:  DefaultDropletStatus,
				Created: time.Now().Format(time.RFC3339),
			},
			DefaultDropletList: []godo.Droplet{
				{
					ID:   DefaultDropletList[0].ID,
					Name: DefaultDropletList[0].Name,
					Networks: &godo.Networks{
						V4: []godo.NetworkV4{
							{
								Type:      "public",
								IPAddress: DefaultDropletList[0].IP,
							},
						},
					},
					Region: &godo.Region{Slug: DefaultDropletRegion},
					Size:   &godo.Size{Slug: DefaultDropletSize},
				},
				{
					ID:   DefaultDropletList[1].ID,
					Name: DefaultDropletList[1].Name,
					Networks: &godo.Networks{
						V4: []godo.NetworkV4{
							{
								Type:      "public",
								IPAddress: DefaultDropletList[1].IP,
							},
						},
					},
					Region: &godo.Region{Slug: DefaultDropletRegion},
					Size:   &godo.Size{Slug: DefaultDropletSize},
				},
			},
			NotFoundError:       ErrDropletNotFound,
			RateLimitError:      ErrRateLimit,
			AuthenticationError: ErrAuthentication,
		},
		Keys: StandardKeyResponses{
			DefaultKey: &godo.Key{
				ID:        DefaultKeyID1,
				Name:      DefaultKeyName1,
				PublicKey: DefaultKeyPublicKey1,
			},
			DefaultKeyList: []godo.Key{
				{
					ID:        DefaultKeyList[0].ID,
					Name:      DefaultKeyList[0].Name,
					PublicKey: DefaultKeyList[0].PublicKey,
				},
				{
					ID:        DefaultKeyList[1].ID,
					Name:      DefaultKeyList[1].Name,
					PublicKey: DefaultKeyList[1].PublicKey,
				},
			},
			NotFoundError:       ErrKeyNotFound,
			RateLimitError:      ErrRateLimit,
			AuthenticationError: ErrAuthentication,
		},
	}
}
