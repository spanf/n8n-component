package wechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// RecognizeVoice 识别语音内容
func RecognizeVoice(mediaID, accessToken, language string) (string, error) {
	req, err := buildRecognitionRequest(mediaID, accessToken, language)
	if err != nil {
		return "", fmt.Errorf("构建请求失败: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("微信API返回错误状态码: %d", resp.StatusCode)
	}

	return parseRecognitionResponse(resp)
}

// buildRecognitionRequest 构建识别请求
func buildRecognitionRequest(mediaID, accessToken, language string) (*http.Request, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/media/voice/recognize?access_token=%s", accessToken)
	
	payload := map[string]string{
		"media_id": mediaID,
		"lang":     language,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	
	return req, nil
}

// parseRecognitionResponse 解析识别响应
func parseRecognitionResponse(resp *http.Response) (string, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var result struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
		Text    string `json:"text"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析JSON失败: %w", err)
	}

	if result.Errcode != 0 {
		return "", fmt.Errorf("微信API错误: %s (代码%d)", result.Errmsg, result.Errcode)
	}

	return result.Text, nil
}
