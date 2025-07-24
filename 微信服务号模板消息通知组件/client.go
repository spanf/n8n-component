package wechat_template_message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Config 组件配置结构
type Config struct {
	AppID      string
	AppSecret  string
	TemplateID string
}

// Client 客户端结构体
type Client struct {
	config *Config
}

// NewClient 创建新客户端
func NewClient(config *Config) *Client {
	return &Client{config: config}
}

// getAccessToken 获取微信access_token
func (c *Client) getAccessToken() (string, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		c.config.AppID, c.config.AppSecret)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response failed: %w", err)
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("json unmarshal failed: %w", err)
	}

	if result.ErrCode != 0 {
		return "", fmt.Errorf("wechat api error[%d]: %s", result.ErrCode, result.ErrMsg)
	}

	return result.AccessToken, nil
}

// Send 发送模板消息
func (c *Client) Send(openID string, data map[string]interface{}, url string, miniprogram map[string]string) (int64, error) {
	accessToken, err := c.getAccessToken()
	if err != nil {
		return 0, fmt.Errorf("get access token failed: %w", err)
	}

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=%s", accessToken)

	payload := map[string]interface{}{
		"touser":      openID,
		"template_id": c.config.TemplateID,
		"data":        data,
	}

	if url != "" {
		payload["url"] = url
	}
	if miniprogram != nil {
		payload["miniprogram"] = miniprogram
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("json marshal failed: %w", err)
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("post request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read response failed: %w", err)
	}

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		MsgID   int64  `json:"msgid"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("json unmarshal failed: %w", err)
	}

	if result.ErrCode != 0 {
		return 0, fmt.Errorf("wechat api error[%d]: %s", result.ErrCode, result.ErrMsg)
	}

	return result.MsgID, nil
}
