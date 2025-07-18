package cosclient

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCOSClient_Success(t *testing.T) {
	secretID := "testSecretID"
	secretKey := "testSecretKey"
	region := "ap-guangzhou"
	bucket := "test-bucket-123"

	client, err := NewCOSClient(secretID, secretKey, region, bucket)

	assert.NoError(t, err)
	assert.NotNil(t, client)

	// 验证基础URL是否正确
	expectedURL, _ := url.Parse("https://test-bucket-123.cos.ap-guangzhou.myqcloud.com")
	actualURL := client.BaseURL.BucketURL
	assert.Equal(t, expectedURL.String(), actualURL.String())

	// 验证认证传输设置
	transport := client.Client.Transport.(*cos.AuthorizationTransport)
	assert.Equal(t, secretID, transport.SecretID)
	assert.Equal(t, secretKey, transport.SecretKey)
}

func TestNewCOSClient_InvalidURL(t *testing.T) {
	secretID := "testSecretID"
	secretKey := "testSecretKey"
	region := "invalid region!@#" // 非法字符
	bucket := "test-bucket"

	client, err := NewCOSClient(secretID, secretKey, region, bucket)

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "parse")
}

func TestNewCOSClient_EmptyParameters(t *testing.T) {
	tests := []struct {
		name      string
		secretID  string
		secretKey string
		region    string
		bucket    string
	}{
		{"Empty Bucket", "id", "key", "region", ""},
		{"Empty Region", "id", "key", "", "bucket"},
		{"Empty SecretID", "", "key", "region", "bucket"},
		{"Empty SecretKey", "id", "", "region", "bucket"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewCOSClient(tt.secretID, tt.secretKey, tt.region, tt.bucket)

			if tt.region == "" || tt.bucket == "" {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				// 仅验证凭据为空时client仍被创建
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}
