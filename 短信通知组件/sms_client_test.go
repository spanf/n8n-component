package sms

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestCreateSignature(t *testing.T) {
	client := &SMSClient{
		secretId:  "testId",
		secretKey: "testKey",
	}

	params := map[string]string{
		"Action":  "SendSms",
		"Version": "2021-01-11",
		"Region":  "ap-guangzhou",
	}

	signature := client.createSignature(params)

	// 手动计算预期签名
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(k + "=" + params[k])
	}
	signString := buf.String()

	mac := hmac.New(sha256.New, []byte("testKey"))
	mac.Write([]byte(signString))
	expected := hex.EncodeToString(mac.Sum(nil))

	if signature != expected {
		t.Errorf("Expected signature %s, got %s", expected, signature)
	}
}

func TestSendRequest_Success(t *testing.T) {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	client := NewClient("testId", "testKey")
	params := map[string]string{"PhoneNumber": "1234567890"}

	resp, err := client.SendRequest(server.URL, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if string(resp) != "success" {
		t.Errorf("Expected response 'success', got '%s'", string(resp))
	}
}

func TestSendRequest_Non200Status(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient("testId", "testKey")
	params := map[string]string{"PhoneNumber": "1234567890"}

	_, err := client.SendRequest(server.URL, params)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedErr := "unexpected status: 400 Bad Request"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%v'", expectedErr, err)
	}
}

func TestSendRequest_RequestError(t *testing.T) {
	// 创建无效URL触发错误
	client := NewClient("testId", "testKey")
	params := map[string]string{"PhoneNumber": "1234567890"}

	_, err := client.SendRequest("invalid-url", params)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "create request failed") {
		t.Errorf("Expected create request error, got: %v", err)
	}
}

func TestSendRequest_ReadResponseError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// 强制关闭连接导致读取错误
		server := r.Context().Value(http.ServerContextKey).(*http.Server)
		server.Close()
	}))
	defer server.Close()

	client := NewClient("testId", "testKey")
	params := map[string]string{"PhoneNumber": "1234567890"}

	_, err := client.SendRequest(server.URL, params)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "read response failed") {
		t.Errorf("Expected read response error, got: %v", err)
	}
}
