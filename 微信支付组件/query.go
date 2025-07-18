package wechatpay

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Credential struct {
	MchID      string
	SerialNo   string
	PrivateKey *rsa.PrivateKey
}

type QueryOrderResponse struct {
	Appid          string `json:"appid"`
	Mchid          string `json:"mchid"`
	OutTradeNo     string `json:"out_trade_no"`
	TransactionID  string `json:"transaction_id"`
	TradeType      string `json:"trade_type"`
	TradeState     string `json:"trade_state"`
	TradeStateDesc string `json:"trade_state_desc"`
	Amount         struct {
		Total         int    `json:"total"`
		Currency      string `json:"currency"`
		PayerTotal    int    `json:"payer_total"`
		PayerCurrency string `json:"payer_currency"`
	} `json:"amount"`
	Payer struct {
		Openid string `json:"openid"`
	} `json:"payer"`
}

func QueryOrder(ctx context.Context, mchid string, outTradeNo string, credential *Credential) (*QueryOrderResponse, error) {
	req, err := buildQueryOrderRequest(mchid, outTradeNo)
	if err != nil {
		return nil, err
	}

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonce := generateNonce(32)
	signature, err := generateSignature(req, timestamp, nonce, credential.PrivateKey)
	if err != nil {
		return nil, err
	}

	authHeader := fmt.Sprintf("WECHATPAY2-SHA256-RSA2048 mchid=\"%s\",nonce_str=\"%s\",signature=\"%s\",timestamp=\"%s\",serial_no=\"%s\"",
		credential.MchID, nonce, signature, timestamp, credential.SerialNo)
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "WechatPay-Go-Component")

	client := &http.Client{}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseQueryOrderResponse(resp)
}

func buildQueryOrderRequest(mchid string, outTradeNo string) (*http.Request, error) {
	url := fmt.Sprintf("https://api.mch.weixin.qq.com/v3/pay/transactions/out-trade-no/%s?mchid=%s",
		outTradeNo, mchid)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func parseQueryOrderResponse(resp *http.Response) (*QueryOrderResponse, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s, body: %s", resp.Status, string(body))
	}

	var orderResp QueryOrderResponse
	if err := json.Unmarshal(body, &orderResp); err != nil {
		return nil, err
	}
	return &orderResp, nil
}

func generateNonce(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

func generateSignature(req *http.Request, timestamp, nonce string, privateKey *rsa.PrivateKey) (string, error) {
	signData := buildSignData(req, timestamp, nonce)
	hashed := sha256.Sum256([]byte(signData))
	signature, err := rsa.SignPKCS1v15(nil, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

func buildSignData(req *http.Request, timestamp, nonce string) string {
	urlParts := strings.Split(req.URL.String(), "://")
	path := urlParts[1]
	if len(urlParts) > 1 {
		path = urlParts[1][strings.Index(urlParts[1], "/"):]
	}
	return fmt.Sprintf("%s\n%s\n%s\n%s\n\n", req.Method, path, timestamp, nonce)
}
