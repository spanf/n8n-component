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
	"strconv"
	"time"
)

// RefundQueryRequest 退款查询请求参数
type RefundQueryRequest struct {
	OutRefundNo string // 商户退款单号
}

// RefundQueryResponse 退款查询响应
type RefundQueryResponse struct {
	RefundID    string `json:"refund_id"`    // 微信退款单号
	OutRefundNo string `json:"out_refund_no"` // 商户退款单号
	Status      string `json:"status"`       // 退款状态
	Amount      int64  `json:"amount"`       // 退款金额
	SuccessTime string `json:"success_time"` // 退款成功时间
}

// QueryRefund 查询退款状态
func QueryRefund(req RefundQueryRequest, mchID string, certSerialNo string, privateKey *rsa.PrivateKey) (*RefundQueryResponse, error) {
	// 1. 构造请求URL
	url := "https://api.mch.weixin.qq.com/v3/refund/domestic/refunds/" + req.OutRefundNo

	// 2. 生成签名
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := generateNonce(32)
	signature, err := generateSignature("GET", url, timestamp, nonce, "", mchID, certSerialNo, privateKey)
	if err != nil {
		return nil, fmt.Errorf("生成签名失败: %v", err)
	}

	// 3. 发送HTTP请求
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	httpReq.Header.Set("Authorization", signature)
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", "WechatPay-Go-Client")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 4. 解析响应
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("微信API返回错误: 状态码 %d, 响应: %s", resp.StatusCode, string(body))
	}

	var result RefundQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &result, nil
}

// generateNonce 生成随机字符串
func generateNonce(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// generateSignature 生成微信支付V3签名
func generateSignature(method, url, timestamp, nonce, body, mchID, serialNo string, privateKey *rsa.PrivateKey) (string, error) {
	// 构造签名串
	signStr := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n", method, url, timestamp, nonce, body)
	
	// 计算SHA256哈希
	hashed := sha256.Sum256([]byte(signStr))
	
	// 使用私钥签名
	signature, err := rsa.SignPKCS1v15(nil, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}
	
	// Base64编码签名
	encodedSig := base64.StdEncoding.EncodeToString(signature)
	
	// 构造Authorization头
	return fmt.Sprintf("WECHATPAY2-SHA256-RSA2048 mchid=\"%s\",nonce_str=\"%s\",signature=\"%s\",timestamp=\"%s\",serial_no=\"%s\"",
		mchID, nonce, encodedSig, timestamp, serialNo), nil
}
