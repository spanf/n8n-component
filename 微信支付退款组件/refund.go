package wxpay

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
	"strconv"
	"time"
)

const refundURL = "https://api.mch.weixin.qq.com/v3/refund/domestic/refunds"

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
	CreateTime  string
}

func Refund(mchID, serialNo string, privateKey *rsa.PrivateKey, req RefundRequest) (*RefundResponse, error) {
	body, err := buildRefundBody(req)
	if err != nil {
		return nil, err
	}

	auth, err := signRequest("POST", refundURL, string(body), mchID, serialNo, privateKey)
	if err != nil {
		return nil, err
	}

	respBody, err := doRequest("POST", refundURL, body, auth)
	if err != nil {
		return nil, err
	}

	return parseRefundResponse(respBody)
}

func buildRefundBody(req RefundRequest) ([]byte, error) {
	data := map[string]interface{}{
		"out_trade_no":  req.OutTradeNo,
		"out_refund_no": req.OutRefundNo,
		"amount": map[string]interface{}{
			"refund":   req.Amount,
			"total":    req.TotalAmount,
			"currency": "CNY",
		},
	}
	if req.Reason != "" {
		data["reason"] = req.Reason
	}
	return json.Marshal(data)
}

func signRequest(method, url, body, mchID, serialNo string, privateKey *rsa.PrivateKey) (string, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := generateNonce(32)

	signStr := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n", method, url, timestamp, nonce, body)
	hashed := sha256.Sum256([]byte(signStr))
	signature, err := rsa.SignPKCS1v15(nil, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("WECHATPAY2-SHA256-RSA2048 mchid=\"%s\",nonce_str=\"%s\",timestamp=\"%s\",serial_no=\"%s\",signature=\"%s\"",
		mchID, nonce, timestamp, serialNo, base64.StdEncoding.EncodeToString(signature)), nil
}

func doRequest(method, url string, body []byte, authHeader string) ([]byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func parseRefundResponse(resp []byte) (*RefundResponse, error) {
	var result struct {
		RefundID    string `json:"refund_id"`
		OutRefundNo string `json:"out_refund_no"`
		Status      string `json:"status"`
		CreateTime  string `json:"create_time"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &RefundResponse{
		RefundID:    result.RefundID,
		OutRefundNo: result.OutRefundNo,
		Status:      result.Status,
		CreateTime:  result.CreateTime,
	}, nil
}

func generateNonce(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
