package wechatpay

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client 微信支付客户端
type Client struct {
	mchID      string
	serialNo   string
	privateKey *rsa.PrivateKey
}

// NewClient 创建新客户端
func NewClient(mchID, serialNo string, privateKeyPEM []byte) (*Client, error) {
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}

	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := privKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not RSA")
	}

	return &Client{
		mchID:      mchID,
		serialNo:   serialNo,
		privateKey: rsaKey,
	}, nil
}

// doRequest 发送HTTP请求
func (c *Client) doRequest(method, urlStr string, body []byte) ([]byte, error) {
	timestamp := time.Now().Unix()
	nonce := generateNonce(16)
	bodyStr := string(body)

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	signURL := u.Path
	if u.RawQuery != "" {
		signURL += "?" + u.RawQuery
	}

	signature, err := c.signRequest(method, signURL, bodyStr, timestamp, nonce)
	if err != nil {
		return nil, err
	}

	authHeader := c.buildAuthorization("WECHATPAY2-SHA256-RSA2048", signature, nonce, timestamp)

	req, err := http.NewRequest(method, urlStr, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Wechatpay-Timestamp", strconv.FormatInt(timestamp, 10))
	req.Header.Set("Wechatpay-Nonce", nonce)
	req.Header.Set("Wechatpay-Serial", c.serialNo)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}

// signRequest 对请求进行签名
func (c *Client) signRequest(method, urlStr, body string, timestamp int64, nonce string) (string, error) {
	message := fmt.Sprintf("%s\n%s\n%d\n%s\n%s\n", method, urlStr, timestamp, nonce, body)
	hashed := sha256.Sum256([]byte(message))

	signature, err := rsa.SignPKCS1v15(rand.Reader, c.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// buildAuthorization 构建Authorization头
func (c *Client) buildAuthorization(authType, signature, nonce string, timestamp int64) string {
	return fmt.Sprintf(`%s mchid="%s",nonce_str="%s",signature="%s",timestamp="%d",serial_no="%s"`,
		authType, c.mchID, nonce, signature, timestamp, c.serialNo)
}

func generateNonce(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b)
}
