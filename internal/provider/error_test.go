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

	// Test Error() with nil cause
	errNilCause := NewProviderError(ErrKindValidation, "invalid input", nil)
	if errNilCause.Error() != "validation: invalid input" {
		t.Errorf("Unexpected error message: %s", errNilCause.Error())
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

	// Test NewValidationError
	validationErr := NewValidationError("invalid params")
	if validationErr.Message != "invalid params" {
		t.Errorf("Expected message 'invalid params', got %s", validationErr.Message)
	}
	if validationErr.Kind != ErrKindValidation {
		t.Errorf("Expected kind validation, got %s", validationErr.Kind)
	}
	if validationErr.IsRetryable() {
		t.Error("Validation error should not be retryable")
	}

	// Test fallback with non-ProviderError
	stdErr := errors.New("standard error")
	if IsRateLimitError(stdErr) {
		t.Error("IsRateLimitError should be false for standard error")
	}
	if IsAuthError(stdErr) {
		t.Error("IsAuthError should be false for standard error")
	}
	if IsRetryableError(stdErr) {
		t.Error("IsRetryableError should be false for standard error")
	}
	if GetRetryAfter(stdErr) != 0 {
		t.Errorf("GetRetryAfter should be 0 for standard error, got %d", GetRetryAfter(stdErr))
	}
}
