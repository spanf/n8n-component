package sms

import (
	"errors"
	"fmt"
	"testing"

	tcerr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
)

func TestIsTencentError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "TencentCloudSDKError",
			err:  tcerr.NewTencentCloudSDKError("ClientError.NetworkError", "network error", ""),
			want: true,
		},
		{
			name: "WrappedTencentError",
			err:  fmt.Errorf("wrapper: %w", tcerr.NewTencentCloudSDKError("ClientError.NetworkError", "network error", "")),
			want: true,
		},
		{
			name: "OtherError",
			err:  ErrInvalidPhoneNumber,
			want: false,
		},
		{
			name: "TemplateNotFound",
			err:  ErrTemplateNotFound,
			want: false,
		},
		{
			name: "SendFailed",
			err:  ErrSendFailed,
			want: false,
		},
		{
			name: "StandardError",
			err:  errors.New("standard error"),
			want: false,
		},
		{
			name: "NilError",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTencentError(tt.err); got != tt.want {
				t.Errorf("IsTencentError() = %v, want %v", got, tt.want)
			}
		})
	}
}
