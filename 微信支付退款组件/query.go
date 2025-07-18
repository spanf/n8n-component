package wechatpay

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	baseURL     = "https://api.mch.weixin.qq.com/v3/refund/domestic/refunds"
	authScheme  = "WECHATPAY2-SHA256-RSA2048"
	contentType = "application/json"
)

type QueryRequest struct {
	OutRefundNo string
}

type QueryResponse struct {
	RefundID     string
	OutRefundNo  string
	Status       string
	Amount       int64
	SuccessTime  string
	UserReceived string
}

func QueryRefund(mchID, serialNo string, privateKey *rsa.PrivateKey, req QueryRequest) (*QueryResponse, error) {
	queryURL := buildQueryURL(req.OutRefundNo)
	authHeader, err := signRequest("GET", queryURL, "", mchID, serialNo, privateKey)
	if err != nil {
		return nil, err
	}

	respBody, err := doRequest("GET", queryURL, nil, authHeader)
	if err != nil {
		return nil, err
	}

	return parseQueryResponse(respBody)
}

func buildQueryURL(outRefundNo string) string {
	escapedNo := url.PathEscape(outRefundNo)
	return fmt.Sprintf("%s/%s", baseURL, escapedNo)
}

func signRequest(method, urlStr, body string, mchID, serialNo string, privateKey *rsa.PrivateKey) (string, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := generateNonce(16)
	
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	signURL := u.RequestURI()

	signatureStr := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n", 
		method, signURL, timestamp, nonce, body)

	hashed := sha256.Sum256([]byte(signatureStr))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}
	signatureBase64 := base64.StdEncoding.EncodeToString(signature)

	return fmt.Sprintf(`%s mchid="%s",nonce_str="%s",signature="%s",timestamp="%s",serial_no="%s"`,
		authScheme, mchID, nonce, signatureBase64, timestamp, serialNo), nil
}

func generateNonce(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)[:length]
}

func doRequest(method, url string, body []byte, authHeader string) ([]byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", contentType)

	client := &http.Client{Timeout: 10 * time.Second}
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

func parseQueryResponse(resp []byte) (*QueryResponse, error) {
	var rawResp struct {
		RefundID         string `json:"refund_id"`
		OutRefundNo      string `json:"out_refund_no"`
		Status           string `json:"status"`
		Amount           int64  `json:"amount"`
		SuccessTime      string `json:"success_time"`
		UserReceivedAcct string `json:"user_received_account"`
	}

	if err := json.Unmarshal(resp, &rawResp); err != nil {
		return nil, err
	}

	return &QueryResponse{
		RefundID:     rawResp.RefundID,
		OutRefundNo:  rawResp.OutRefundNo,
		Status:       strings.ToUpper(rawResp.Status),
		Amount:       rawResp.Amount,
		SuccessTime:  rawResp.SuccessTime,
		UserReceived: rawResp.UserReceivedAcct,
	}, nil
}
