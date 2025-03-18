package types

import (
	"context"

	"github.com/digitalocean/godo"
)

// DOClient defines the interface for Digital Ocean client operations
type DOClient interface {
	Droplets() DropletService
	Keys() KeyService
}

// DropletService defines the interface for droplet operations
type DropletService interface {
	Create(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error)
	CreateMultiple(ctx context.Context, createRequest *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error)
	Get(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error)
	Delete(ctx context.Context, id int) (*godo.Response, error)
	List(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error)
}

// KeyService defines the interface for SSH key operations
type KeyService interface {
	List(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error)
}
