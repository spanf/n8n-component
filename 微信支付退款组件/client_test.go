package wechatpay_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"path/to/your/package/wechatpay"
)

func generateTestPrivateKey(t *testing.T) *rsa.PrivateKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return privateKey
}

func encodePrivateKeyToPEM(key *rsa.PrivateKey) []byte {
	keyBytes, _ := x509.MarshalPKCS8PrivateKey(key)
	return pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyBytes,
	})
}

func TestNewClient(t *testing.T) {
	t.Run("创建有效客户端", func(t *testing.T) {
		privateKey := generateTestPrivateKey(t)
		pem := encodePrivateKeyToPEM(privateKey)
		
		client, err := wechatpay.NewClient("mch123", "serial001", pem)
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "mch123", client.MchID)
		assert.Equal(t, "serial001", client.SerialNo)
	})
	
	t.Run("无效PEM格式", func(t *testing.T) {
		_, err := wechatpay.NewClient("mch123", "serial001", []byte("INVALID_PEM"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse PEM block")
	})
	
	t.Run("非RSA密钥", func(t *testing.T) {
		invalidPem := pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: []byte("non-rsa-key"),
		})
		_, err := wechatpay.NewClient("mch123", "serial001", invalidPem)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "private key is not RSA")
	})
}

func TestClient_DoRequest(t *testing.T) {
	// 准备测试数据
	privateKey := generateTestPrivateKey(t)
	pem := encodePrivateKeyToPEM(privateKey)
	client, _ := wechatpay.NewClient("mch123", "serial001", pem)
	
	// 创建模拟HTTP服务器
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求头
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.NotEmpty(t, r.Header.Get("Authorization"))
		assert.NotEmpty(t, r.Header.Get("Wechatpay-Timestamp"))
		assert.NotEmpty(t, r.Header.Get("Wechatpay-Nonce"))
		assert.Equal(t, "serial001", r.Header.Get("Wechatpay-Serial"))
		
		// 返回成功响应
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer ts.Close()
	
	// 执行请求
	response, err := client.DoRequest("POST", ts.URL, []byte(`{"amount":100}`))
	require.NoError(t, err)
	assert.JSONEq(t, `{"status": "ok"}`, string(response))
}

func TestDoRequest_HTTPError(t *testing.T) {
	// 准备测试数据
	privateKey := generateTestPrivateKey(t)
	pem := encodePrivateKeyToPEM(privateKey)
	client, _ := wechatpay.NewClient("mch123", "serial001", pem)
	
	// 创建返回500错误的模拟服务器
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()
	
	// 执行请求
	_, err := client.DoRequest("POST", ts.URL, []byte(`{}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP error: 500 Internal Server Error")
}

func TestSignRequest(t *testing.T) {
	privateKey := generateTestPrivateKey(t)
	pem := encodePrivateKeyToPEM(privateKey)
	client, _ := wechatpay.NewClient("mch123", "serial001", pem)
	
	timestamp := time.Now().Unix()
	nonce := "random_nonce"
	
	signature, err := client.SignRequest("POST", "/v3/pay", "request_body", timestamp, nonce)
	require.NoError(t, err)
	assert.NotEmpty(t, signature)
	
	// 验证签名格式为base64
	_, err = base64.StdEncoding.DecodeString(signature)
	assert.NoError(t, err)
}

func TestBuildAuthorization(t *testing.T) {
	privateKey := generateTestPrivateKey(t)
	pem := encodePrivateKeyToPEM(privateKey)
	client, _ := wechatpay.NewClient("mch123", "serial001", pem)
	
	authHeader := client.BuildAuthorization("WECHATPAY2-SHA256-RSA2048", "signature_data", "nonce123", 1234567890)
	expected := `WECHATPAY2-SHA256-RSA2048 mchid="mch123",nonce_str="nonce123",signature="signature_data",timestamp="1234567890",serial_no="serial001"`
	assert.Equal(t, expected, authHeader)
}

func TestGenerateNonce(t *testing.T) {
	nonce := wechatpay.GenerateNonce(16)
	assert.Len(t, nonce, 16)
	
	// 验证字符集
	for _, c := range nonce {
		assert.True(t, (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9'))
	}
}
