// Package openai provides an OpenAI-compatible provider implementation.
package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/shantoislamdev/kothaset/internal/config"
)

// OpenAIProvider implements the Provider interface for OpenAI and compatible APIs
type OpenAIProvider struct {
	name   string
	model  string
	client *openai.Client
	config *config.ProviderConfig
}

// NewOpenAIProvider creates a new OpenAI-compatible provider
func NewOpenAIProvider(cfg *config.ProviderConfig) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, NewAuthError("API key is required")
	}

	clientConfig := openai.DefaultConfig(cfg.APIKey)

	// Custom base URL for compatible APIs (DeepSeek, vLLM, etc.)
	if cfg.BaseURL != "" {
		clientConfig.BaseURL = cfg.BaseURL
	}

	// Custom timeout
	if cfg.Timeout.Duration > 0 {
		clientConfig.HTTPClient = &http.Client{
			Timeout: cfg.Timeout.Duration,
		}
	}

	return &OpenAIProvider{
		name:   cfg.Name,
		model:  cfg.Model,
		client: openai.NewClientWithConfig(clientConfig),
		config: cfg,
	}, nil
}

// Generate implements Provider.Generate
func (p *OpenAIProvider) Generate(ctx context.Context, req GenerationRequest) (*GenerationResponse, error) {
	start := time.Now()

	// Convert messages
	messages := make([]openai.ChatCompletionMessage, 0, len(req.Messages)+1)

	// Add system prompt if provided
	if req.SystemPrompt != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: req.SystemPrompt,
		})
	}

	// Add conversation messages
	for _, msg := range req.Messages {
		role := msg.Role
		// Normalize role names
		switch strings.ToLower(role) {
		case "system":
			role = openai.ChatMessageRoleSystem
		case "user", "human":
			role = openai.ChatMessageRoleUser
		case "assistant", "ai", "bot":
			role = openai.ChatMessageRoleAssistant
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: msg.Content,
			Name:    msg.Name,
		})
	}

	// Build request
	chatReq := openai.ChatCompletionRequest{
		Model:       p.model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: float32(req.Temperature),
	}

	// Optional parameters
	if req.TopP > 0 {
		chatReq.TopP = float32(req.TopP)
	}
	if len(req.StopSequences) > 0 {
		chatReq.Stop = req.StopSequences
	}
	if req.Seed != nil {
		seedInt := int(*req.Seed)
		chatReq.Seed = &seedInt
	}
	if req.FrequencyPenalty != 0 {
		chatReq.FrequencyPenalty = float32(req.FrequencyPenalty)
	}
	if req.PresencePenalty != 0 {
		chatReq.PresencePenalty = float32(req.PresencePenalty)
	}
	if req.ResponseFormat == "json" {
		chatReq.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		}
	}

	// Make request
	resp, err := p.client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, p.convertError(err)
	}

	// Extract response
	if len(resp.Choices) == 0 {
		return nil, NewProviderError(ErrKindServer, "no choices returned", nil)
	}

	choice := resp.Choices[0]
	return &GenerationResponse{
		Content:      choice.Message.Content,
		FinishReason: string(choice.FinishReason),
		Usage: TokenUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		Model:     resp.Model,
		RequestID: resp.ID,
		Latency:   time.Since(start),
	}, nil
}

// GenerateStream implements Provider.GenerateStream
func (p *OpenAIProvider) GenerateStream(ctx context.Context, req GenerationRequest) (<-chan StreamChunk, error) {
	// Convert messages
	messages := make([]openai.ChatCompletionMessage, 0, len(req.Messages)+1)

	if req.SystemPrompt != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: req.SystemPrompt,
		})
	}

	for _, msg := range req.Messages {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
			Name:    msg.Name,
		})
	}

	// Build request
	chatReq := openai.ChatCompletionRequest{
		Model:       p.model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: float32(req.Temperature),
		Stream:      true,
	}

	// Create stream
	stream, err := p.client.CreateChatCompletionStream(ctx, chatReq)
	if err != nil {
		return nil, p.convertError(err)
	}

	// Create output channel
	out := make(chan StreamChunk, 100)

	go func() {
		defer close(out)
		defer stream.Close()

		for {
			resp, err := stream.Recv()
			if errors.Is(err, context.Canceled) {
				out <- StreamChunk{Done: true, Error: err}
				return
			}
			if err != nil {
				if err.Error() == "EOF" {
					out <- StreamChunk{Done: true, FinishReason: "stop"}
					return
				}
				out <- StreamChunk{Done: true, Error: p.convertError(err)}
				return
			}

			if len(resp.Choices) > 0 {
				choice := resp.Choices[0]
				chunk := StreamChunk{
					Content: choice.Delta.Content,
				}
				if choice.FinishReason != "" {
					chunk.Done = true
					chunk.FinishReason = string(choice.FinishReason)
				}
				out <- chunk
			}
		}
	}()

	return out, nil
}

// Name implements Provider.Name
func (p *OpenAIProvider) Name() string {
	return p.name
}

// Type implements Provider.Type
func (p *OpenAIProvider) Type() string {
	return "openai"
}

// Model implements Provider.Model
func (p *OpenAIProvider) Model() string {
	return p.model
}

// SupportedModels implements Provider.SupportedModels
func (p *OpenAIProvider) SupportedModels() []string {
	return []string{
		"gpt-4", "gpt-4-turbo", "gpt-4o", "gpt-4o-mini",
		"gpt-3.5-turbo", "gpt-3.5-turbo-16k",
		// Compatible APIs may support other models
	}
}

// SupportsStreaming implements Provider.SupportsStreaming
func (p *OpenAIProvider) SupportsStreaming() bool {
	return true
}

// SupportsBatching implements Provider.SupportsBatching
func (p *OpenAIProvider) SupportsBatching() bool {
	return false // OpenAI batch API has different semantics
}

// Validate implements Provider.Validate
func (p *OpenAIProvider) Validate() error {
	if p.config.APIKey == "" {
		return NewValidationError("API key is required")
	}
	if p.model == "" {
		return NewValidationError("model is required")
	}
	return nil
}

// HealthCheck implements Provider.HealthCheck
func (p *OpenAIProvider) HealthCheck(ctx context.Context) error {
	// Make a minimal request to verify connectivity
	_, err := p.client.ListModels(ctx)
	if err != nil {
		return p.convertError(err)
	}
	return nil
}

// Close implements Provider.Close
func (p *OpenAIProvider) Close() error {
	// OpenAI client doesn't need explicit cleanup
	return nil
}

// convertError converts OpenAI SDK errors to ProviderError
func (p *OpenAIProvider) convertError(err error) error {
	if err == nil {
		return nil
	}

	var apiErr *openai.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.HTTPStatusCode {
		case http.StatusUnauthorized:
			return NewAuthError(apiErr.Message)
		case http.StatusTooManyRequests:
			return NewRateLimitError(apiErr.Message, 60) // Default retry after 60s
		case http.StatusBadRequest:
			if strings.Contains(apiErr.Message, "context_length") {
				return &ProviderError{
					Kind:       ErrKindContextLength,
					Message:    apiErr.Message,
					StatusCode: apiErr.HTTPStatusCode,
				}
			}
			return NewValidationError(apiErr.Message)
		case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
			return &ProviderError{
				Kind:       ErrKindServer,
				Message:    apiErr.Message,
				Retryable:  true,
				StatusCode: apiErr.HTTPStatusCode,
			}
		}
		return NewProviderError(ErrKindUnknown, apiErr.Message, nil)
	}

	// Check for context errors
	if errors.Is(err, context.DeadlineExceeded) {
		return &ProviderError{
			Kind:      ErrKindTimeout,
			Message:   "request timed out",
			Cause:     err,
			Retryable: true,
		}
	}
	if errors.Is(err, context.Canceled) {
		return &ProviderError{
			Kind:      ErrKindNetwork,
			Message:   "request canceled",
			Cause:     err,
			Retryable: false,
		}
	}

	return NewProviderError(ErrKindNetwork, fmt.Sprintf("network error: %v", err), err)
}
