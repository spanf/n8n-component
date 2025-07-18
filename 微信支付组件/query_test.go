package wechatpay_test

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"your_module_path/wechatpay"
)

func TestQueryOrder_Success(t *testing.T) {
	// 创建模拟服务器
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := fmt.Sprintf("/v3/pay/transactions/out-trade-no/testOrder?mchid=testMchID")
		if r.URL.Path+r.URL.RawQuery != expectedPath[1:] {
			t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path+"?"+r.URL.RawQuery)
		}

		resp := wechatpay.QueryOrderResponse{
			Appid:          "app123",
			Mchid:          "testMchID",
			OutTradeNo:     "testOrder",
			TransactionID:  "trans123",
			TradeType:      "JSAPI",
			TradeState:     "SUCCESS",
			TradeStateDesc: "支付成功",
			Amount: struct {
				Total         int    `json:"total"`
				Currency      string `json:"currency"`
				PayerTotal    int    `json:"payer_total"`
				PayerCurrency string `json:"payer_currency"`
			}{
				Total:    100,
				Currency: "CNY",
			},
			Payer: struct {
				Openid string `json:"openid"`
			}{
				Openid: "user123",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	// 重写URL构建函数
	originalBuild := wechatpay.BuildQueryOrderRequest
	wechatpay.BuildQueryOrderRequest = func(mchid, outTradeNo string) (*http.Request, error) {
		url := fmt.Sprintf("%s/v3/pay/transactions/out-trade-no/%s?mchid=%s", ts.URL, outTradeNo, mchid)
		return http.NewRequest("GET", url, nil)
	}
	defer func() { wechatpay.BuildQueryOrderRequest = originalBuild }()

	// 准备测试数据
	cred := &wechatpay.Credential{
		MchID:      "testMchID",
		SerialNo:   "serial123",
		PrivateKey: &rsa.PrivateKey{}, // 实际测试中应使用有效私钥
	}

	// 调用被测函数
	resp, err := wechatpay.QueryOrder(context.Background(), "testMchID", "testOrder", cred)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证结果
	if resp.Appid != "app123" || resp.OutTradeNo != "testOrder" {
		t.Errorf("unexpected response data: %+v", resp)
	}
	if resp.Amount.Total != 100 {
		t.Errorf("expected amount 100, got %d", resp.Amount.Total)
	}
}

func TestQueryOrder_HTTPError(t *testing.T) {
	// 创建模拟服务器返回错误
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer ts.Close()

	// 重写URL构建函数
	originalBuild := wechatpay.BuildQueryOrderRequest
	wechatpay.BuildQueryOrderRequest = func(mchid, outTradeNo string) (*http.Request, error) {
		url := fmt.Sprintf("%s/v3/pay/transactions/out-trade-no/%s?mchid=%s", ts.URL, outTradeNo, mchid)
		return http.NewRequest("GET", url, nil)
	}
	defer func() { wechatpay.BuildQueryOrderRequest = originalBuild }()

	cred := &wechatpay.Credential{
		MchID:      "testMchID",
		SerialNo:   "serial123",
		PrivateKey: &rsa.PrivateKey{},
	}

	_, err := wechatpay.QueryOrder(context.Background(), "testMchID", "testOrder", cred)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expectedErr := "HTTP error: 500 Internal Server Error, body: server error"
	if err.Error() != expectedErr {
		t.Errorf("expected error '%s', got '%v'", expectedErr, err)
	}
}

func TestQueryOrder_InvalidJSON(t *testing.T) {
	// 创建返回无效JSON的模拟服务器
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{invalid json}"))
	}))
	defer ts.Close()

	// 重写URL构建函数
	originalBuild := wechatpay.BuildQueryOrderRequest
	wechatpay.BuildQueryOrderRequest = func(mchid, outTradeNo string) (*http.Request, error) {
		url := fmt.Sprintf("%s/v3/pay/transactions/out-trade-no/%s?mchid=%s", ts.URL, outTradeNo, mchid)
		return http.NewRequest("GET", url, nil)
	}
	defer func() { wechatpay.BuildQueryOrderRequest = originalBuild }()

	cred := &wechatpay.Credential{
		MchID:      "testMchID",
		SerialNo:   "serial123",
		PrivateKey: &rsa.PrivateKey{},
	}

	_, err := wechatpay.QueryOrder(context.Background(), "testMchID", "testOrder", cred)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBuildQueryOrderRequest(t *testing.T) {
	req, err := wechatpay.BuildQueryOrderRequest("mch123", "order456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedURL := "https://api.mch.weixin.qq.com/v3/pay/transactions/out-trade-no/order456?mchid=mch123"
	if req.URL.String() != expectedURL {
		t.Errorf("expected URL %s, got %s", expectedURL, req.URL.String())
	}
	if req.Method != "GET" {
		t.Errorf("expected GET method, got %s", req.Method)
	}
}

func TestParseQueryOrderResponse_Success(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       mockBody(`{"appid":"app789","out_trade_no":"order123","amount":{"total":200}}`),
	}

	result, err := wechatpay.ParseQueryOrderResponse(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Appid != "app789" || result.OutTradeNo != "order123" {
		t.Errorf("unexpected response data: %+v", result)
	}
	if result.Amount.Total != 200 {
		t.Errorf("expected amount 200, got %d", result.Amount.Total)
	}
}

func TestParseQueryOrderResponse_ErrorStatus(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       mockBody(`{"code":"INVALID_REQUEST"}`),
	}

	_, err := wechatpay.ParseQueryOrderResponse(resp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expectedErr := "HTTP error: 400 Bad Request, body: {\"code\":\"INVALID_REQUEST\"}"
	if err.Error() != expectedErr {
		t.Errorf("expected error '%s', got '%v'", expectedErr, err)
	}
}

func mockBody(content string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(content))
}

func TestGenerateNonce(t *testing.T) {
	nonce := wechatpay.GenerateNonce(16)
	if len(nonce) != 16 {
		t.Errorf("expected length 16, got %d", len(nonce))
	}

	// 验证字符集
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for _, c := range nonce {
		if !strings.ContainsRune(charset, c) {
			t.Errorf("invalid character in nonce: %c", c)
		}
	}
}
