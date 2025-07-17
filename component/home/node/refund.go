package wechatpay

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"time"
)

var (
	apiKey     string
	httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 15 * time.Second,
	}
)

func SetApiKey(key string) {
	apiKey = key
}

func SetHttpClient(client *http.Client) {
	httpClient = client
}

type RefundRequest struct {
	AppID         string
	MchID         string
	OutTradeNo    string
	OutRefundNo   string
	TotalFee      int
	RefundFee     int
	RefundDesc    string
	NotifyURL     string
	SignType      string
	TransactionID string
}

type xmlRequest struct {
	AppID         string `xml:"appid"`
	MchID         string `xml:"mch_id"`
	NonceStr      string `xml:"nonce_str"`
	Sign          string `xml:"sign"`
	SignType      string `xml:"sign_type,omitempty"`
	TransactionID string `xml:"transaction_id,omitempty"`
	OutTradeNo    string `xml:"out_trade_no,omitempty"`
	OutRefundNo   string `xml:"out_refund_no"`
	TotalFee      int    `xml:"total_fee"`
	RefundFee     int    `xml:"refund_fee"`
	RefundDesc    string `xml:"refund_desc,omitempty"`
	NotifyURL     string `xml:"notify_url,omitempty"`
}

type xmlResponse struct {
	ReturnCode    string `xml:"return_code"`
	ReturnMsg     string `xml:"return_msg"`
	ResultCode    string `xml:"result_code"`
	ErrCode       string `xml:"err_code,omitempty"`
	ErrCodeDes    string `xml:"err_code_des,omitempty"`
	AppID         string `xml:"appid,omitempty"`
	MchID         string `xml:"mch_id,omitempty"`
	NonceStr      string `xml:"nonce_str,omitempty"`
	Sign          string `xml:"sign,omitempty"`
	TransactionID string `xml:"transaction_id,omitempty"`
	OutTradeNo    string `xml:"out_trade_no,omitempty"`
	OutRefundNo   string `xml:"out_refund_no,omitempty"`
	RefundID      string `xml:"refund_id,omitempty"`
	RefundFee     int    `xml:"refund_fee,omitempty"`
	TotalFee      int    `xml:"total_fee,omitempty"`
	CashFee       int    `xml:"cash_fee,omitempty"`
}

func generateNonceStr(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func generateSign(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" && params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var builder strings.Builder
	for _, k := range keys {
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
		builder.WriteString("&")
	}
	builder.WriteString("key=")
	builder.WriteString(apiKey)

	hash := md5.Sum([]byte(builder.String()))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

func InitiateRefund(req RefundRequest) (map[string]interface{}, error) {
	if apiKey == "" {
		return nil, errors.New("API key not configured")
	}
	if req.AppID == "" || req.MchID == "" || req.OutRefundNo == "" || req.TotalFee == 0 || req.RefundFee == 0 {
		return nil, errors.New("missing required parameters")
	}
	if req.OutTradeNo == "" && req.TransactionID == "" {
		return nil, errors.New("either OutTradeNo or TransactionID is required")
	}

	nonceStr := generateNonceStr(32)
	if req.SignType == "" {
		req.SignType = "MD5"
	}

	params := map[string]string{
		"appid":          req.AppID,
		"mch_id":         req.MchID,
		"nonce_str":      nonceStr,
		"sign_type":      req.SignType,
		"out_trade_no":   req.OutTradeNo,
		"out_refund_no":  req.OutRefundNo,
		"total_fee":      fmt.Sprintf("%d", req.TotalFee),
		"refund_fee":     fmt.Sprintf("%d", req.RefundFee),
		"refund_desc":    req.RefundDesc,
		"notify_url":     req.NotifyURL,
		"transaction_id": req.TransactionID,
	}
	sign := generateSign(params)

	xmlReq := xmlRequest{
		AppID:         req.AppID,
		MchID:         req.MchID,
		NonceStr:      nonceStr,
		Sign:          sign,
		SignType:      req.SignType,
		TransactionID: req.TransactionID,
		OutTradeNo:    req.OutTradeNo,
		OutRefundNo:   req.OutRefundNo,
		TotalFee:      req.TotalFee,
		RefundFee:     req.RefundFee,
		RefundDesc:    req.RefundDesc,
		NotifyURL:     req.NotifyURL,
	}

	xmlData, err := xml.Marshal(xmlReq)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Post(
		"https://api.mch.weixin.qq.com/secapi/pay/refund",
		"application/xml",
		bytes.NewBuffer(xmlData),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var xmlResp xmlResponse
	if err := xml.Unmarshal(body, &xmlResp); err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"return_code":    xmlResp.ReturnCode,
		"return_msg":     xmlResp.ReturnMsg,
		"result_code":    xmlResp.ResultCode,
		"err_code":       xmlResp.ErrCode,
		"err_code_des":   xmlResp.ErrCodeDes,
		"appid":          xmlResp.AppID,
		"mch_id":         xmlResp.MchID,
		"nonce_str":      xmlResp.NonceStr,
		"sign":           xmlResp.Sign,
		"transaction_id": xmlResp.TransactionID,
		"out_trade_no":   xmlResp.OutTradeNo,
		"out_refund_no":  xmlResp.OutRefundNo,
		"refund_id":      xmlResp.RefundID,
		"refund_fee":     xmlResp.RefundFee,
		"total_fee":      xmlResp.TotalFee,
		"cash_fee":       xmlResp.CashFee,
	}

	if xmlResp.ReturnCode != "SUCCESS" {
		return result, fmt.Errorf("wechat error: %s", xmlResp.ReturnMsg)
	}
	if xmlResp.ResultCode != "SUCCESS" {
		return result, fmt.Errorf("refund failed: %s - %s", xmlResp.ErrCode, xmlResp.ErrCodeDes)
	}

	return result, nil
}
