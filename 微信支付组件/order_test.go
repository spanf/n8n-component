package payment_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"your_module_path/payment"
)

func TestValidateCreateOrderParams(t *testing.T) {
	tests := []struct {
		name    string
		params  payment.CreateOrderParams
		wantErr error
	}{
		{
			name:    "missing appid",
			params:  payment.CreateOrderParams{Mchid: "mch", Description: "desc", OutTradeNo: "no", NotifyURL: "url", Amount: payment.Amount{Total: 100, Currency: "CNY"}},
			wantErr: errors.New("appid is required"),
		},
		{
			name:    "invalid amount",
			params:  payment.CreateOrderParams{Appid: "app", Mchid: "mch", Description: "desc", OutTradeNo: "no", NotifyURL: "url", Amount: payment.Amount{Total: -1, Currency: "CNY"}},
			wantErr: errors.New("amount.total must be positive"),
		},
		{
			name: "valid params",
			params: payment.CreateOrderParams{
				Appid: "app", Mchid: "mch", Description: "desc", 
				OutTradeNo: "no", NotifyURL: "url", 
				Amount: payment.Amount{Total: 100, Currency: "CNY"},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := payment.ValidateCreateOrderParams(&tt.params)
			if (err != nil) != (tt.wantErr != nil) {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}
			if err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestCreateOrder_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(payment.CreateOrderResponse{PrepayID: "prepay_123"})
	}))
	defer ts.Close()

	// Override API URL for testing
	originalURL := payment.APIURL
	payment.APIURL = ts.URL
	defer func() { payment.APIURL = originalURL }()

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	credential := &payment.Credential{
		MchAPIv3Key:  "key",
		CertSerialNo: "serial",
		PrivateKey:   privateKey,
	}

	params := &payment.CreateOrderParams{
		Appid:       "appid",
		Mchid:       "mchid",
		Description: "desc",
		OutTradeNo:  "order123",
		NotifyURL:   "https://example.com/notify",
		Amount:      payment.Amount{Total: 100, Currency: "CNY"},
	}

	resp, err := payment.CreateOrder(context.Background(), params, credential)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.PrepayID != "prepay_123" {
		t.Fatalf("expected prepay_id 'prepay_123', got '%s'", resp.PrepayID)
	}
}

func TestCreateOrder_ErrorResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":"INVALID_REQUEST"}`))
	}))
	defer ts.Close()

	originalURL := payment.APIURL
	payment.APIURL = ts.URL
	defer func() { payment.APIURL = originalURL }()

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	credential := &payment.Credential{
		MchAPIv3Key:  "key",
		CertSerialNo: "serial",
		PrivateKey:   privateKey,
	}

	params := &payment.CreateOrderParams{
		Appid:       "appid",
		Mchid:       "mchid",
		Description: "desc",
		OutTradeNo:  "order123",
		NotifyURL:   "https://example.com/notify",
		Amount:      payment.Amount{Total: 100, Currency: "CNY"},
	}

	_, err := payment.CreateOrder(context.Background(), params, credential)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	expectedErr := "wechat pay error: 400 Bad Request, {\"code\":\"INVALID_REQUEST\"}"
	if err.Error() != expectedErr {
		t.Fatalf("expected error: %s, got: %v", expectedErr, err)
	}
}

func TestCreateOrder_ContextCancel(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	originalURL := payment.APIURL
	payment.APIURL = ts.URL
	defer func() { payment.APIURL = originalURL }()

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	credential := &payment.Credential{
		MchAPIv3Key:  "key",
		CertSerialNo: "serial",
		PrivateKey:   privateKey,
	}

	params := &payment.CreateOrderParams{
		Appid:       "appid",
		Mchid:       "mchid",
		Description: "desc",
		OutTradeNo:  "order123",
		NotifyURL:   "https://example.com/notify",
		Amount:      payment.Amount{Total: 100, Currency: "CNY"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := payment.CreateOrder(ctx, params, credential)
	if err == nil {
		t.Fatal("expected context canceled error, got nil")
	}
}
