package wechatpay

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type WechatPayClient struct {
	mchID      string
	apiKey     string
	privateKey *rsa.PrivateKey
	serialNo   string
}

func NewClient(mchID, apiKey, serialNo string, privateKey *rsa.PrivateKey) *WechatPayClient {
	return &WechatPayClient{
		mchID:      mchID,
		apiKey:     apiKey,
		privateKey: privateKey,
		serialNo:   serialNo,
	}
}

func (c *WechatPayClient) sendRequest(method, rawURL string, body []byte) (*http.Response, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	pathWithQuery := parsedURL.Path
	if parsedURL.RawQuery != "" {
		pathWithQuery += "?" + parsedURL.RawQuery
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce, err := generateNonce()
	if err != nil {
		return nil, fmt.Errorf("generate nonce failed: %w", err)
	}

	signature, err := c.signRequest(method, pathWithQuery, timestamp, nonce, body)
	if err != nil {
		return nil, fmt.Errorf("sign request failed: %w", err)
	}

	authHeader := fmt.Sprintf(
		`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",signature="%s",timestamp="%s",serial_no="%s"`,
		c.mchID, nonce, signature, timestamp, c.serialNo,
	)

	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, rawURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "WechatPay-Go-Client")

	client := &http.Client{}
	return client.Do(req)
}

func (c *WechatPayClient) signRequest(method, path, timestamp, nonce string, body []byte) (string, error) {
	var bodyStr string
	if len(body) > 0 {
		bodyStr = string(body)
	}
	signMessage := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n", method, path, timestamp, nonce, bodyStr)

	hashed := sha256.Sum256([]byte(signMessage))
	signature, err := rsa.SignPKCS1v15(rand.Reader, c.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

func generateNonce() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}
