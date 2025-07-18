package wechatpay_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"path/to/your/package/wechatpay"
)

type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestCloseOrder_Success(t *testing.T) {
	mockClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			resp := httptest.NewRecorder()
			resp.WriteHeader(http.StatusNoContent)
			return resp.Result(), nil
		},
	}

	origClient := wechatpay.HTTPClient
	wechatpay.HTTPClient = mockClient
	defer func() { wechatpay.HTTPClient = origClient }()

	cred := &wechatpay.Credential{ApiKey: "test-key"}
	err := wechatpay.CloseOrder(context.Background(), "mch123", "order456", cred)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCloseOrder_HTTPError(t *testing.T) {
	mockClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("connection failed")
		},
	}

	origClient := wechatpay.HTTPClient
	wechatpay.HTTPClient = mockClient
	defer func() { wechatpay.HTTPClient = origClient }()

	cred := &wechatpay.Credential{ApiKey: "test-key"}
	err := wechatpay.CloseOrder(context.Background(), "mch123", "order456", cred)
	if err == nil || err.Error() != "connection failed" {
		t.Errorf("Expected connection error, got %v", err)
	}
}

func TestCloseOrder_Non204Response(t *testing.T) {
	mockClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			resp := httptest.NewRecorder()
			resp.WriteHeader(http.StatusBadRequest)
			resp.WriteString(`{"code":"INVALID_REQUEST"}`)
			return resp.Result(), nil
		},
	}

	origClient := wechatpay.HTTPClient
	wechatpay.HTTPClient = mockClient
	defer func() { wechatpay.HTTPClient = origClient }()

	cred := &wechatpay.Credential{ApiKey: "test-key"}
	err := wechatpay.CloseOrder(context.Background(), "mch123", "order456", cred)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expected := "unexpected status code: 400, body: {\"code\":\"INVALID_REQUEST\"}"
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestCloseOrder_ContextCancel(t *testing.T) {
	mockClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			time.Sleep(100 * time.Millisecond)
			return &http.Response{StatusCode: http.StatusNoContent}, nil
		},
	}

	origClient := wechatpay.HTTPClient
	wechatpay.HTTPClient = mockClient
	defer func() { wechatpay.HTTPClient = origClient }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cred := &wechatpay.Credential{ApiKey: "test-key"}
	err := wechatpay.CloseOrder(ctx, "mch123", "order456", cred)
	if err == nil || err != context.Canceled {
		t.Errorf("Expected context canceled error, got %v", err)
	}
}

func TestCloseOrder_InvalidPayload(t *testing.T) {
	// 测试无效的商户ID导致JSON序列化失败
	// 注意：当前实现中商户ID是字符串类型，不会序列化失败
	// 添加此测试作为占位符，实际应用中可能需要其他测试
	t.Skip("No invalid payload scenario in current implementation")
}

func TestBuildCloseOrderRequest_InvalidURL(t *testing.T) {
	// 保存原始函数以便恢复
	origBuild := wechatpay.BuildCloseOrderRequest
	defer func() { wechatpay.BuildCloseOrderRequest = origBuild }()

	// 注入会返回错误的构建函数
	wechatpay.BuildCloseOrderRequest = func(mchid, outTradeNo string) (*http.Request, error) {
		return nil, errors.New("invalid URL format")
	}

	cred := &wechatpay.Credential{ApiKey: "test-key"}
	err := wechatpay.CloseOrder(context.Background(), "mch123", "order456", cred)
	if err == nil || err.Error() != "invalid URL format" {
		t.Errorf("Expected URL error, got %v", err)
	}
}
