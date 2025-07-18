package cosutil

import (
	"context"
	"io"
	"os"

	"github.com/tencentyun/cos-go-sdk-v5"
)

func DownloadFile(client *cos.Client, cosPath, localPath string) error {
	// 创建本地文件
	file, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 从COS下载文件
	resp, err := client.Object.Get(context.Background(), cosPath, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 将内容写入本地文件
	_, err = io.Copy(file, resp.Body)
	return err
}

func DownloadBytes(client *cos.Client, cosPath string) ([]byte, error) {
	// 从COS下载文件
	resp, err := client.Object.Get(context.Background(), cosPath, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取字节数据
	return io.ReadAll(resp.Body)
}
