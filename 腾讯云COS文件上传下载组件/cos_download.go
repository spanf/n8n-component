package cosop

import (
	"context"
	"io"
	"net/http"
	"os"

	cos "github.com/tencentyun/cos-go-sdk-v5"
)

func DownloadFile(client *cos.Client, bucketName string, cosPath string, localPath string) error {
	opt := initDownloadRequest(client, bucketName, cosPath)
	return doDownload(client, opt, cosPath, localPath)
}

func initDownloadRequest(client *cos.Client, bucketName string, cosPath string) *cos.ObjectGetOptions {
	return &cos.ObjectGetOptions{}
}

func doDownload(client *cos.Client, opt *cos.ObjectGetOptions, cosPath, localPath string) error {
	resp, err := client.Object.Get(context.Background(), cosPath, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &cos.ErrorResponse{
			Response: resp,
			Message:  "download failed with non-200 status",
		}
	}

	outFile, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	return err
}
