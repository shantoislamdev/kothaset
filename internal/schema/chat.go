package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ChatMessage represents a message in a conversation
type ChatMessage struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"` // message content
}

// ChatSchema implements the ShareGPT-style multi-turn conversation format
type ChatSchema struct{}

// NewChatSchema creates a new chat schema
func NewChatSchema() *ChatSchema {
	return &ChatSchema{}
}

func (s *ChatSchema) Name() string        { return "chat" }
func (s *ChatSchema) Style() DatasetStyle { return StyleChat }
func (s *ChatSchema) Version() string     { return "1.0" }

func (s *ChatSchema) Description() string {
	return "ShareGPT-style multi-turn conversations for conversational AI training"
}

func (s *ChatSchema) Fields() []FieldDefinition {
	return []FieldDefinition{
		{
			Name:        "conversations",
			Type:        FieldTypeList,
			Description: "List of messages in the conversation",
			Required:    true,
		},
		{
			Name:        "system",
			Type:        FieldTypeString,
			Description: "Optional system prompt for the conversation",
			Required:    false,
		},
	}
}

func (s *ChatSchema) RequiredFields() []string {
	return []string{"conversations"}
}

func (s *ChatSchema) GeneratePrompt(ctx context.Context, opts PromptOptions) (string, error) {
	var sb strings.Builder

	// Inject user context first (from context.yaml)
	if opts.UserContext != "" {
		sb.WriteString(opts.UserContext)
		sb.WriteString("\n\n")
	} else {
		// Default context if none provided
		sb.WriteString("Generate a high-quality multi-turn conversation between a user and an AI assistant.\n\n")
	}

	if opts.Topic != "" {
		sb.WriteString(fmt.Sprintf("Topic/Context: %s\n", opts.Topic))
	}
	if opts.Category != "" {
		sb.WriteString(fmt.Sprintf("Category: %s\n", opts.Category))
	}
	if opts.Complexity > 0 {
		sb.WriteString(fmt.Sprintf("Conversation depth: %d/5 (more turns for higher values)\n", opts.Complexity))
	}

	sb.WriteString("\n")

	sb.WriteString(`Generate a conversation in the following JSON format:
{
  "system": "Optional system prompt defining the assistant's behavior",
  "conversations": [
    {"role": "user", "content": "User's first message"},
    {"role": "assistant", "content": "Assistant's helpful response"},
    {"role": "user", "content": "User's follow-up"},
    {"role": "assistant", "content": "Assistant's response"}
  ]
}

Requirements:
- Include 2-6 turns (exchanges between user and assistant)
- The conversation should be coherent and natural
- Assistant responses should be helpful, accurate, and engaging
- User messages can include questions, requests, or follow-ups
- Vary the conversation style and complexity`)

	// Inject user instructions (from context.yaml)
	if opts.UserInstruction != "" {
		sb.WriteString("\n\nAdditional Instructions:\n")
		sb.WriteString(opts.UserInstruction)
	}

	sb.WriteString("\n\nRespond with ONLY the JSON object, no additional text.")

	return sb.String(), nil
}

func (s *ChatSchema) ParseResponse(raw string) (*Sample, error) {
	raw = strings.TrimSpace(raw)

	// Clean code blocks
	if strings.HasPrefix(raw, "```json") {
		raw = strings.TrimPrefix(raw, "```json")
		raw = strings.TrimSuffix(raw, "```")
		raw = strings.TrimSpace(raw)
	} else if strings.HasPrefix(raw, "```") {
		raw = strings.TrimPrefix(raw, "```")
		raw = strings.TrimSuffix(raw, "```")
		raw = strings.TrimSpace(raw)
	}

	var data struct {
		System        string        `json:"system"`
		Conversations []ChatMessage `json:"conversations"`
	}

	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	sample := &Sample{
		Fields: map[string]any{
			"system":        data.System,
			"conversations": data.Conversations,
		},
	}

	return sample, nil
}

func (s *ChatSchema) ValidateSample(sample *Sample) error {
	convs, ok := sample.Fields["conversations"]
	if !ok {
		return NewSchemaError(s.Name(), "conversations", "conversations is required")
	}

	convList, ok := convs.([]ChatMessage)
	if !ok {
		// Try to convert from []any
		if rawList, ok := convs.([]any); ok {
			convList = make([]ChatMessage, 0, len(rawList))
			for _, item := range rawList {
				if m, ok := item.(map[string]any); ok {
					cm := ChatMessage{
						Role:    fmt.Sprint(m["role"]),
						Content: fmt.Sprint(m["content"]),
					}
					convList = append(convList, cm)
				}
			}
		} else {
			return NewSchemaError(s.Name(), "conversations", "invalid conversations format")
		}
	}

	if len(convList) < 2 {
		return NewSchemaError(s.Name(), "conversations", "at least 2 messages required")
	}

	// Validate alternating roles
	for i, msg := range convList {
		if msg.Role != "user" && msg.Role != "assistant" && msg.Role != "system" {
			return NewSchemaError(s.Name(), "conversations", fmt.Sprintf("invalid role at index %d: %s", i, msg.Role))
		}
		if msg.Content == "" {
			return NewSchemaError(s.Name(), "conversations", fmt.Sprintf("empty content at index %d", i))
		}
	}

	return nil
}

func (s *ChatSchema) ToJSON(sample *Sample) ([]byte, error) {
	return json.MarshalIndent(sample.Fields, "", "  ")
}

func (s *ChatSchema) ToJSONL(sample *Sample) ([]byte, error) {
	data, err := json.Marshal(sample.Fields)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
