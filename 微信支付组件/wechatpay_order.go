package wechatpay

import (
	"encoding/json"
)

type Payer struct {
	OpenID string `json:"openid"`
}

type Amount struct {
	Total    int    `json:"total"`
	Currency string `json:"currency"`
}

type CreateOrderRequest struct {
	AppID       string `json:"appid"`
	MchID       string `json:"mchid"`
	Description string `json:"description"`
	OutTradeNo  string `json:"out_trade_no"`
	NotifyURL   string `json:"notify_url"`
	Amount      Amount `json:"amount"`
	Payer       *Payer `json:"payer,omitempty"`
}

type CreateOrderResponse struct {
	PrepayID    string `json:"prepay_id"`
	AppID       string `json:"appid"`
	MchID       string `json:"mchid"`
	OutTradeNo  string `json:"out_trade_no"`
	Description string `json:"description"`
}

func CreateOrder(client *WechatPayClient, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	url := "/v3/pay/transactions/jsapi"
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	respBody, err := client.SendRequest("POST", url, reqBody)
	if err != nil {
		return nil, err
	}

	var resp CreateOrderResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
