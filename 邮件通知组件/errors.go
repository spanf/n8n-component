package mail

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
)

var (
	ErrInvalidEmail    = errors.New("invalid email address")
	ErrEmptyRecipients = errors.New("empty recipients")
	ErrEmptySubject    = errors.New("empty subject")
	ErrEmptyBody       = errors.New("empty body")
	ErrInvalidBodyType = errors.New("invalid body type, must be text or html")
)

type APIError struct {
	Code    string
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error: %s - %s", e.Code, e.Message)
}

var retryableCodes = map[string]bool{
	"TooManyRequests":    true,
	"RateLimitExceeded":  true,
	"Timeout":            true,
	"ServiceUnavailable": true,
}

func IsRetryableError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		if code, e := strconv.Atoi(apiErr.Code); e == nil {
			if code >= 500 && code < 600 {
				return true
			}
		} else if retryableCodes[apiErr.Code] {
			return true
		}
	}

	return false
}
