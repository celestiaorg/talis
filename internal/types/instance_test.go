package types

import (
	"strings"
	"testing"
)

// TestInstancesRequest_Validate tests the Validate method for InstancesRequest. It does not test the Validate method for InstanceRequest.
// TestInstancesRequest_Validate tests the Validate method for InstancesRequest. It does not test the Validate method for InstanceRequest.
func TestInstancesRequest_Validate(t *testing.T) {
	// create a default valid InstanceRequest
	defaultInstanceRequest := InstanceRequest{
		Provider:          "do",
		NumberOfInstances: 1,
		Region:            "nyc1",
		Size:              "s-1vcpu-1gb",
		Image:             "ubuntu-20-04-x64",
		SSHKeyName:        "test-key",
		Volumes: []VolumeConfig{
			{
				Name:       "test-volume",
				SizeGB:     10,
				MountPoint: "/mnt/data",
			},
		},
	}
	tests := []struct {
		name    string
		request *InstancesRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid_request_with_instance_name",
			request: &InstancesRequest{
				ProjectName: "test-project",
				Instances: []InstanceRequest{
					{
						Name:              "valid-instance",
						Provider:          "do",
						NumberOfInstances: 1,
						Region:            "nyc1",
						Size:              "s-1vcpu-1gb",
						Image:             "ubuntu-20-04-x64",
						SSHKeyName:        "test-key",
						Volumes: []VolumeConfig{
							{
								Name:       "test-volume",
								SizeGB:     10,
								MountPoint: "/mnt/data",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid_request_with_multiple_instances",
			request: &InstancesRequest{
				ProjectName: "test-project",
				Instances: []InstanceRequest{
					{
						Name:              "instance-1",
						Provider:          "do",
						NumberOfInstances: 1,
						Region:            "nyc1",
						Size:              "s-1vcpu-1gb",
						Image:             "ubuntu-20-04-x64",
						SSHKeyName:        "test-key",
						Volumes: []VolumeConfig{
							{
								Name:       "test-volume",
								SizeGB:     10,
								MountPoint: "/mnt/data",
							},
						},
					},
					{
						Name:              "instance-2",
						Provider:          "do",
						NumberOfInstances: 1,
						Region:            "nyc1",
						Size:              "s-1vcpu-1gb",
						Image:             "ubuntu-20-04-x64",
						SSHKeyName:        "test-key",
						Volumes: []VolumeConfig{
							{
								Name:       "test-volume",
								SizeGB:     10,
								MountPoint: "/mnt/data",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid_request_using_InstanceName",
			request: &InstancesRequest{
				ProjectName:  "test-project",
				InstanceName: "valid-instance",
				Instances:    []InstanceRequest{defaultInstanceRequest, defaultInstanceRequest},
			},
			wantErr: false,
		},
		{
			name: "missing project name",
			request: &InstancesRequest{
				Instances: []InstanceRequest{
					{
						Name:              "valid-instance",
						Provider:          "do",
						NumberOfInstances: 1,
						Region:            "nyc1",
						Size:              "s-1vcpu-1gb",
						Image:             "ubuntu-20-04-x64",
						SSHKeyName:        "test-key",
						Volumes: []VolumeConfig{
							{
								Name:       "test-volume",
								SizeGB:     10,
								MountPoint: "/mnt/data",
							},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "project_name is required",
		},
		{
			name: "missing instances",
			request: &InstancesRequest{
				ProjectName: "test-project",
			},
			wantErr: true,
			errMsg:  "at least one instance configuration is required",
		},
		{
			name: "missing instance name and instance name in request",
			request: &InstancesRequest{
				ProjectName: "test-project",
				Instances: []InstanceRequest{
					{
						Provider:          "do",
						NumberOfInstances: 1,
						Region:            "nyc1",
						Size:              "s-1vcpu-1gb",
						Image:             "ubuntu-20-04-x64",
						SSHKeyName:        "test-key",
						Volumes: []VolumeConfig{
							{
								Name:       "test-volume",
								SizeGB:     10,
								MountPoint: "/mnt/data",
							},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "instance_name or instance.name is required",
		},
		{
			name: "empty instances array",
			request: &InstancesRequest{
				ProjectName: "test-project",
				Instances:   []InstanceRequest{},
			},
			wantErr: true,
			errMsg:  "at least one instance configuration is required",
		},
		{
			name: "invalid hostname",
			request: &InstancesRequest{
				TaskName:     "test-job",
				ProjectName:  "test-project",
				InstanceName: "invalid_hostname$123",
				Instances:    []InstanceRequest{defaultInstanceRequest},
			},
			wantErr: true,
			errMsg:  "invalid hostname format",
		},
		{
			name: "missing instance name and instance name in request",
			request: &InstancesRequest{
				TaskName:    "test-job",
				ProjectName: "test-project",
				Instances:   []InstanceRequest{defaultInstanceRequest},
			},
			wantErr: true,
			errMsg:  "instance_name or instance.name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("InstancesRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("InstancesRequest.Validate() error message = %v, want to contain %v", err, tt.errMsg)
			}
		})
	}
}

func TestInstanceRequest_Validate(t *testing.T) {
	defaultVolumeConfig := VolumeConfig{
		Name:       "test-volume",
		SizeGB:     10,
		MountPoint: "/mnt/data",
	}
	tests := []struct {
		name    string
		request *InstanceRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: &InstanceRequest{
				Name:              "valid-instance",
				Provider:          "do",
				NumberOfInstances: 1,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu-20-04-x64",
				SSHKeyName:        "test-key",
				Volumes:           []VolumeConfig{defaultVolumeConfig},
			},
			wantErr: false,
		},
		{
			name: "missing provider",
			request: &InstanceRequest{
				Name:              "valid-instance",
				NumberOfInstances: 1,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu-20-04-x64",
				SSHKeyName:        "test-key",
				Volumes:           []VolumeConfig{defaultVolumeConfig},
			},
			wantErr: true,
			errMsg:  "provider is required",
		},
		{
			name: "invalid number of instances",
			request: &InstanceRequest{
				Name:              "valid-instance",
				Provider:          "do",
				NumberOfInstances: 0,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu-20-04-x64",
				SSHKeyName:        "test-key",
				Volumes:           []VolumeConfig{defaultVolumeConfig},
			},
			wantErr: true,
			errMsg:  "number_of_instances must be greater than 0",
		},
		{
			name: "missing region",
			request: &InstanceRequest{
				Name:              "valid-instance",
				Provider:          "do",
				NumberOfInstances: 1,
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu-20-04-x64",
				SSHKeyName:        "test-key",
				Volumes:           []VolumeConfig{defaultVolumeConfig},
			},
			wantErr: true,
			errMsg:  "region is required",
		},
		{
			name: "missing size",
			request: &InstanceRequest{
				Name:              "valid-instance",
				Provider:          "do",
				NumberOfInstances: 1,
				Region:            "nyc1",
				Image:             "ubuntu-20-04-x64",
				SSHKeyName:        "test-key",
				Volumes:           []VolumeConfig{defaultVolumeConfig},
			},
			wantErr: true,
			errMsg:  "size is required",
		},
		{
			name: "missing image",
			request: &InstanceRequest{
				Name:              "valid-instance",
				Provider:          "do",
				NumberOfInstances: 1,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
				SSHKeyName:        "test-key",
				Volumes:           []VolumeConfig{defaultVolumeConfig},
			},
			wantErr: true,
			errMsg:  "image is required",
		},
		{
			name: "missing ssh key name",
			request: &InstanceRequest{
				Name:              "valid-instance",
				Provider:          "do",
				NumberOfInstances: 1,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu-20-04-x64",
				Volumes:           []VolumeConfig{defaultVolumeConfig},
			},
			wantErr: true,
			errMsg:  "ssh_key_name is required",
		},
		{
			name: "missing volumes",
			request: &InstanceRequest{
				Name:              "valid-instance",
				Provider:          "do",
				NumberOfInstances: 1,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu-20-04-x64",
				SSHKeyName:        "test-key",
			},
			wantErr: true,
			errMsg:  "at least one volume configuration is required",
		},
		{
			name: "invalid instance name",
			request: &InstanceRequest{
				Name:              "invalid_hostname$123",
				Provider:          "do",
				NumberOfInstances: 1,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu-20-04-x64",
				SSHKeyName:        "test-key",
				Volumes:           []VolumeConfig{defaultVolumeConfig},
			},
			wantErr: true,
			errMsg:  "invalid instance name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("InstanceRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("InstanceRequest.Validate() error message = %v, want to contain %v", err, tt.errMsg)
			}
		})
	}
}

func TestDeleteRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request *DeleteRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: &DeleteRequest{
				InstanceName: "test-instance",
				ProjectName:  "test-project",
				Instances: []DeleteInstance{
					{
						Provider:          "do",
						Name:              "instance-1",
						NumberOfInstances: 1,
						Region:            "nyc1",
						Size:              "s-1vcpu-1gb",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing instance name",
			request: &DeleteRequest{
				ProjectName: "test-project",
				Instances: []DeleteInstance{
					{
						Provider:          "do",
						NumberOfInstances: 1,
						Region:            "nyc1",
						Size:              "s-1vcpu-1gb",
					},
				},
			},
			wantErr: true,
			errMsg:  "instance_name is required",
		},
		{
			name: "missing project name",
			request: &DeleteRequest{
				InstanceName: "test-instance",
				Instances: []DeleteInstance{
					{
						Provider:          "do",
						NumberOfInstances: 1,
						Region:            "nyc1",
						Size:              "s-1vcpu-1gb",
					},
				},
			},
			wantErr: true,
			errMsg:  "project_name is required",
		},
		{
			name: "empty instances array",
			request: &DeleteRequest{
				InstanceName: "test-instance",
				ProjectName:  "test-project",
				Instances:    []DeleteInstance{},
			},
			wantErr: true,
			errMsg:  "at least one instance configuration is required",
		},
		{
			name: "invalid instance configuration",
			request: &DeleteRequest{
				InstanceName: "test-instance",
				ProjectName:  "test-project",
				Instances: []DeleteInstance{
					{
						NumberOfInstances: 1,
						Region:            "nyc1",
						Size:              "s-1vcpu-1gb",
					},
				},
			},
			wantErr: true,
			errMsg:  "provider is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("DeleteRequest.Validate() error message = %v, want to contain %v", err, tt.errMsg)
			}
		})
	}
}

func TestDeleteInstance_Validate(t *testing.T) {
	tests := []struct {
		name     string
		instance *DeleteInstance
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid instance",
			instance: &DeleteInstance{
				Provider:          "do",
				Name:              "instance-1",
				NumberOfInstances: 1,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
			},
			wantErr: false,
		},
		{
			name: "missing provider",
			instance: &DeleteInstance{
				Name:              "instance-1",
				NumberOfInstances: 1,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
			},
			wantErr: true,
			errMsg:  "provider is required",
		},
		{
			name: "invalid number of instances",
			instance: &DeleteInstance{
				Provider:          "do",
				Name:              "instance-1",
				NumberOfInstances: 0,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
			},
			wantErr: true,
			errMsg:  "number_of_instances must be greater than 0",
		},
		{
			name: "missing region",
			instance: &DeleteInstance{
				Provider:          "do",
				Name:              "instance-1",
				NumberOfInstances: 1,
				Size:              "s-1vcpu-1gb",
			},
			wantErr: true,
			errMsg:  "region is required",
		},
		{
			name: "missing size",
			instance: &DeleteInstance{
				Provider:          "do",
				Name:              "instance-1",
				NumberOfInstances: 1,
				Region:            "nyc1",
			},
			wantErr: true,
			errMsg:  "size is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.instance.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteInstance.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("DeleteInstance.Validate() error message = %v, want to contain %v", err, tt.errMsg)
			}
		})
	}
}
