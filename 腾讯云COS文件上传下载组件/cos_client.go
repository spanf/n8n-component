package cosclient

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/tencentyun/cos-go-sdk-v5"
)

func InitCOSClient(secretID string, secretKey string, region string) (*cos.Client, error) {
	u, _ := url.Parse("https://service.cos.myqcloud.com")
	baseTransport := createTransport()
	transport := &cos.AuthorizationTransport{
		SecretID:  secretID,
		SecretKey: secretKey,
		Transport: baseTransport,
	}
	client := cos.NewClient(&cos.BaseURL{ServiceURL: u}, &http.Client{
		Transport: transport,
	})
	return client, nil
}

func getBaseURL(bucketName string, region string) string {
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucketName, region)
}

func createTransport() *cos.BaseTransport {
	return &cos.BaseTransport{
		Transport: &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
		},
	}
}
