package cosuploader

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/tencentyun/cos-go-sdk-v5"
)

func UploadFile(ctx context.Context, cosConfig struct {
	Endpoint   string
	SecretID   string
	SecretKey  string
	BucketName string
}, filePath string, cosPath string) (string, error) {
	if cosConfig.Endpoint == "" || cosConfig.SecretID == "" || cosConfig.SecretKey == "" || cosConfig.BucketName == "" {
		return "", errors.New("invalid cos configuration")
	}

	u, err := url.Parse("https://" + cosConfig.BucketName + "." + cosConfig.Endpoint)
	if err != nil {
		return "", err
	}

	client := cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cosConfig.SecretID,
			SecretKey: cosConfig.SecretKey,
		},
	})

	_, err = client.Object.PutFromFile(ctx, cosPath, filePath, nil)
	if err != nil {
		return "", err
	}

	return "https://" + cosConfig.BucketName + "." + cosConfig.Endpoint + "/" + cosPath, nil
}
