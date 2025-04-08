package types

import (
	"strings"
	"testing"
)

func TestJobRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request *JobRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: &JobRequest{
				Name: "test-job",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			request: &JobRequest{
				Name: "",
			},
			wantErr: true,
			errMsg:  "job_name is required",
		},
		{
			name:    "nil request",
			request: nil,
			wantErr: true,
			errMsg:  "job_name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.request != nil {
				err = tt.request.Validate()
			} else {
				err = (&JobRequest{}).Validate()
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("JobRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("JobRequest.Validate() error message = %v, want to contain %v", err, tt.errMsg)
			}
		})
	}
}
