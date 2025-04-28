package compute

import (
	"testing"

	"github.com/celestiaorg/talis/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestCalculateTotalDiskSize(t *testing.T) {
	tests := []struct {
		name      string
		baseSize  int
		volumes   []types.Volume
		wantTotal int
	}{
		{
			name:      "no volumes",
			baseSize:  50,
			volumes:   nil,
			wantTotal: 50,
		},
		{
			name:     "single volume",
			baseSize: 50,
			volumes: []types.Volume{
				{
					Name:       "data",
					SizeGB:     20,
					MountPoint: "/mnt/data",
				},
			},
			wantTotal: 70,
		},
		{
			name:     "multiple volumes",
			baseSize: 50,
			volumes: []types.Volume{
				{
					Name:       "data1",
					SizeGB:     20,
					MountPoint: "/mnt/data1",
				},
				{
					Name:       "data2",
					SizeGB:     30,
					MountPoint: "/mnt/data2",
				},
			},
			wantTotal: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total := CalculateTotalDiskSize(tt.baseSize, tt.volumes)
			assert.Equal(t, tt.wantTotal, total)
		})
	}
}

func TestValidateVolumes(t *testing.T) {
	tests := []struct {
		name       string
		volumes    []types.Volume
		wantErrMsg string
	}{
		{
			name: "valid volumes",
			volumes: []types.Volume{
				{
					Name:       "data1",
					SizeGB:     20,
					MountPoint: "/mnt/data1",
				},
				{
					Name:       "data2",
					SizeGB:     30,
					MountPoint: "/mnt/data2",
				},
			},
			wantErrMsg: "",
		},
		{
			name: "empty mount point",
			volumes: []types.Volume{
				{
					Name:       "data",
					SizeGB:     20,
					MountPoint: "",
				},
			},
			wantErrMsg: "volume mount point cannot be empty",
		},
		{
			name: "invalid size",
			volumes: []types.Volume{
				{
					Name:       "data",
					SizeGB:     0,
					MountPoint: "/mnt/data",
				},
			},
			wantErrMsg: "volume size must be positive",
		},
		{
			name: "duplicate mount points",
			volumes: []types.Volume{
				{
					Name:       "data1",
					SizeGB:     20,
					MountPoint: "/mnt/data",
				},
				{
					Name:       "data2",
					SizeGB:     30,
					MountPoint: "/mnt/data",
				},
			},
			wantErrMsg: "duplicate mount point: /mnt/data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVolumes(tt.volumes)
			if tt.wantErrMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErrMsg, err.Error())
			}
		})
	}
}
