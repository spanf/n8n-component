package wechat_template_message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_GetAccessToken_Success(t *testing.T) {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"access_token": "mock_token",
			"expires_in":   7200,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// 创建客户端并覆盖API地址
	client := &Client{
		config: &Config{
			AppID:     "test_appid",
			AppSecret: "test_secret",
		},
	}
	originalURL := wechatAPIURL
	wechatAPIURL = server.URL + "/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s"
	defer func() { wechatAPIURL = originalURL }()

	token, err := client.getAccessToken()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if token != "mock_token" {
		t.Errorf("Expected token 'mock_token', got '%s'", token)
	}
}

func TestClient_GetAccessToken_WechatError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"errcode": 40013,
			"errmsg":  "invalid appid",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		config: &Config{
			AppID:     "invalid_appid",
			AppSecret: "invalid_secret",
		},
	}
	originalURL := wechatAPIURL
	wechatAPIURL = server.URL + "/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s"
	defer func() { wechatAPIURL = originalURL }()

	_, err := client.getAccessToken()
	if err == nil {
		t.Fatal("Expected error but got none")
	}
	if !strings.Contains(err.Error(), "invalid appid") {
		t.Errorf("Expected 'invalid appid' error, got: %v", err)
	}
}

func TestClient_Send_Success(t *testing.T) {
	// 创建模拟服务器处理两个API请求
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{"access_token": "test_token"}
		json.NewEncoder(w).Encode(response)
	}))
	defer tokenServer.Close()

	sendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求参数
		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		json.Unmarshal(body, &payload)

		if payload["touser"] != "user123" {
			t.Errorf("Expected touser 'user123', got '%v'", payload["touser"])
		}

		response := map[string]interface{}{
			"msgid":   123456,
			"errcode": 0,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer sendServer.Close()

	// 配置客户端
	client := &Client{
		config: &Config{
			AppID:      "app123",
			AppSecret:  "secret123",
			TemplateID: "tpl_123",
		},
	}

	// 覆盖API地址
	originalTokenURL := wechatAPIURL
	originalSendURL := wechatSendURL
	wechatAPIURL = tokenServer.URL + "/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s"
	wechatSendURL = sendServer.URL + "/cgi-bin/message/template/send?access_token=%s"
	defer func() {
		wechatAPIURL = originalTokenURL
		wechatSendURL = originalSendURL
	}()

	// 发送消息
	data := map[string]interface{}{
		"key1": map[string]string{"value": "test_value"},
	}
	msgID, err := client.Send("user123", data, "https://example.com", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if msgID != 123456 {
		t.Errorf("Expected msgID 123456, got %d", msgID)
	}
}

func TestClient_Send_AccessTokenFailure(t *testing.T) {
	// 模拟获取token失败
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"errcode": 40001,
			"errmsg":  "invalid credential",
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
	}))
	defer tokenServer.Close()

	client := &Client{
		config: &Config{
			AppID:     "app123",
			AppSecret: "invalid_secret",
		},
	}

	originalTokenURL := wechatAPIURL
	wechatAPIURL = tokenServer.URL + "/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s"
	defer func() { wechatAPIURL = originalTokenURL }()

	_, err := client.Send("user123", nil, "", nil)
	if err == nil {
		t.Fatal("Expected error but got none")
	}
	if !strings.Contains(err.Error(), "invalid credential") {
		t.Errorf("Expected credential error, got: %v", err)
	}
}

func TestClient_Send_WechatAPIError(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"access_token": "test_token"})
	}))
	defer tokenServer.Close()

	sendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"errcode": 40037,
			"errmsg":  "invalid template_id",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
	}))
	defer sendServer.Close()

	client := &Client{
		config: &Config{
			AppID:      "app123",
			AppSecret:  "secret123",
			TemplateID: "invalid_template",
		},
	}

	originalTokenURL := wechatAPIURL
	originalSendURL := wechatSendURL
	wechatAPIURL = tokenServer.URL + "/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s"
	wechatSendURL = sendServer.URL + "/cgi-bin/message/template/send?access_token=%s"
	defer func() {
		wechatAPIURL = originalTokenURL
		wechatSendURL = originalSendURL
	}()

	_, err := client.Send("user123", nil, "", nil)
	if err == nil {
		t.Fatal("Expected error but got none")
	}
	if !strings.Contains(err.Error(), "invalid template_id") {
		t.Errorf("Expected template ID error, got: %v", err)
	}
}
