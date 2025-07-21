package tencent_vector_db_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	. "tencent_vector_db" // 替换为实际包名
)

func TestExtractRegion(t *testing.T) {
	tests := []struct {
		endpoint string
		expected string
	}{
		{"https://vector.ap-beijing.tencentcloudapi.com", "ap-beijing"},
		{"http://vector.na-siliconvalley.tencentcloudapi.com", "na-siliconvalley"},
		{"https://invalid.url.example.com", ""},
		{"http://[::1]:8080", ""},
	}

	for _, tt := range tests {
		client := NewVectorClient("id", "key", tt.endpoint)
		result := client.extractRegion()
		if result != tt.expected {
			t.Errorf("For endpoint %s, expected %s, got %s", tt.endpoint, tt.expected, result)
		}
	}
}

func TestExtractHost(t *testing.T) {
	tests := []struct {
		endpoint string
		expected string
	}{
		{"https://vector.ap-shanghai.tencentcloudapi.com", "vector.ap-shanghai.tencentcloudapi.com"},
		{"http://localhost:8080", "localhost"},
		{"invalid-url", ""},
	}

	for _, tt := range tests {
		client := NewVectorClient("id", "key", tt.endpoint)
		result := client.extractHost()
		if result != tt.expected {
			t.Errorf("For endpoint %s, expected %s, got %s", tt.endpoint, tt.expected, result)
		}
	}
}

func TestInternalSendRequest_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/test" {
			t.Errorf("Expected path /test, got %s", r.URL.Path)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", ct)
		}

		authHeader := r.Header.Get("Authorization")
		if !strings.Contains(authHeader, "TC3-HMAC-SHA256 Credential=testId") {
			t.Errorf("Invalid Authorization header: %s", authHeader)
		}

		// 验证请求体
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal("Failed to decode request body")
		}
		if body["action"] != "query" {
			t.Errorf("Unexpected body content: %v", body)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"success": true}`)
	}))
	defer ts.Close()

	client := NewVectorClient("testId", "testKey", ts.URL)
	requestBody := map[string]interface{}{"action": "query"}

	resp, err := client.internalSendRequest("POST", "/test", requestBody)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := `{"success": true}` + "\n"
	if string(resp) != expected {
		t.Errorf("Expected response %q, got %q", expected, string(resp))
	}
}

func TestInternalSendRequest_ErrorStatusCode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, `{"error": "invalid request"}`)
	}))
	defer ts.Close()

	client := NewVectorClient("testId", "testKey", ts.URL)
	_, err := client.internalSendRequest("GET", "/error", nil)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedErr := "API error: 400 Bad Request"
	if err.Error() != expectedErr {
		t.Errorf("Expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestInternalSendRequest_InvalidEndpoint(t *testing.T) {
	client := NewVectorClient("id", "key", "invalid-url://test")
	_, err := client.internalSendRequest("POST", "/path", nil)
	if err == nil || !strings.Contains(err.Error(), "invalid endpoint") {
		t.Errorf("Expected invalid endpoint error, got: %v", err)
	}
}

func TestSignatureGeneration(t *testing.T) {
	client := NewVectorClient("id", "key", "https://vector.ap-guangzhou.tencentcloudapi.com")
	signature := client.internalGetSignature("POST", "/test", nil)

	if len(signature) != 64 {
		t.Errorf("Expected 64-character signature, got %d: %s", len(signature), signature)
	}
}

func TestRequestWithParams(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "param1=value1&param2=value2" {
			t.Errorf("Unexpected query string: %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewVectorClient("id", "key", ts.URL)
	// 注意：原代码中internalSendRequest未使用params参数，这里仅展示测试结构
	_, err := client.internalSendRequest("GET", "/path?param1=value1&param2=value2", nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestTimeSensitivity(t *testing.T) {
	start := time.Now()
	client := NewVectorClient("id", "key", "https://vector.ap-hongkong.tencentcloudapi.com")
	sig1 := client.internalGetSignature("GET", "/path", nil)

	time.Sleep(2 * time.Second)
	sig2 := client.internalGetSignature("GET", "/path", nil)

	if sig1 == sig2 {
		t.Error("Signature should change over time")
	}
}
