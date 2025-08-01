package wechatspeech

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidFile      = errors.New("invalid voice file")
	ErrUploadFailed     = errors.New("voice upload failed")
	ErrRecognitionFailed = errors.New("voice recognition failed")
	ErrInvalidResponse  = errors.New("invalid response from wechat")
)

type WechatAPIError struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (e WechatAPIError) Error() string {
	return fmt.Sprintf("wechat api error: %d - %s", e.ErrCode, e.ErrMsg)
}

func IsWechatAPIError(err error) bool {
	_, ok := err.(WechatAPIError)
	return ok
}
