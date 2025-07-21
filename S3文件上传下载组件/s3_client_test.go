package s3client

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
)

func TestValidateCredentials(t *testing.T) {
	tests := []struct {
		name        string
		accessKey   string
		secretKey   string
		expectedErr string
	}{
		{
			name:        "EmptyAccessKey",
			accessKey:   "",
			secretKey:   "validSecretKey12345",
			expectedErr: "accessKey cannot be empty",
		},
		{
			name:        "EmptySecretKey",
			accessKey:   "validAccessKey12345",
			secretKey:   "",
			expectedErr: "secretKey cannot be empty",
		},
		{
			name:        "ShortAccessKey",
			accessKey:   "short",
			secretKey:   "validSecretKey12345",
			expectedErr: "accessKey format invalid: minimum length 16 characters",
		},
		{
			name:        "ShortSecretKey",
			accessKey:   "validAccessKey12345",
			secretKey:   "short",
			expectedErr: "secretKey format invalid: minimum length 16 characters",
		},
		{
			name:        "ValidCredentials",
			accessKey:   "validAccessKey12345",
			secretKey:   "validSecretKey12345",
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCredentials(tt.accessKey, tt.secretKey)
			if tt.expectedErr == "" {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			} else {
				if err == nil || err.Error() != tt.expectedErr {
					t.Errorf("expected error %q, got %v", tt.expectedErr, err)
				}
			}
		})
	}
}

func TestNewS3Client(t *testing.T) {
	tests := []struct {
		name        string
		accessKey   string
		secretKey   string
		region      string
		expectError bool
		errMsg      string
	}{
		{
			name:        "InvalidCredentials",
			accessKey:   "short",
			secretKey:   "key",
			region:      "us-west-1",
			expectError: true,
			errMsg:      "accessKey format invalid: minimum length 16 characters",
		},
		{
			name:        "ValidParameters",
			accessKey:   "validAccessKey12345",
			secretKey:   "validSecretKey12345",
			region:      "us-east-1",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewS3Client(tt.accessKey, tt.secretKey, tt.region)
			
			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				if err.Error() != tt.errMsg {
					t.Errorf("expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if client == nil {
					t.Fatal("expected S3 client but got nil")
				}
				if _, ok := client.(*s3.S3); !ok {
					t.Errorf("returned object is not of type *s3.S3")
				}
			}
		})
	}
}
