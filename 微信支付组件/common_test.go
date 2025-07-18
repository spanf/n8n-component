package wechatpay

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func generateTestKeyPair() (*rsa.PrivateKey, string) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKey := &privateKey.PublicKey
	pubASN1, _ := x509.MarshalPKIXPublicKey(publicKey)
	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	})
	return privateKey, base64.StdEncoding.EncodeToString(pubBytes)
}

func TestGenerateNonce(t *testing.T) {
	nonce1 := generateNonce()
	nonce2 := generateNonce()
	if nonce1 == nonce2 {
		t.Errorf("Expected unique nonces, got %s and %s", nonce1, nonce2)
	}

	decoded, err := base64.RawURLEncoding.DecodeString(nonce1)
	if err != nil || len(decoded) != 32 {
		t.Errorf("Invalid nonce format: %s", nonce1)
	}
}

func TestGenerateTimestamp(t *testing.T) {
	ts := generateTimestamp()
	tsInt, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		t.Fatalf("Timestamp parse error: %v", err)
	}

	now := time.Now().Unix()
	if tsInt < now-2 || tsInt > now+2 {
		t.Errorf("Timestamp out of range, got %d, expected around %d", tsInt, now)
	}
}

func TestGenerateSignature(t *testing.T) {
	privateKey, _ := generateTestKeyPair()
	cred := &Credential{
		PrivateKey: privateKey,
	}

	req := httptest.NewRequest("POST", "/test?q=1", strings.NewReader(`{"data":"value"}`))
	req.Header.Set("Content-Type", "application/json")

	sig, err := generateSignature(req, cred, "1650000000", "testnonce")
	if err != nil {
		t.Fatalf("Signature generation failed: %v", err)
	}

	sigBytes, err := base64.StdEncoding.DecodeString(sig)
	if err != nil || len(sigBytes) != 256 {
		t.Errorf("Invalid signature format: %s", sig)
	}
}

func TestVerifyResponse(t *testing.T) {
	privateKey, pubKey := generateTestKeyPair()
	cred := &Credential{
		PrivateKey: privateKey,
	}

	body := `{"code":"SUCCESS"}`
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := "testnonce"

	msg := fmt.Sprintf("%s\n%s\n%s\n", timestamp, nonce, body)
	hashed := sha256.Sum256([]byte(msg))
	signature, _ := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	sigStr := base64.StdEncoding.EncodeToString(signature)

	resp := &http.Response{
		Header: http.Header{
			"Wechatpay-Signature": []string{sigStr},
			"Wechatpay-Timestamp": []string{timestamp},
			"Wechatpay-Nonce":     []string{nonce},
			"Wechatpay-Serial":    []string{"1234567890"},
		},
		Body: io.NopCloser(strings.NewReader(body)),
	}

	if err := verifyResponse(resp, cred); err != nil {
		t.Errorf("Valid signature verification failed: %v", err)
	}

	resp.Header.Set("Wechatpay-Signature", "invalid")
	if err := verifyResponse(resp, cred); err == nil {
		t.Error("Invalid signature verification should fail")
	}
}

func TestDoRequest(t *testing.T) {
	privateKey, pubKey := generateTestKeyPair()
	cred := &Credential{
		MchAPIv3Key:  "testkey",
		CertSerialNo: "1234567890",
		PrivateKey:   privateKey,
	}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Wechatpay-Signature", "test")
		w.Header().Set("Wechatpay-Timestamp", "1234567890")
		w.Header().Set("Wechatpay-Nonce", "nonce")
		w.Header().Set("Wechatpay-Serial", "1234567890")
		w.Write([]byte(`{"code":"SUCCESS"}`))
	}))
	defer ts.Close()

	ctx := context.Background()
	req, _ := http.NewRequest("GET", ts.URL, nil)
	resp, err := doRequest(ctx, req, cred)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != `{"code":"SUCCESS"}` {
		t.Errorf("Unexpected response body: %s", body)
	}
}

func TestDoRequestError(t *testing.T) {
	privateKey, _ := generateTestKeyPair()
	cred := &Credential{
		PrivateKey: privateKey,
	}

	tests := []struct {
		name string
		req  *http.Request
	}{
		{"Invalid URL", httptest.NewRequest("GET", "http://invalid.url", nil)},
		{"Read Body Error", func() *http.Request {
			req := httptest.NewRequest("POST", "/", errorReader{})
			req.Body = errorReader{}
			return req
		}()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := doRequest(ctx, tt.req, cred)
			if err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

type errorReader struct{}

func (errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated read error")
}
func (errorReader) Close() error { return nil }
