package cosclient

import (
	"net/http"
	"net/url"

	"github.com/tencentyun/cos-go-sdk-v5"
)

func NewCOSClient(secretID, secretKey, region, bucket string) (*cos.Client, error) {
	u, err := url.Parse("https://" + bucket + ".cos." + region + ".myqcloud.com")
	if err != nil {
		return nil, err
	}

	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretID,
			SecretKey: secretKey,
		},
	})

	return client, nil
}
