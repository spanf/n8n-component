package wechatpay

import (
	"errors"
	"fmt"
)

type QueryOrderRequest struct {
	TransactionID string `json:"transaction_id,omitempty"`
	OutTradeNo    string `json:"out_trade_no,omitempty"`
	MchID         string `json:"mchid"`
}

type QueryOrderResponse struct {
	AppID          string `json:"appid"`
	MchID          string `json:"mchid"`
	OutTradeNo     string `json:"out_trade_no"`
	TransactionID  string `json:"transaction_id"`
	TradeState     string `json:"trade_state"`
	TradeStateDesc string `json:"trade_state_desc"`
	TradeType      string `json:"trade_type"`
	BankType       string `json:"bank_type"`
	Amount         struct {
		Total    int    `json:"total"`
		Currency string `json:"currency"`
	} `json:"amount"`
	Payer struct {
		Openid string `json:"openid"`
	} `json:"payer"`
	SuccessTime string `json:"success_time"`
}

func QueryOrder(client *WechatPayClient, req *QueryOrderRequest) (*QueryOrderResponse, error) {
	if req.MchID == "" {
		return nil, errors.New("mchid is required")
	}
	if req.TransactionID == "" && req.OutTradeNo == "" {
		return nil, errors.New("transaction_id or out_trade_no is required")
	}

	var url string
	if req.TransactionID != "" {
		url = fmt.Sprintf("/v3/pay/transactions/id/%s?mchid=%s", req.TransactionID, req.MchID)
	} else {
		url = fmt.Sprintf("/v3/pay/transactions/out-trade-no/%s?mchid=%s", req.OutTradeNo, req.MchID)
	}

	response, err := client.sendRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var result QueryOrderResponse
	if err := client.decodeResponse(response, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
