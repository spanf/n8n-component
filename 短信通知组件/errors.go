package sms

import (
	"errors"
	tcerr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
)

var (
	ErrInvalidPhoneNumber = errors.New("invalid phone number")
	ErrTemplateNotFound   = errors.New("template not found")
	ErrSendFailed         = errors.New("send sms failed")
)

func IsTencentError(err error) bool {
	var sdkErr *tcerr.TencentCloudSDKError
	return errors.As(err, &sdkErr)
}
