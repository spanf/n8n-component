package wechatpay

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Credential struct {
	MchAPIv3Key string
	CertSerialNo string
	PrivateKey  *rsa.PrivateKey
}

func doRequest(ctx context.Context, req *http.Request, credential *Credential) (*http.Response, error) {
	timestamp := generateTimestamp()
	nonce := generateNonce()
	
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body.Close()
		req.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
		req.ContentLength = int64(len(bodyBytes))
	}

	signature, err := generateSignature(req, credential, timestamp, nonce)
	if err != nil {
		return nil, err
	}

	authHeader := fmt.Sprintf(
		`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",timestamp="%s",serial_no="%s",signature="%s"`,
		credential.CertSerialNo[:10], nonce, timestamp, credential.CertSerialNo, signature,
	)
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "WechatPay-Go/1.0")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		},
	}

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	if err := verifyResponse(resp, credential); err != nil {
		resp.Body.Close()
		return nil, err
	}

	return resp, nil
}

func generateSignature(req *http.Request, cred *Credential, timestamp, nonce string) (string, error) {
	var body string
	if req.Body != nil {
		bodyBytes, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
		body = string(bodyBytes)
	}

	url := req.URL.Path
	if req.URL.RawQuery != "" {
		url += "?" + req.URL.RawQuery
	}

	signStr := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n", 
		req.Method, url, timestamp, nonce, body)

	hashed := sha256.Sum256([]byte(signStr))
	signature, err := rsa.SignPKCS1v15(rand.Reader, cred.PrivateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

func verifyResponse(resp *http.Response, cred *Credential) error {
	signature := resp.Header.Get("Wechatpay-Signature")
	timestamp := resp.Header.Get("Wechatpay-Timestamp")
	nonce := resp.Header.Get("Wechatpay-Nonce")
	serial := resp.Header.Get("Wechatpay-Serial")

	if signature == "" || timestamp == "" || nonce == "" || serial == "" {
		return fmt.Errorf("missing wechatpay headers")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body = io.NopCloser(strings.NewReader(string(body)))

	message := fmt.Sprintf("%s\n%s\n%s\n", timestamp, nonce, string(body))
	hashed := sha256.Sum256([]byte(message))

	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return err
	}

	return rsa.VerifyPKCS1v15(&cred.PrivateKey.PublicKey, crypto.SHA256, hashed[:], sigBytes)
}

func generateNonce() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1<<62))
	return base64.RawURLEncoding.EncodeToString(n.Bytes())
}

func generateTimestamp() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}
