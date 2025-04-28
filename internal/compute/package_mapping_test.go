package compute

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPackageMapping(t *testing.T) {
	tests := []struct {
		name        string
		packageSize string
		wantExists  bool
		wantMemory  int
		wantDisk    int
		wantCPU     int
	}{
		{
			name:        "s-1vcpu-1gb package exists",
			packageSize: "s-1vcpu-1gb",
			wantExists:  true,
			wantMemory:  1024,
			wantDisk:    25,
			wantCPU:     1,
		},
		{
			name:        "s-2vcpu-2gb package exists",
			packageSize: "s-2vcpu-2gb",
			wantExists:  true,
			wantMemory:  2048,
			wantDisk:    50,
			wantCPU:     2,
		},
		{
			name:        "s-4vcpu-4gb package exists",
			packageSize: "s-4vcpu-4gb",
			wantExists:  true,
			wantMemory:  4096,
			wantDisk:    80,
			wantCPU:     4,
		},
		{
			name:        "s-8vcpu-8gb package exists",
			packageSize: "s-8vcpu-8gb",
			wantExists:  true,
			wantMemory:  8192,
			wantDisk:    160,
			wantCPU:     8,
		},
		{
			name:        "non-existent package",
			packageSize: "invalid-package",
			wantExists:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping, exists := GetPackageMapping(tt.packageSize)
			assert.Equal(t, tt.wantExists, exists)

			if tt.wantExists {
				assert.Equal(t, tt.wantMemory, mapping.Memory)
				assert.Equal(t, tt.wantDisk, mapping.Disk)
				assert.Equal(t, tt.wantCPU, mapping.CPU)
			}
		})
	}
}
