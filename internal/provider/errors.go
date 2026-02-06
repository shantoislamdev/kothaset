package provider

import (
	"errors"
	"fmt"
)

// ErrorKind categorizes provider errors
type ErrorKind string

const (
	ErrKindValidation    ErrorKind = "validation"     // Invalid request parameters
	ErrKindAuth          ErrorKind = "auth"           // Authentication failure
	ErrKindRateLimit     ErrorKind = "rate_limit"     // Rate limit exceeded
	ErrKindQuota         ErrorKind = "quota"          // Quota exceeded
	ErrKindNetwork       ErrorKind = "network"        // Network connectivity issue
	ErrKindTimeout       ErrorKind = "timeout"        // Request timeout
	ErrKindServer        ErrorKind = "server"         // Provider server error
	ErrKindContentFilter ErrorKind = "content_filter" // Content filtered
	ErrKindContextLength ErrorKind = "context_length" // Context too long
	ErrKindUnknown       ErrorKind = "unknown"        // Unknown error
)

// ProviderError represents an error from a provider
type ProviderError struct {
	Kind       ErrorKind
	Message    string
	Cause      error
	Retryable  bool
	RetryAfter int // seconds to wait before retry
	StatusCode int
	RequestID  string
}

func (e *ProviderError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Kind, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

func (e *ProviderError) Unwrap() error {
	return e.Cause
}

// IsRetryable returns true if the error can be retried
func (e *ProviderError) IsRetryable() bool {
	return e.Retryable
}

// NewProviderError creates a new provider error
func NewProviderError(kind ErrorKind, message string, cause error) *ProviderError {
	retryable := kind == ErrKindRateLimit || kind == ErrKindTimeout || kind == ErrKindServer
	return &ProviderError{
		Kind:      kind,
		Message:   message,
		Cause:     cause,
		Retryable: retryable,
	}
}

// NewRateLimitError creates a rate limit error with retry-after
func NewRateLimitError(message string, retryAfter int) *ProviderError {
	return &ProviderError{
		Kind:       ErrKindRateLimit,
		Message:    message,
		Retryable:  true,
		RetryAfter: retryAfter,
	}
}

// NewAuthError creates an authentication error
func NewAuthError(message string) *ProviderError {
	return &ProviderError{
		Kind:      ErrKindAuth,
		Message:   message,
		Retryable: false,
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string) *ProviderError {
	return &ProviderError{
		Kind:      ErrKindValidation,
		Message:   message,
		Retryable: false,
	}
}

// Error checking helpers

// IsRateLimitError checks if an error is a rate limit error
func IsRateLimitError(err error) bool {
	var pe *ProviderError
	if errors.As(err, &pe) {
		return pe.Kind == ErrKindRateLimit
	}
	return false
}

// IsAuthError checks if an error is an authentication error
func IsAuthError(err error) bool {
	var pe *ProviderError
	if errors.As(err, &pe) {
		return pe.Kind == ErrKindAuth
	}
	return false
}

// IsRetryableError checks if an error can be retried
func IsRetryableError(err error) bool {
	var pe *ProviderError
	if errors.As(err, &pe) {
		return pe.Retryable
	}
	return false
}

// GetRetryAfter returns the retry-after time, or 0 if not applicable
func GetRetryAfter(err error) int {
	var pe *ProviderError
	if errors.As(err, &pe) {
		return pe.RetryAfter
	}
	return 0
}
