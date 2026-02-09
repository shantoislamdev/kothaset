package provider

import (
	"errors"
	"testing"
)

func TestProviderError(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewProviderError(ErrKindTimeout, "request timed out", cause)

	if err.Error() != "timeout: request timed out: underlying error" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
	if !err.IsRetryable() {
		t.Error("Timeout should be retryable")
	}

	unwrapped := errors.Unwrap(err)
	if unwrapped != cause {
		t.Error("Unwrap should return cause")
	}
}

func TestErrorHelpers(t *testing.T) {
	rateLimitErr := NewRateLimitError("too many requests", 10)
	authErr := NewAuthError("bad token")

	if !IsRateLimitError(rateLimitErr) {
		t.Error("IsRateLimitError failed")
	}
	if IsRateLimitError(authErr) {
		t.Error("IsRateLimitError should be false for auth error")
	}

	if !IsAuthError(authErr) {
		t.Error("IsAuthError failed")
	}

	if !IsRetryableError(rateLimitErr) {
		t.Error("RateLimit should be retryable")
	}
	if IsRetryableError(authErr) {
		t.Error("Auth error should not be retryable")
	}

	if GetRetryAfter(rateLimitErr) != 10 {
		t.Errorf("Expected RetryAfter 10, got %d", GetRetryAfter(rateLimitErr))
	}
}
