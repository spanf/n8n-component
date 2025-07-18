package cosdownloader

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"

	"github.com/tencentyun/cos-go-sdk-v5"
)

type CosConfig struct {
	Endpoint   string
	SecretID   string
	SecretKey  string
	BucketName string
}

func DownloadFile(ctx context.Context, cosConfig CosConfig, cosPath string, localPath string) error {
	if cosConfig.Endpoint == "" || cosConfig.SecretID == "" || cosConfig.SecretKey == "" || cosConfig.BucketName == "" {
		return errors.New("cos config is invalid")
	}

	bucketURL := "https://" + cosConfig.BucketName + "." + cosConfig.Endpoint
	u, err := url.Parse(bucketURL)
	if err != nil {
		return err
	}

	client := cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cosConfig.SecretID,
			SecretKey: cosConfig.SecretKey,
		},
	})

	file, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = client.Object.GetToFile(ctx, cosPath, localPath, nil)
	return err
}
