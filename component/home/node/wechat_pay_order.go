package wechatpay

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
	RefundID    string `json:"refund_id"`
	OutRefundNo string `json:"out_refund_no"`
	Status      string `json:"status"`
	CreateTime  string `json:"create_time"`
}

type amountDetail struct {
	Refund  int64  `json:"refund"`
	Total   int64  `json:"total"`
	Currency string `json:"currency"`
}

type refundRequestBody struct {
	OutTradeNo  string       `json:"out_trade_no"`
	OutRefundNo string       `json:"out_refund_no"`
	Reason      string       `json:"reason"`
	Amount      amountDetail `json:"amount"`
}

func Refund(req RefundRequest, mchID, certSerialNo, privateKey, apiV3Key string) (*RefundResponse, error) {
	// 解析私钥
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		return nil, errors.New("failed to parse private key")
	}
	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key error: %w", err)
	}
	rsaKey, ok := privKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not RSA type")
	}

	// 构造请求体
	reqBody := refundRequestBody{
		OutTradeNo:  req.OutTradeNo,
		OutRefundNo: req.OutRefundNo,
		Reason:      req.Reason,
		Amount: amountDetail{
			Refund:  req.Amount,
			Total:   req.TotalAmount,
			Currency: "CNY",
		},
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request body error: %w", err)
	}

	// 生成随机数和时间戳
	nonce := generateNonce()
	timestamp := time.Now().Unix()

	// 构造签名串
	signatureStr := fmt.Sprintf("POST\n/v3/refund/domestic/refunds\n%d\n%s\n%s\n", 
		timestamp, nonce, string(bodyBytes))
	signature := signWithSHA256(signatureStr, rsaKey)

	// 创建HTTP请求
	url := "https://api.mch.weixin.qq.com/v3/refund/domestic/refunds"
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Authorization", fmt.Sprintf(
		"WECHATPAY2-SHA256-RSA2048 mchid=\"%s\",nonce_str=\"%s\",timestamp=\"%d\",serial_no=\"%s\",signature=\"%s\"",
		mchID, nonce, timestamp, certSerialNo, signature,
	))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request error: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("wechatpay error: %s, body: %s", resp.Status, string(body))
	}

	// 解析响应
	var refundResp RefundResponse
	if err := json.NewDecoder(resp.Body).Decode(&refundResp); err != nil {
		return nil, fmt.Errorf("decode response error: %w", err)
	}

	return &refundResp, nil
}

func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func signWithSHA256(data string, key *rsa.PrivateKey) string {
	hashed := sha256.Sum256([]byte(data))
	signed, _ := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hashed[:])
	return base64.StdEncoding.EncodeToString(signed)
}
