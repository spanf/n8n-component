package upload_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/yourproject/upload"
)

type mockS3Client struct {
	s3iface.S3API
	putObjectFunc func(input *s3.PutObjectInput) (*s3.PutObjectOutput, error)
}

func (m *mockS3Client) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	return m.putObjectFunc(input)
}

func TestValidateBucketName(t *testing.T) {
	tests := []struct {
		name     string
		bucket   string
		hasError bool
	}{
		{"ValidBucket", "my-bucket123", false},
		{"EmptyBucket", "", true},
		{"TooShort", "ab", true},
		{"TooLong", strings.Repeat("a", 64), true},
		{"InvalidStartChar", "-bucket", true},
		{"InvalidEndChar", "bucket-", true},
		{"InvalidChars", "my_bucket", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := upload.ValidateBucketName(tt.bucket)
			if (err != nil) != tt.hasError {
				t.Errorf("validateBucketName(%q) error = %v, wantErr %v", tt.bucket, err, tt.hasError)
			}
		})
	}
}

func TestUploadFile(t *testing.T) {
	mockFile := bytes.NewReader([]byte("test file content"))

	t.Run("InvalidBucketName", func(t *testing.T) {
		mockClient := &mockS3Client{
			putObjectFunc: func(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
				t.Fatal("PutObject should not be called for invalid bucket")
				return nil, nil
			},
		}

		err := upload.UploadFile(mockClient, "invalid.bucket", "key", mockFile)
		if err == nil {
			t.Fatal("Expected error for invalid bucket name")
		}
	})

	t.Run("S3PutObjectError", func(t *testing.T) {
		expectedErr := awserr.New("TestError", "test error", nil)
		mockClient := &mockS3Client{
			putObjectFunc: func(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
				return nil, expectedErr
			},
		}

		err := upload.UploadFile(mockClient, "valid-bucket", "key", mockFile)
		if err != expectedErr {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}
	})

	t.Run("NonAwsError", func(t *testing.T) {
		expectedErr := errors.New("generic error")
		mockClient := &mockS3Client{
			putObjectFunc: func(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
				return nil, expectedErr
			},
		}

		err := upload.UploadFile(mockClient, "valid-bucket", "key", mockFile)
		if err != expectedErr {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		mockClient := &mockS3Client{
			putObjectFunc: func(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
				if aws.StringValue(input.Bucket) != "valid-bucket" {
					t.Errorf("Expected bucket 'valid-bucket', got %s", aws.StringValue(input.Bucket))
				}
				if aws.StringValue(input.Key) != "key" {
					t.Errorf("Expected key 'key', got %s", aws.StringValue(input.Key))
				}
				return &s3.PutObjectOutput{}, nil
			},
		}

		err := upload.UploadFile(mockClient, "valid-bucket", "key", mockFile)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}

func TestUploadFile_EmptyReader(t *testing.T) {
	mockClient := &mockS3Client{
		putObjectFunc: func(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
			if input.Body == nil {
				t.Error("Expected non-nil body")
			}
			return &s3.PutObjectOutput{}, nil
		},
	}

	emptyReader := bytes.NewReader([]byte{})
	err := upload.UploadFile(mockClient, "valid-bucket", "key", emptyReader)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestUploadFile_NilReader(t *testing.T) {
	mockClient := &mockS3Client{
		putObjectFunc: func(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
			if input.Body != nil {
				t.Error("Expected nil body")
			}
			return &s3.PutObjectOutput{}, nil
		},
	}

	err := upload.UploadFile(mockClient, "valid-bucket", "key", nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
