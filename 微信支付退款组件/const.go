package wechatpayrefund

import "errors"

const (
    apiHost    = "https://api.mch.weixin.qq.com"
    refundPath = "/v3/refund/domestic/refunds"
    queryPath  = "/v3/refund/domestic/refunds/"
    authType   = "WECHATPAY2-SHA256-RSA2048"
)

var (
    ErrInvalidRequest  = errors.New("invalid request")
    ErrRequestFailed   = errors.New("request failed")
    ErrInvalidResponse = errors.New("invalid response")
)
