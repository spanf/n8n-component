package utils

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPresignClient struct {
	mock.Mock
}

func (m *MockPresignClient) PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*aws.PresignedHTTPRequest, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*aws.PresignedHTTPRequest), args.Error(1)
}

type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return nil, nil
}

func TestGeneratePresignedURL_Success(t *testing.T) {
	mockPresign := new(MockPresignClient)
	mockPresign.On("PresignGetObject", mock.Anything, mock.Anything, mock.Anything).
		Return(&aws.PresignedHTTPRequest{URL: "https://example.com"}, nil)

	origNewPresign := newPresignClient
	newPresignClient = func(client *s3.S3) presignClient {
		return mockPresign
	}
	defer func() { newPresignClient = origNewPresign }()

	client := &s3.S3{}
	url, err := GeneratePresignedURL(client, "test-bucket", "test-key", 5*time.Minute)

	assert.NoError(t, err)
	assert.Equal(t, "https://example.com", url)
	mockPresign.AssertExpectations(t)
}

func TestGeneratePresignedURL_S3Error(t *testing.T) {
	mockPresign := new(MockPresignClient)
	apiError := &smithy.GenericAPIError{Code: "NoSuchKey", Message: "Object not found"}
	mockPresign.On("PresignGetObject", mock.Anything, mock.Anything, mock.Anything).
		Return(&aws.PresignedHTTPRequest{}, apiError)

	origNewPresign := newPresignClient
	newPresignClient = func(client *s3.S3) presignClient {
		return mockPresign
	}
	defer func() { newPresignClient = origNewPresign }()

	client := &s3.S3{}
	_, err := GeneratePresignedURL(client, "test-bucket", "invalid-key", 5*time.Minute)

	assert.Error(t, err)
	s3Err, ok := err.(*S3Error)
	assert.True(t, ok)
	assert.Equal(t, "NoSuchKey", s3Err.Code)
	assert.Equal(t, "Object not found", s3Err.Message)
}

func TestGeneratePresignedURL_GenericError(t *testing.T) {
	mockPresign := new(MockPresignClient)
	genericError := errors.New("network failure")
	mockPresign.On("PresignGetObject", mock.Anything, mock.Anything, mock.Anything).
		Return(&aws.PresignedHTTPRequest{}, genericError)

	origNewPresign := newPresignClient
	newPresignClient = func(client *s3.S3) presignClient {
		return mockPresign
	}
	defer func() { newPresignClient = origNewPresign }()

	client := &s3.S3{}
	_, err := GeneratePresignedURL(client, "test-bucket", "test-key", 5*time.Minute)

	assert.Error(t, err)
	assert.Equal(t, "network failure", err.Error())
}

func TestParseS3Error_WithAPIError(t *testing.T) {
	apiError := &types.NoSuchKey{Message: aws.String("Object missing")}
	err := parseS3Error(apiError)

	s3Err, ok := err.(*S3Error)
	assert.True(t, ok)
	assert.Equal(t, "NoSuchKey", s3Err.Code)
	assert.Equal(t, "Object missing", s3Err.Message)
}

func TestParseS3Error_WithGenericAPIError(t *testing.T) {
	apiError := &smithy.GenericAPIError{Code: "AccessDenied", Message: "Permission denied"}
	err := parseS3Error(apiError)

	s3Err, ok := err.(*S3Error)
	assert.True(t, ok)
	assert.Equal(t, "AccessDenied", s3Err.Code)
	assert.Equal(t, "Permission denied", s3Err.Message)
}

func TestParseS3Error_WithGenericError(t *testing.T) {
	genericError := errors.New("generic error")
	err := parseS3Error(genericError)

	assert.Equal(t, genericError, err)
}

type presignClient interface {
	PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*aws.PresignedHTTPRequest, error)
}

var newPresignClient = func(client *s3.S3) presignClient {
	return s3.NewPresignClient(client)
}
