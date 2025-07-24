package wechat_template_message

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

func GetAccessToken(appID, appSecret string) (string, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appID, appSecret)
	
	body, err := sendTokenRequest(url)
	if err != nil {
		return "", fmt.Errorf("请求失败: %v", err)
	}
	
	var resp tokenResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("JSON解析失败: %v", err)
	}
	
	if resp.ErrCode != 0 {
		return "", fmt.Errorf("微信接口错误[%d]: %s", resp.ErrCode, resp.ErrMsg)
	}
	
	if resp.AccessToken == "" {
		return "", fmt.Errorf("获取的AccessToken为空")
	}
	
	return resp.AccessToken, nil
}

func sendTokenRequest(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP状态码异常: %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}
	
	return body, nil
}
