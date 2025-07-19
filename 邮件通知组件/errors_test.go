package mail

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "context deadline exceeded",
			err:  context.DeadlineExceeded,
			want: true,
		},
		{
			name: "net timeout error",
			err:  &net.DNSError{IsTimeout: true},
			want: true,
		},
		{
			name: "API error with 500 status",
			err:  &APIError{Code: "500"},
			want: true,
		},
		{
			name: "API error with 599 status",
			err:  &APIError{Code: "599"},
			want: true,
		},
		{
			name: "API error with retryable code",
			err:  &APIError{Code: "TooManyRequests"},
			want: true,
		},
		{
			name: "API error with 400 status",
			err:  &APIError{Code: "400"},
			want: false,
		},
		{
			name: "API error with non-retryable code",
			err:  &APIError{Code: "NotFound"},
			want: false,
		},
		{
			name: "non-timeout net error",
			err:  &net.DNSError{IsTimeout: false},
			want: false,
		},
		{
			name: "standard error",
			err:  errors.New("generic error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "wrapped context error",
			err:  fmt.Errorf("wrapped: %w", context.DeadlineExceeded),
			want: true,
		},
		{
			name: "wrapped API error",
			err:  fmt.Errorf("wrapped: %w", &APIError{Code: "ServiceUnavailable"}),
			want: true,
		},
		{
			name: "custom timeout error",
			err:  &customTimeoutError{},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRetryableError(tt.err)
			if got != tt.want {
				t.Errorf("IsRetryableError() = %v, want %v", got, tt.want)
			}
		})
	}
}

type customTimeoutError struct{}

func (e *customTimeoutError) Error() string   { return "timeout" }
func (e *customTimeoutError) Timeout() bool   { return true }
func (e *customTimeoutError) Temporary() bool { return true }
