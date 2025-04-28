// Package types provides interface definitions for compute providers
package types

import (
	"context"

	"github.com/celestiaorg/talis/pkg/types"
	"github.com/digitalocean/godo"
)

// DOClient defines the interface for Digital Ocean client operations
type DOClient interface {
	Droplets() DropletService
	Keys() KeyService
	Storage() StorageService
	ValidateCredentials() error
	GetEnvironmentVars() map[string]string
	ConfigureProvider(stack interface{}) error
	CreateInstance(ctx context.Context, config *types.InstanceRequest) error
	DeleteInstance(ctx context.Context, dropletID int) error
}

// DropletService defines the interface for droplet operations
type DropletService interface {
	Create(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error)
	Get(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error)
	Delete(ctx context.Context, id int) (*godo.Response, error)
	List(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error)
}

// KeyService defines the interface for SSH key operations
type KeyService interface {
	List(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error)
}

// StorageService is the interface for volume operations
type StorageService interface {
	CreateVolume(ctx context.Context, request *godo.VolumeCreateRequest) (*godo.Volume, *godo.Response, error)
	DeleteVolume(ctx context.Context, id string) (*godo.Response, error)
	ListVolumes(ctx context.Context, opt *godo.ListVolumeParams) ([]godo.Volume, *godo.Response, error)
	GetVolume(ctx context.Context, id string) (*godo.Volume, *godo.Response, error)
	GetVolumeAction(ctx context.Context, volumeID string, actionID int) (*godo.Action, *godo.Response, error)
	AttachVolume(ctx context.Context, volumeID string, dropletID int) (*godo.Response, error)
	DetachVolume(ctx context.Context, volumeID string, dropletID int) (*godo.Response, error)
}
