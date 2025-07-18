package wechatpayrefund

import (
	"testing"
)

func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		actual   string
		expected string
	}{
		{"API Host", apiHost, "https://api.mch.weixin.qq.com"},
		{"Refund Path", refundPath, "/v3/refund/domestic/refunds"},
		{"Query Path", queryPath, "/v3/refund/domestic/refunds/"},
		{"Auth Type", authType, "WECHATPAY2-SHA256-RSA2048"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.actual != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.actual)
			}
		})
	}
}

func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"ErrInvalidRequest", ErrInvalidRequest, "invalid request"},
		{"ErrRequestFailed", ErrRequestFailed, "request failed"},
		{"ErrInvalidResponse", ErrInvalidResponse, "invalid response"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.err.Error())
			}
		})
	}
}

func TestErrorTypes(t *testing.T) {
	if ErrInvalidRequest == nil || ErrRequestFailed == nil || ErrInvalidResponse == nil {
		t.Error("error variables should not be nil")
	}

	if ErrInvalidRequest == ErrRequestFailed || 
	   ErrInvalidRequest == ErrInvalidResponse || 
	   ErrRequestFailed == ErrInvalidResponse {
		t.Error("error variables should be distinct")
	}
}
