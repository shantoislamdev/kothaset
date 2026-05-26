package provider

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/openai/openai-go/v3"
)

func TestConvertMessages_Empty(t *testing.T) {
	req := GenerationRequest{}
	msgs := convertMessages(req)
	if len(msgs) != 0 {
		t.Errorf("expected 0 messages, got %d", len(msgs))
	}
}

func TestConvertMessages_SystemPrompt(t *testing.T) {
	req := GenerationRequest{
		SystemPrompt: "You are helpful",
		Messages:     []Message{{Role: "user", Content: "hi"}},
	}
	msgs := convertMessages(req)
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
}

func TestConvertMessages_RoleMapping(t *testing.T) {
	tests := []struct {
		role string
	}{
		{"system"},
		{"user"},
		{"human"},
		{"assistant"},
		{"ai"},
		{"bot"},
	}
	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			req := GenerationRequest{
				Messages: []Message{{Role: tt.role, Content: "test"}},
			}
			msgs := convertMessages(req)
			if len(msgs) != 1 {
				t.Fatalf("expected 1 message, got %d", len(msgs))
			}
		})
	}
}

func TestConvertMessages_UnknownRoleDefaultsToUser(t *testing.T) {
	req := GenerationRequest{
		Messages: []Message{{Role: "custom_role", Content: "test"}},
	}
	msgs := convertMessages(req)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
}

func TestConvertMessages_MixedRoles(t *testing.T) {
	req := GenerationRequest{
		SystemPrompt: "system prompt",
		Messages: []Message{
			{Role: "user", Content: "hello"},
			{Role: "assistant", Content: "hi"},
			{Role: "user", Content: "bye"},
		},
	}
	msgs := convertMessages(req)
	if len(msgs) != 4 {
		t.Fatalf("expected 4 messages (1 system + 3 chat), got %d", len(msgs))
	}
}

func TestNewOpenAIProvider_NoAPIKey(t *testing.T) {
	_, err := NewOpenAIProvider(&Config{
		Name:  "test",
		Model: "gpt-4",
	})
	if err == nil {
		t.Fatal("expected error for empty API key")
	}
	var pe *ProviderError
	if !errors.As(err, &pe) {
		t.Fatalf("expected ProviderError, got %T", err)
	}
	if pe.Kind != ErrKindAuth {
		t.Errorf("expected auth error, got %s", pe.Kind)
	}
}

func TestNewOpenAIProvider_Valid(t *testing.T) {
	p, err := NewOpenAIProvider(&Config{
		Name:   "test",
		Model:  "gpt-4",
		APIKey: "sk-test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name() != "test" {
		t.Errorf("expected name 'test', got %s", p.Name())
	}
	if p.Model() != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got %s", p.Model())
	}
	if p.Type() != "openai" {
		t.Errorf("expected type 'openai', got %s", p.Type())
	}
	if !p.SupportsStreaming() {
		t.Error("expected streaming support")
	}
}

func TestNewOpenAIProvider_WithBaseURL(t *testing.T) {
	p, err := NewOpenAIProvider(&Config{
		Name:    "custom",
		Model:   "gpt-4",
		APIKey:  "sk-test",
		BaseURL: "https://custom.api.com/v1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name() != "custom" {
		t.Errorf("expected name 'custom', got %s", p.Name())
	}
}

func TestOpenAIProvider_Validate_NoKey(t *testing.T) {
	p := &OpenAIProvider{
		name:  "test",
		model: "gpt-4",
	}
	err := p.Validate()
	if err == nil {
		t.Fatal("expected error for empty api key")
	}
	var pe *ProviderError
	if !errors.As(err, &pe) {
		t.Fatalf("expected ProviderError, got %T", err)
	}
	if pe.Kind != ErrKindValidation {
		t.Errorf("expected validation error, got %s", pe.Kind)
	}
}

func TestOpenAIProvider_Validate_NoModel(t *testing.T) {
	p := &OpenAIProvider{
		name:   "test",
		apiKey: "sk-test",
	}
	err := p.Validate()
	if err == nil {
		t.Fatal("expected error for empty model")
	}
	var pe *ProviderError
	if !errors.As(err, &pe) {
		t.Fatalf("expected ProviderError, got %T", err)
	}
}

func TestOpenAIProvider_Validate_Valid(t *testing.T) {
	p := &OpenAIProvider{
		name:   "test",
		model:  "gpt-4",
		apiKey: "sk-test",
	}
	if err := p.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestOpenAIProvider_Close(t *testing.T) {
	p := &OpenAIProvider{name: "test"}
	if err := p.Close(); err != nil {
		t.Errorf("Close() should return nil, got %v", err)
	}
}

func TestConvertError_Nil(t *testing.T) {
	p := &OpenAIProvider{}
	if result := p.convertError(nil); result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestConvertError_ContextDeadline(t *testing.T) {
	p := &OpenAIProvider{}
	err := p.convertError(context.DeadlineExceeded)
	var pe *ProviderError
	if !errors.As(err, &pe) {
		t.Fatalf("expected ProviderError, got %T", err)
	}
	if pe.Kind != ErrKindTimeout {
		t.Errorf("expected timeout, got %s", pe.Kind)
	}
	if !pe.Retryable {
		t.Error("timeout should be retryable")
	}
}

func TestConvertError_ContextCanceled(t *testing.T) {
	p := &OpenAIProvider{}
	err := p.convertError(context.Canceled)
	var pe *ProviderError
	if !errors.As(err, &pe) {
		t.Fatalf("expected ProviderError, got %T", err)
	}
	if pe.Kind != ErrKindNetwork {
		t.Errorf("expected network, got %s", pe.Kind)
	}
	if pe.Retryable {
		t.Error("canceled should not be retryable")
	}
}

func TestConvertError_GenericError(t *testing.T) {
	p := &OpenAIProvider{}
	err := p.convertError(errors.New("connection refused"))
	var pe *ProviderError
	if !errors.As(err, &pe) {
		t.Fatalf("expected ProviderError, got %T", err)
	}
	if pe.Kind != ErrKindNetwork {
		t.Errorf("expected network, got %s", pe.Kind)
	}
}

func TestConvertError_APIError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		message    string
		wantKind   ErrorKind
		wantRetry  bool
	}{
		{"unauthorized", http.StatusUnauthorized, "invalid key", ErrKindAuth, false},
		{"rate_limit", http.StatusTooManyRequests, "rate limited", ErrKindRateLimit, true},
		{"bad_request", http.StatusBadRequest, "invalid param", ErrKindValidation, false},
		{"context_length", http.StatusBadRequest, "context_length exceeded", ErrKindContextLength, false},
		{"server_500", http.StatusInternalServerError, "internal error", ErrKindServer, true},
		{"server_502", http.StatusBadGateway, "bad gateway", ErrKindServer, true},
		{"server_503", http.StatusServiceUnavailable, "unavailable", ErrKindServer, true},
		{"unknown_418", http.StatusTeapot, "teapot", ErrKindUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &OpenAIProvider{}
			apiErr := &openai.Error{
				StatusCode: tt.statusCode,
				Message:    tt.message,
			}
			err := p.convertError(apiErr)
			var pe *ProviderError
			if !errors.As(err, &pe) {
				t.Fatalf("expected ProviderError, got %T", err)
			}
			if pe.Kind != tt.wantKind {
				t.Errorf("expected kind %s, got %s", tt.wantKind, pe.Kind)
			}
			if pe.Retryable != tt.wantRetry {
				t.Errorf("expected retryable=%v, got %v", tt.wantRetry, pe.Retryable)
			}
		})
	}
}
