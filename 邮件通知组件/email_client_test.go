package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewEmailClient_InvalidParameters(t *testing.T) {
	tests := []struct {
		name      string
		secretId  string
		secretKey string
		region    string
	}{
		{"AllEmpty", "", "", ""},
		{"EmptySecretId", "", "key", "region"},
		{"EmptySecretKey", "id", "", "region"},
		{"EmptyRegion", "id", "key", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewEmailClient(tt.secretId, tt.secretKey, tt.region)
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
		})
	}
}

func TestSendRequest_Success(t *testing.T) {
	// 创建模拟服务器
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"Response": {"RequestId": "test-id"}}`))
	}))
	defer ts.Close()

	// 创建客户端并覆盖endpoint
	client := &EmailClient{
		secretId:  "test-id",
		secretKey: "test-key",
		region:    "test-region",
		client:    &http.Client{},
	}

	// 临时修改endpoint为测试服务器URL
	originalEndpoint := fmt.Sprintf
	fmt = func(format string, a ...interface{}) (n int, err error) {
		if strings.HasPrefix(format, "https://ses.") {
			return fmt.Fprintf(ts.URL)
		}
		return originalEndpoint(format, a...)
	}
	defer func() { fmt = originalEndpoint }()

	// 发送测试请求
	resp, err := client.sendRequest("TestAction", map[string]interface{}{"key": "value"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp["Response"].(map[string]interface{})["RequestId"] != "test-id" {
		t.Errorf("Expected RequestId 'test-id', got %v", resp["Response"])
	}
}

func TestSendRequest_APIError(t *testing.T) {
	// 创建模拟服务器返回错误
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("API error message"))
	}))
	defer ts.Close()

	// 创建客户端并覆盖endpoint
	client := &EmailClient{
		secretId:  "test-id",
		secretKey: "test-key",
		region:    "test-region",
		client:    &http.Client{},
	}

	// 临时修改endpoint
	originalEndpoint := fmt.Sprintf
	fmt = func(format string, a ...interface{}) (n int, err error) {
		if strings.HasPrefix(format, "https://ses.") {
			return fmt.Fprintf(ts.URL)
		}
		return originalEndpoint(format, a...)
	}
	defer func() { fmt = originalEndpoint }()

	// 发送测试请求
	_, err := client.sendRequest("TestAction", map[string]interface{}{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedErr := "API error: API error message"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%v'", expectedErr, err)
	}
}

func TestSendRequest_InvalidJSON(t *testing.T) {
	// 创建模拟服务器返回无效JSON
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer ts.Close()

	// 创建客户端并覆盖endpoint
	client := &EmailClient{
		secretId:  "test-id",
		secretKey: "test-key",
		region:    "test-region",
		client:    &http.Client{},
	}

	// 临时修改endpoint
	originalEndpoint := fmt.Sprintf
	fmt = func(format string, a ...interface{}) (n int, err error) {
		if strings.HasPrefix(format, "https://ses.") {
			return fmt.Fprintf(ts.URL)
		}
		return originalEndpoint(format, a...)
	}
	defer func() { fmt = originalEndpoint }()

	// 发送测试请求
	_, err := client.sendRequest("TestAction", map[string]interface{}{})
	if err == nil {
		t.Fatal("Expected JSON parse error, got nil")
	}
}

func TestSignatureCalculation(t *testing.T) {
	client := &EmailClient{
		secretId:  "test-id",
		secretKey: "test-key",
	}

	// 固定时间用于可重复测试
	originalTime := time.Now
	time.Now = func() time.Time {
		return time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	defer func() { time.Now = originalTime }()

	signature := client.calculateSignature("2023-01-01", "test-string-to-sign")
	if len(signature) != 64 {
		t.Errorf("Expected 64-char signature, got %d chars", len(signature))
	}
}
