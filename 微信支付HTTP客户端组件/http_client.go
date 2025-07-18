package wechatpay

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

const (
	baseURL     = "https://api.mch.weixin.qq.com"
	authType    = "WECHATPAY2-SHA256-RSA2048"
	contentType = "application/json"
)

type WeChatPayClient struct {
	mchID        string
	certSerialNo string
	privateKey   *rsa.PrivateKey
	httpClient   *http.Client
}

func NewWeChatPayClient(mchID, certSerialNo string, privateKey *rsa.PrivateKey) *WeChatPayClient {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
	}
	
	return &WeChatPayClient{
		mchID:        mchID,
		certSerialNo: certSerialNo,
		privateKey:   privateKey,
		httpClient:   &http.Client{Transport: transport},
	}
}

func (c *WeChatPayClient) SendRequest(method, path, body string) ([]byte, error) {
	fullURL := baseURL + path
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := generateNonce(16)
	signature, err := c.generateSignature(method, fullURL, timestamp, nonce, body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, fullURL, bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf(
		`%s mchid="%s",nonce_str="%s",signature="%s",timestamp="%s",serial_no="%s"`,
		authType, c.mchID, nonce, signature, timestamp, c.certSerialNo,
	))
	req.Header.Set("Accept", contentType)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", "WeChatPay-Go-Client/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error: %d, body: %s", resp.StatusCode, respBody)
	}

	return respBody, nil
}

func (c *WeChatPayClient) generateSignature(method, url, timestamp, nonce, body string) (string, error) {
	signStr := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n", method, url, timestamp, nonce, body)
	hashed := sha256.Sum256([]byte(signStr))

	signature, err := rsa.SignPKCS1v15(nil, c.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

func generateNonce(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
