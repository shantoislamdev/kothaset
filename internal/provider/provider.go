// Package provider implements LLM provider abstractions for KothaSet.
package provider

import (
	"context"
	"time"
)

// Provider is the interface that all LLM providers must implement
type Provider interface {
	// Generate creates a completion for the given request
	Generate(ctx context.Context, req GenerationRequest) (*GenerationResponse, error)

	// Metadata
	Name() string
	Type() string
	Model() string
	SupportsStreaming() bool

	// Lifecycle
	Validate() error
	HealthCheck(ctx context.Context) error
	Close() error
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`           // system, user, assistant
	Content string `json:"content"`        // message content
	Name    string `json:"name,omitempty"` // optional name for multi-agent
}

// GenerationRequest contains all parameters for a generation request
type GenerationRequest struct {
	// Messages is the conversation history
	Messages []Message `json:"messages"`

	// SystemPrompt overrides the system message (convenience)
	SystemPrompt string `json:"system_prompt,omitempty"`

	// Temperature controls randomness (0-2)
	Temperature float64 `json:"temperature"`

	// MaxTokens limits the response length
	MaxTokens int `json:"max_tokens"`

	// TopP for nucleus sampling (0-1)
	TopP float64 `json:"top_p,omitempty"`

	// StopSequences are strings that stop generation
	StopSequences []string `json:"stop,omitempty"`

	// Seed for reproducibility
	Seed *int64 `json:"seed,omitempty"`

	// FrequencyPenalty reduces repetition (-2 to 2)
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`

	// PresencePenalty encourages new topics (-2 to 2)
	PresencePenalty float64 `json:"presence_penalty,omitempty"`

	// ResponseFormat for structured output (e.g., "json")
	ResponseFormat string `json:"response_format,omitempty"`

	// Metadata for tracking/logging
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// GenerationResponse contains the result of a generation request
type GenerationResponse struct {
	// Content is the generated text
	Content string `json:"content"`

	// FinishReason indicates why generation stopped
	FinishReason string `json:"finish_reason"`

	// Usage contains token counts
	Usage TokenUsage `json:"usage"`

	// Model is the actual model used
	Model string `json:"model"`

	// RequestID from the provider (for debugging)
	RequestID string `json:"request_id,omitempty"`

	// Latency of the request
	Latency time.Duration `json:"latency"`

	// Cached indicates if this was a cached response
	Cached bool `json:"cached,omitempty"`
}

// TokenUsage contains token consumption information
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk represents a piece of a streaming response
type StreamChunk struct {
	// Content is the text delta
	Content string `json:"content"`

	// Done indicates the stream is complete
	Done bool `json:"done"`

	// FinishReason when Done is true
	FinishReason string `json:"finish_reason,omitempty"`

	// Usage when Done is true
	Usage *TokenUsage `json:"usage,omitempty"`

	// Error if something went wrong
	Error error `json:"error,omitempty"`
}
