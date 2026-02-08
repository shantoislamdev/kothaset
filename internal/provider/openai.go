// Package provider provides an OpenAI-compatible provider implementation.
package provider

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
)

// OpenAIProvider implements the Provider interface for OpenAI and compatible APIs
type OpenAIProvider struct {
	name   string
	model  string
	apiKey string
	client *openai.Client
}

// NewOpenAIProvider creates a new OpenAI-compatible provider
func NewOpenAIProvider(cfg *Config) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, NewAuthError("API key is required")
	}

	opts := []option.RequestOption{option.WithAPIKey(cfg.APIKey)}

	// Custom base URL for compatible APIs (DeepSeek, vLLM, etc.)
	if cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	}

	// Custom timeout via HTTP client
	if cfg.Timeout > 0 {
		httpClient := &http.Client{
			Timeout: cfg.Timeout,
		}
		opts = append(opts, option.WithHTTPClient(httpClient))
	}

	client := openai.NewClient(opts...)

	return &OpenAIProvider{
		name:   cfg.Name,
		model:  cfg.Model,
		apiKey: cfg.APIKey,
		client: &client,
	}, nil
}

// Generate implements Provider.Generate
func (p *OpenAIProvider) Generate(ctx context.Context, req GenerationRequest) (*GenerationResponse, error) {
	start := time.Now()

	// Convert messages
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(req.Messages)+1)

	// Add system prompt if provided
	if req.SystemPrompt != "" {
		messages = append(messages, openai.SystemMessage(req.SystemPrompt))
	}

	// Add conversation messages
	for _, msg := range req.Messages {
		role := strings.ToLower(msg.Role)
		switch role {
		case "system":
			messages = append(messages, openai.SystemMessage(msg.Content))
		case "user", "human":
			messages = append(messages, openai.UserMessage(msg.Content))
		case "assistant", "ai", "bot":
			messages = append(messages, openai.AssistantMessage(msg.Content))
		default:
			messages = append(messages, openai.UserMessage(msg.Content))
		}
	}

	// Build request parameters
	params := openai.ChatCompletionNewParams{
		Model:       openai.ChatModel(p.model),
		Messages:    messages,
		MaxTokens:   openai.Int(int64(req.MaxTokens)),
		Temperature: openai.Float(req.Temperature),
	}

	// Optional parameters
	if req.TopP > 0 {
		params.TopP = openai.Float(req.TopP)
	}
	if len(req.StopSequences) > 0 {
		params.Stop = openai.ChatCompletionNewParamsStopUnion{OfStringArray: req.StopSequences}
	}
	params.Seed = openai.Int(req.Seed)
	if req.FrequencyPenalty != 0 {
		params.FrequencyPenalty = openai.Float(req.FrequencyPenalty)
	}
	if req.PresencePenalty != 0 {
		params.PresencePenalty = openai.Float(req.PresencePenalty)
	}
	if req.ResponseFormat == "json" {
		params.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &shared.ResponseFormatJSONObjectParam{Type: "json_object"},
		}
	}

	// Make request
	resp, err := p.client.Chat.Completions.New(ctx, params)
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
			PromptTokens:     int(resp.Usage.PromptTokens),
			CompletionTokens: int(resp.Usage.CompletionTokens),
			TotalTokens:      int(resp.Usage.TotalTokens),
		},
		Model:     resp.Model,
		RequestID: resp.ID,
		Latency:   time.Since(start),
	}, nil
}

// GenerateStream implements Provider.GenerateStream
func (p *OpenAIProvider) GenerateStream(ctx context.Context, req GenerationRequest) (<-chan StreamChunk, error) {
	// Convert messages
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(req.Messages)+1)

	if req.SystemPrompt != "" {
		messages = append(messages, openai.SystemMessage(req.SystemPrompt))
	}

	for _, msg := range req.Messages {
		role := strings.ToLower(msg.Role)
		switch role {
		case "system":
			messages = append(messages, openai.SystemMessage(msg.Content))
		case "user", "human":
			messages = append(messages, openai.UserMessage(msg.Content))
		case "assistant", "ai", "bot":
			messages = append(messages, openai.AssistantMessage(msg.Content))
		default:
			messages = append(messages, openai.UserMessage(msg.Content))
		}
	}

	// Build request parameters
	params := openai.ChatCompletionNewParams{
		Model:       openai.ChatModel(p.model),
		Messages:    messages,
		MaxTokens:   openai.Int(int64(req.MaxTokens)),
		Temperature: openai.Float(req.Temperature),
	}

	// Create stream
	stream := p.client.Chat.Completions.NewStreaming(ctx, params)

	// Create output channel
	out := make(chan StreamChunk, 100)

	go func() {
		defer close(out)

		for stream.Next() {
			evt := stream.Current()
			if len(evt.Choices) > 0 {
				choice := evt.Choices[0]
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

		if err := stream.Err(); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
				out <- StreamChunk{Done: true, FinishReason: "stop"}
			} else {
				out <- StreamChunk{Done: true, Error: p.convertError(err)}
			}
		} else {
			out <- StreamChunk{Done: true, FinishReason: "stop"}
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
		"gpt-5.2", "gemini-3", "deepseek-chat-3.2",
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
	if p.apiKey == "" {
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
	_, err := p.client.Models.List(ctx)
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

	var apiErr *openai.Error
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case http.StatusUnauthorized:
			return NewAuthError(apiErr.Message)
		case http.StatusTooManyRequests:
			return NewRateLimitError(apiErr.Message, 60) // Default retry after 60s
		case http.StatusBadRequest:
			if strings.Contains(apiErr.Message, "context_length") {
				return &ProviderError{
					Kind:       ErrKindContextLength,
					Message:    apiErr.Message,
					StatusCode: apiErr.StatusCode,
				}
			}
			return NewValidationError(apiErr.Message)
		case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
			return &ProviderError{
				Kind:       ErrKindServer,
				Message:    apiErr.Message,
				Retryable:  true,
				StatusCode: apiErr.StatusCode,
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
