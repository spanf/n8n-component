package wxpay_test

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"testing"
	"time"
	"wxpay"

	"github.com/jarcoal/httpmock"
)

func TestRefund_Success(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// 模拟私钥
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	
	// 模拟微信支付API响应
	mockResponse := `{
		"refund_id": "REF123456789",
		"out_refund_no": "REFUND_2023",
		"status": "PROCESSING",
		"create_time": "2023-01-01T10:00:00Z"
	}`
	httpmock.RegisterResponder("POST", "https://api.mch.weixin.qq.com/v3/refund/domestic/refunds",
		httpmock.NewStringResponder(200, mockResponse))

	req := wxpay.RefundRequest{
		OutTradeNo:  "ORDER_123",
		OutRefundNo: "REFUND_2023",
		Amount:      1000,
		TotalAmount: 2000,
		Reason:      "Test refund",
	}

	resp, err := wxpay.Refund("MCH123", "SERIAL001", privateKey, req)
	if err != nil {
		t.Fatalf("Refund failed: %v", err)
	}

	if resp.RefundID != "REF123456789" {
		t.Errorf("Expected RefundID 'REF123456789', got '%s'", resp.RefundID)
	}
	if resp.Status != "PROCESSING" {
		t.Errorf("Expected status 'PROCESSING', got '%s'", resp.Status)
	}
}

func TestRefund_HTTPError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	
	// 模拟服务器错误
	httpmock.RegisterResponder("POST", "https://api.mch.weixin.qq.com/v3/refund/domestic/refunds",
		httpmock.NewStringResponder(500, "Internal Server Error"))

	req := wxpay.RefundRequest{
		OutTradeNo:  "ORDER_123",
		OutRefundNo: "REFUND_2023",
		Amount:      1000,
		TotalAmount: 2000,
	}

	_, err := wxpay.Refund("MCH123", "SERIAL001", privateKey, req)
	if err == nil {
		t.Fatal("Expected HTTP error, got nil")
	}
	
	expectedErr := "HTTP error: 500 Internal Server Error"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestRefund_InvalidResponse(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	
	// 模拟无效JSON响应
	httpmock.RegisterResponder("POST", "https://api.mch.weixin.qq.com/v3/refund/domestic/refunds",
		httpmock.NewStringResponder(200, "{invalid json}"))

	req := wxpay.RefundRequest{
		OutTradeNo:  "ORDER_123",
		OutRefundNo: "REFUND_2023",
		Amount:      1000,
		TotalAmount: 2000,
	}

	_, err := wxpay.Refund("MCH123", "SERIAL001", privateKey, req)
	if err == nil {
		t.Fatal("Expected JSON parse error, got nil")
	}
}

func TestRefund_RequestBuild(t *testing.T) {
	// 测试请求体构建逻辑
	req := wxpay.RefundRequest{
		OutTradeNo:  "ORDER_1001",
		OutRefundNo: "REF_1001",
		Amount:      500,
		TotalAmount: 1500,
		Reason:      "Customer request",
	}

	body, err := wxpay.BuildRefundBody(req)
	if err != nil {
		t.Fatalf("BuildRefundBody failed: %v", err)
	}

	expected := `{"amount":{"currency":"CNY","refund":500,"total":1500},` +
		`"out_refund_no":"REF_1001","out_trade_no":"ORDER_1001","reason":"Customer request"}`
	if string(body) != expected {
		t.Errorf("Expected body:\n%s\nGot:\n%s", expected, string(body))
	}
}

func TestRefund_Signature(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	
	// 捕获请求头进行验证
	var authHeader string
	httpmock.RegisterResponder("POST", "https://api.mch.weixin.qq.com/v3/refund/domestic/refunds",
		func(req *http.Request) (*http.Response, error) {
			authHeader = req.Header.Get("Authorization")
			return httpmock.NewStringResponse(200, `{"refund_id":"TEST123"}`), nil
		})

	req := wxpay.RefundRequest{
		OutTradeNo:  "SIGN_TEST",
		OutRefundNo: "REF_SIGN",
		Amount:      100,
		TotalAmount: 100,
	}

	_, err := wxpay.Refund("MCH_SIGN", "SERIAL_ABC", privateKey, req)
	if err != nil {
		t.Fatalf("Refund failed: %v", err)
	}

	if authHeader == "" {
		t.Fatal("Authorization header missing")
	}
	if len(authHeader) < 100 || authHeader[:12] != "WECHATPAY2-S" {
		t.Errorf("Invalid Authorization header format: %s", authHeader)
	}
}

func TestRefund_Timeout(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	
	// 模拟超时响应
	httpmock.RegisterResponder("POST", "https://api.mch.weixin.qq.com/v3/refund/domestic/refunds",
		func(req *http.Request) (*http.Response, error) {
			time.Sleep(35 * time.Second) // 超过30秒超时设置
			return httpmock.NewStringResponse(200, ""), nil
		})

	req := wxpay.RefundRequest{
		OutTradeNo:  "ORDER_123",
		OutRefundNo: "REFUND_2023",
		Amount:      1000,
		TotalAmount: 2000,
	}

	_, err := wxpay.Refund("MCH123", "SERIAL001", privateKey, req)
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
}
