package types

import (
	"context"

	"github.com/digitalocean/godo"
)

// DropletService defines the interface for DigitalOcean droplet operations
type DropletService interface {
	Create(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error)
	CreateMultiple(ctx context.Context, createRequest *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error)
	Delete(ctx context.Context, dropletID int) (*godo.Response, error)
	Get(ctx context.Context, dropletID int) (*godo.Droplet, *godo.Response, error)
	List(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error)
}

// KeyService defines the interface for DigitalOcean SSH key operations
type KeyService interface {
	List(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error)
}

// StorageService defines the interface for DigitalOcean volume operations
type StorageService interface {
	CreateVolume(ctx context.Context, request *godo.VolumeCreateRequest) (*godo.Volume, *godo.Response, error)
	DeleteVolume(ctx context.Context, id string) (*godo.Response, error)
	ListVolumes(ctx context.Context, opt *godo.ListVolumeParams) ([]godo.Volume, *godo.Response, error)
	GetVolume(ctx context.Context, id string) (*godo.Volume, *godo.Response, error)
	GetVolumeAction(ctx context.Context, volumeID string, actionID int) (*godo.Action, *godo.Response, error)
	AttachVolume(ctx context.Context, volumeID string, dropletID int) (*godo.Response, error)
	DetachVolume(ctx context.Context, volumeID string, dropletID int) (*godo.Response, error)
}
