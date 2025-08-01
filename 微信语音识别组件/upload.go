package wechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func UploadVoiceFile(filePath string, accessToken string, format string) (string, error) {
	if valid, err := validateFile(filePath); !valid || err != nil {
		return "", fmt.Errorf("文件验证失败: %v", err)
	}

	req, err := buildUploadRequest(filePath, accessToken, format)
	if err != nil {
		return "", fmt.Errorf("构建请求失败: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求发送失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("响应读取失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("微信服务器错误: 状态码%d, 响应:%s", resp.StatusCode, body)
	}

	var result struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
		MediaID string `json:"media_id"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("响应解析失败: %v", err)
	}

	if result.Errcode != 0 {
		return "", fmt.Errorf("微信接口错误[%d]: %s", result.Errcode, result.Errmsg)
	}

	return result.MediaID, nil
}

func validateFile(filePath string) (bool, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Errorf("文件不存在")
		}
		return false, err
	}

	if info.IsDir() {
		return false, fmt.Errorf("路径指向目录而非文件")
	}

	if info.Size() == 0 {
		return false, fmt.Errorf("文件内容为空")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return false, fmt.Errorf("文件无法打开: %v", err)
	}
	defer file.Close()

	return true, nil
}

func buildUploadRequest(filePath string, accessToken string, format string) (*http.Request, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("media", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}

	if _, err = io.Copy(part, file); err != nil {
		return nil, err
	}
	writer.Close()

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/media/upload?access_token=%s&type=%s", accessToken, format)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}
