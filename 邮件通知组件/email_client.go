package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type EmailClient struct {
	secretId  string
	secretKey string
	region    string
	client    *http.Client
}

func NewEmailClient(secretId, secretKey, region string) (*EmailClient, error) {
	if secretId == "" || secretKey == "" || region == "" {
		return nil, errors.New("missing required parameters")
	}
	return &EmailClient{
		secretId:  secretId,
		secretKey: secretKey,
		region:    region,
		client:    &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *EmailClient) sendRequest(action string, payload map[string]interface{}) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("https://ses.%s.tencentcloudapi.com", c.region)
	timestamp := time.Now().Unix()
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")

	// 构造规范请求
	canonicalHeaders := fmt.Sprintf("content-type:application/json\nhost:ses.%s.tencentcloudapi.com\nx-tc-action:%s\n", c.region, strings.ToLower(action))
	signedHeaders := "content-type;host;x-tc-action"
	payload["Action"] = action
	payloadBytes, _ := json.Marshal(payload)
	hashedPayload := sha256Hex(payloadBytes)
	canonicalRequest := fmt.Sprintf("POST\n/\n\n%s\n%s\n%s", canonicalHeaders, signedHeaders, hashedPayload)

	// 构造签名字符串
	credentialScope := fmt.Sprintf("%s/ses/tc3_request", date)
	hashedCanonicalRequest := sha256Hex([]byte(canonicalRequest))
	stringToSign := fmt.Sprintf("TC3-HMAC-SHA256\n%d\n%s\n%s", timestamp, credentialScope, hashedCanonicalRequest)

	// 计算签名
	signature := c.calculateSignature(date, stringToSign)

	// 构造Authorization头
	authorization := fmt.Sprintf("TC3-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		c.secretId, credentialScope, signedHeaders, signature)

	// 创建HTTP请求
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authorization)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TC-Action", action)
	req.Header.Set("X-TC-Timestamp", fmt.Sprintf("%d", timestamp))
	req.Header.Set("X-TC-Version", "2020-10-02")
	req.Header.Set("Host", fmt.Sprintf("ses.%s.tencentcloudapi.com", c.region))

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 处理响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", body)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *EmailClient) calculateSignature(date, stringToSign string) string {
	secretDate := hmacSha256([]byte("TC3"+c.secretKey), date)
	secretService := hmacSha256(secretDate, "ses")
	secretSigning := hmacSha256(secretService, "tc3_request")
	return hex.EncodeToString(hmacSha256(secretSigning, stringToSign))
}

func hmacSha256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

func sha256Hex(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
