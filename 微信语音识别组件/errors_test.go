package wechatspeech

import (
	"errors"
	"testing"
)

func TestWechatAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      WechatAPIError
		expected string
	}{
		{
			name:     "standard error",
			err:      WechatAPIError{ErrCode: 40001, ErrMsg: "invalid credential"},
			expected: "wechat api error: 40001 - invalid credential",
		},
		{
			name:     "empty message",
			err:      WechatAPIError{ErrCode: 500},
			expected: "wechat api error: 500 - ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsWechatAPIError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "WechatAPIError type",
			err:      WechatAPIError{ErrCode: 40001},
			expected: true,
		},
		{
			name:     "standard error",
			err:      errors.New("generic error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "wrapped WechatAPIError",
			err:      fmt.Errorf("wrapped: %w", WechatAPIError{ErrCode: 40002}),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsWechatAPIError(tt.err); got != tt.expected {
				t.Errorf("IsWechatAPIError() = %v, want %v", got, tt.expected)
			}
		})
	}
}
