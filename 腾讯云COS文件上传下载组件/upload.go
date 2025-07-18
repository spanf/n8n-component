package cosutil

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/tencentyun/cos-go-sdk-v5"
)

func UploadFile(client *cos.Client, localPath, cosPath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentLength: stat.Size(),
		},
	}

	_, err = client.Object.Put(context.Background(), cosPath, file, opt)
	return err
}

func UploadBytes(client *cos.Client, data []byte, cosPath string) error {
	reader := bytes.NewReader(data)
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentLength: int64(len(data)),
		},
	}

	_, err := client.Object.Put(context.Background(), cosPath, reader, opt)
	return err
}
