package main

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/tencentyun/cos-go-sdk-v5"
)

func UploadFile(client *cos.Client, bucketName string, cosPath string, localPath string) error {
	opt := initUploadRequest(client, bucketName, cosPath, localPath)
	if opt == nil {
		return os.ErrInvalid
	}
	return doUpload(client, cosPath, opt)
}

func initUploadRequest(client *cos.Client, bucketName string, cosPath string, localPath string) *cos.ObjectPutOptions {
	file, err := os.Open(localPath)
	if err != nil {
		return nil
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil
	}

	return &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentLength: stat.Size(),
		},
		Body: file,
	}
}

func doUpload(client *cos.Client, cosPath string, opt *cos.ObjectPutOptions) error {
	if closer, ok := opt.Body.(io.Closer); ok {
		defer closer.Close()
	}

	_, err := client.Object.Put(
		context.Background(),
		cosPath,
		opt.Body,
		opt,
	)
	return err
}
