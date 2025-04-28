package compute

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTemplateMapping(t *testing.T) {
	tests := []struct {
		name       string
		imageName  string
		wantExists bool
		wantName   string
	}{
		{
			name:       "ubuntu-22.04 exists",
			imageName:  "ubuntu-22.04",
			wantExists: true,
			wantName:   "ubuntu-22.04-x64",
		},
		{
			name:       "ubuntu-20.04 exists",
			imageName:  "ubuntu-20.04",
			wantExists: true,
			wantName:   "ubuntu-20.04-x64",
		},
		{
			name:       "debian-11 exists",
			imageName:  "debian-11",
			wantExists: true,
			wantName:   "debian-11-x64",
		},
		{
			name:       "debian-10 exists",
			imageName:  "debian-10",
			wantExists: true,
			wantName:   "debian-10-x64",
		},
		{
			name:       "centos-7 exists",
			imageName:  "centos-7",
			wantExists: true,
			wantName:   "centos-7-x64",
		},
		{
			name:       "rocky-8 exists",
			imageName:  "rocky-8",
			wantExists: true,
			wantName:   "rocky-8-x64",
		},
		{
			name:       "rocky-9 exists",
			imageName:  "rocky-9",
			wantExists: true,
			wantName:   "rocky-9-x64",
		},
		{
			name:       "non-existent image",
			imageName:  "invalid-image",
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, exists := GetTemplateMapping(tt.imageName)
			assert.Equal(t, tt.wantExists, exists)

			if tt.wantExists {
				assert.Equal(t, tt.wantName, template)
			}
		})
	}
}
