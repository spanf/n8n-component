package wechatpay

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

type RefundRequest struct {
	OutTradeNo  string
	OutRefundNo string
	Amount      int64
	TotalAmount int64
	Reason      string
}

type RefundResponse struct {
	RefundID    string
	OutRefundNo string
	Status      string
}

type refundRequestBody struct {
	OutTradeNo  string `json:"out_trade_no"`
	OutRefundNo string `json:"out_refund_no"`
	Amount      struct {
		Refund int64 `json:"refund"`
		Total  int64 `json:"total"`
	} `json:"amount"`
	Reason string `json:"reason,omitempty"`
}

type refundResponseBody struct {
	RefundID    string `json:"refund_id"`
	OutRefundNo string `json:"out_refund_no"`
	Status      string `json:"status"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func CreateRefund(req RefundRequest, mchID string, certSerialNo string, privateKey *rsa.PrivateKey) (*RefundResponse, error) {
	const apiURL = "https://api.mch.weixin.qq.com/v3/refund/domestic/refunds"
	
	// 1. 构造请求体
	requestBody := refundRequestBody{
		OutTradeNo:  req.OutTradeNo,
		OutRefundNo: req.OutRefundNo,
		Reason:      req.Reason,
	}
	requestBody.Amount.Refund = req.Amount
	requestBody.Amount.Total = req.TotalAmount

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 2. 生成签名
	timestamp := time.Now().Unix()
	nonce := generateNonce(16)
	signature, err := generateSignature(http.MethodPost, apiURL, timestamp, nonce, jsonBody, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signature: %w", err)
	}

	// 3. 发送HTTP请求
	httpReq, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("WECHATPAY2-SHA256-RSA2048 mchid=\"%s\",nonce_str=\"%s\",signature=\"%s\",timestamp=\"%d\",serial_no=\"%s\"",
		mchID, nonce, signature, timestamp, certSerialNo))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 4. 处理响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp errorResponse
		if json.Unmarshal(respBody, &errResp) == nil {
			return nil, fmt.Errorf("wechatpay error [%s]: %s", errResp.Code, errResp.Message)
		}
		return nil, fmt.Errorf("unexpected http status: %d, body: %s", resp.StatusCode, respBody)
	}

	var wxResp refundResponseBody
	if err := json.Unmarshal(respBody, &wxResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &RefundResponse{
		RefundID:    wxResp.RefundID,
		OutRefundNo: wxResp.OutRefundNo,
		Status:      wxResp.Status,
	}, nil
}

func generateSignature(method, url string, timestamp int64, nonce string, body []byte, privateKey *rsa.PrivateKey) (string, error) {
	message := fmt.Sprintf("%s\n%s\n%d\n%s\n%s\n", method, url, timestamp, nonce, string(body))
	hashed := sha256.Sum256([]byte(message))
	signature, err := rsa.SignPKCS1v15(nil, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

func generateNonce(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
