package sms

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type SMSClient struct {
	secretId  string
	secretKey string
	client    *http.Client
}

func NewClient(secretId, secretKey string) *SMSClient {
	return &SMSClient{
		secretId:  secretId,
		secretKey: secretKey,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *SMSClient) createSignature(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(k + "=" + params[k])
	}
	signString := buf.String()

	mac := hmac.New(sha256.New, []byte(c.secretKey))
	mac.Write([]byte(signString))
	return hex.EncodeToString(mac.Sum(nil))
}

func (c *SMSClient) SendRequest(endpoint string, params map[string]string) ([]byte, error) {
	fullParams := make(map[string]string)
	for k, v := range params {
		fullParams[k] = v
	}
	fullParams["SecretId"] = c.secretId

	signature := c.createSignature(fullParams)
	fullParams["Signature"] = signature

	values := url.Values{}
	for k, v := range fullParams {
		values.Set(k, v)
	}
	reqBody := values.Encode()

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}
	return body, nil
}
