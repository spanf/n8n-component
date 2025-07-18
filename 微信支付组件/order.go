package payment

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

type CreateOrderParams struct {
	Appid       string
	Mchid       string
	Description string
	OutTradeNo  string
	NotifyURL   string
	Amount      Amount
}

type Amount struct {
	Total    int
	Currency string
}

type CreateOrderResponse struct {
	PrepayID string `json:"prepay_id"`
}

type Credential struct {
	MchAPIv3Key  string
	CertSerialNo string
	PrivateKey   *rsa.PrivateKey
}

func CreateOrder(ctx context.Context, params *CreateOrderParams, credential *Credential) (*CreateOrderResponse, error) {
	if err := validateCreateOrderParams(params); err != nil {
		return nil, err
	}

	req, err := buildCreateOrderRequest(params)
	if err != nil {
		return nil, err
	}

	if err := signRequest(req, credential); err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("wechat pay error: %s, %s", resp.Status, body)
	}

	var orderResp CreateOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderResp); err != nil {
		return nil, err
	}

	return &orderResp, nil
}

func buildCreateOrderRequest(params *CreateOrderParams) (*http.Request, error) {
	const apiURL = "https://api.mch.weixin.qq.com/v3/pay/transactions/native"
	requestBody := map[string]interface{}{
		"appid":        params.Appid,
		"mchid":        params.Mchid,
		"description":  params.Description,
		"out_trade_no": params.OutTradeNo,
		"notify_url":   params.NotifyURL,
		"amount": map[string]interface{}{
			"total":    params.Amount.Total,
			"currency": params.Amount.Currency,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func validateCreateOrderParams(params *CreateOrderParams) error {
	if params.Appid == "" {
		return errors.New("appid is required")
	}
	if params.Mchid == "" {
		return errors.New("mchid is required")
	}
	if params.Description == "" {
		return errors.New("description is required")
	}
	if params.OutTradeNo == "" {
		return errors.New("out_trade_no is required")
	}
	if params.NotifyURL == "" {
		return errors.New("notify_url is required")
	}
	if params.Amount.Total <= 0 {
		return errors.New("amount.total must be positive")
	}
	if params.Amount.Currency == "" {
		return errors.New("currency is required")
	}
	return nil
}

func signRequest(req *http.Request, credential *Credential) error {
	const authType = "WECHATPAY2-SHA256-RSA2048"
	timestamp := time.Now().Unix()
	nonce := generateNonce(16)
	body, err := getRequestBody(req)
	if err != nil {
		return err
	}

	signatureStr := fmt.Sprintf("%s\n%s\n%d\n%s\n%s\n",
		req.Method, req.URL.Path, timestamp, nonce, body)

	hashed := sha256.Sum256([]byte(signatureStr))
	signature, err := signWithPrivateKey(hashed[:], credential.PrivateKey)
	if err != nil {
		return err
	}

	authValue := fmt.Sprintf("%s mchid=\"%s\",nonce_str=\"%s\",timestamp=\"%d\",serial_no=\"%s\",signature=\"%s\"",
		authType, credential.Mchid, nonce, timestamp, credential.CertSerialNo, signature)

	req.Header.Set("Authorization", authValue)
	return nil
}

func generateNonce(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func getRequestBody(req *http.Request) (string, error) {
	if req.Body == nil {
		return "", nil
	}
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return string(bodyBytes), nil
}

func signWithPrivateKey(data []byte, privateKey *rsa.PrivateKey) (string, error) {
	hashed := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(nil, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}
