package wechatpay

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestQueryRefund_Success(t *testing.T) {
	// 创建模拟服务器
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"refund_id": "REF123456789",
			"out_refund_no": "ORDER_123",
			"status": "success",
			"amount": 1000,
			"success_time": "2023-04-01T12:34:56+08:00",
			"user_received_account": "招商银行信用卡0403"
		}`))
	}))
	defer ts.Close()

	// 临时替换全局URL
	originalURL := baseURL
	baseURL = ts.URL + "/v3/refund/domestic/refunds"
	defer func() { baseURL = originalURL }()

	// 生成测试私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成私钥失败: %v", err)
	}

	// 构建请求
	req := QueryRequest{OutRefundNo: "ORDER_123"}

	// 调用测试函数
	resp, err := QueryRefund("MCH123", "SERIAL001", privateKey, req)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}

	// 验证结果
	expectedTime := "2023-04-01T12:34:56+08:00"
	if resp.RefundID != "REF123456789" || resp.OutRefundNo != "ORDER_123" ||
		resp.Status != "SUCCESS" || resp.Amount != 1000 || resp.SuccessTime != expectedTime ||
		resp.UserReceived != "招商银行信用卡0403" {
		t.Errorf("响应不匹配\n期望: REF123456789, ORDER_123, SUCCESS, 1000, %s, 招商银行信用卡0403\n实际: %s, %s, %s, %d, %s, %s",
			expectedTime, resp.RefundID, resp.OutRefundNo, resp.Status, resp.Amount, resp.SuccessTime, resp.UserReceived)
	}
}

func TestQueryRefund_HTTPError(t *testing.T) {
	// 创建返回500错误的模拟服务器
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	// 临时替换全局URL
	originalURL := baseURL
	baseURL = ts.URL + "/v3/refund/domestic/refunds"
	defer func() { baseURL = originalURL }()

	// 生成测试私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成私钥失败: %v", err)
	}

	// 调用测试函数
	_, err = QueryRefund("MCH123", "SERIAL001", privateKey, QueryRequest{OutRefundNo: "ORDER_123"})
	if err == nil || err.Error() != "HTTP error: 500 Internal Server Error" {
		t.Errorf("期望HTTP错误，实际错误: %v", err)
	}
}

func TestQueryRefund_InvalidJSON(t *testing.T) {
	// 创建返回无效JSON的模拟服务器
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer ts.Close()

	// 临时替换全局URL
	originalURL := baseURL
	baseURL = ts.URL + "/v3/refund/domestic/refunds"
	defer func() { baseURL = originalURL }()

	// 生成测试私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成私钥失败: %v", err)
	}

	// 调用测试函数
	_, err = QueryRefund("MCH123", "SERIAL001", privateKey, QueryRequest{OutRefundNo: "ORDER_123"})
	if err == nil || !strings.Contains(err.Error(), "invalid character") {
		t.Errorf("期望JSON解析错误，实际错误: %v", err)
	}
}

func TestBuildQueryURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"REF123", baseURL + "/REF123"},
		{"ref/und", baseURL + "/ref%2Fund"},
		{"订单@123", baseURL + "/%E8%AE%A2%E5%8D%95%40123"},
	}

	for _, test := range tests {
		result := buildQueryURL(test.input)
		if result != test.expected {
			t.Errorf("输入: %s\n期望: %s\n实际: %s", test.input, test.expected, result)
		}
	}
}

func TestParseQueryResponse(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected *QueryResponse
		hasError bool
	}{
		{
			name:  "ValidResponse",
			input: `{"refund_id":"R123","out_refund_no":"O123","status":"success","amount":100,"success_time":"2023-01-01T00:00:00Z","user_received_account":"account123"}`,
			expected: &QueryResponse{
				RefundID:     "R123",
				OutRefundNo:  "O123",
				Status:       "SUCCESS",
				Amount:       100,
				SuccessTime:  "2023-01-01T00:00:00Z",
				UserReceived: "account123",
			},
			hasError: false,
		},
		{
			name:     "InvalidJSON",
			input:    `invalid json`,
			expected: nil,
			hasError: true,
		},
		{
			name:     "EmptyResponse",
			input:    `{}`,
			expected: &QueryResponse{Status: ""},
			hasError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := parseQueryResponse([]byte(tc.input))
			
			if tc.hasError {
				if err == nil {
					t.Error("期望错误，但未返回错误")
				}
				return
			}
			
			if err != nil {
				t.Fatalf("意外错误: %v", err)
			}
			
			if resp.RefundID != tc.expected.RefundID ||
				resp.OutRefundNo != tc.expected.OutRefundNo ||
				resp.Status != tc.expected.Status ||
				resp.Amount != tc.expected.Amount ||
				resp.SuccessTime != tc.expected.SuccessTime ||
				resp.UserReceived != tc.expected.UserReceived {
				t.Errorf("结果不匹配\n期望: %+v\n实际: %+v", tc.expected, resp)
			}
		})
	}
}
