package schema

import (
	"context"
	"strings"
	"testing"
)

func TestChatSchema_GeneratePrompt(t *testing.T) {
	s := NewChatSchema()
	ctx := context.Background()

	opts := PromptOptions{
		Topic:           "Science",
		UserInstruction: "Make it funny",
		Complexity:      3,
	}

	prompt, err := s.GeneratePrompt(ctx, opts)
	if err != nil {
		t.Fatalf("GeneratePrompt failed: %v", err)
	}

	if !strings.Contains(prompt, "Topic/Context: Science") {
		t.Error("Prompt missing topic")
	}
	if !strings.Contains(prompt, "Make it funny") {
		t.Error("Prompt missing user instruction")
	}
	if !strings.Contains(prompt, "Conversation depth: 3/5") {
		t.Error("Prompt missing complexity")
	}
}

func TestChatSchema_ParseResponse(t *testing.T) {
	s := NewChatSchema()

	// Test valid JSON response
	validJSON := `{
		"system": "You are a helpful assistant",
		"conversations": [
			{"role": "user", "content": "Hello"},
			{"role": "assistant", "content": "Hi there"}
		]
	}`

	sample, err := s.ParseResponse(validJSON)
	if err != nil {
		t.Fatalf("ParseResponse failed: %v", err)
	}

	if sample.GetString("system") != "You are a helpful assistant" {
		t.Error("Parsed incorrect system prompt")
	}

	convs := sample.Fields["conversations"].([]ChatMessage)
	if len(convs) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(convs))
	}

	// Test with markdown code blocks
	markdownJSON := "```json\n" + validJSON + "\n```"
	sample, err = s.ParseResponse(markdownJSON)
	if err != nil {
		t.Fatalf("ParseResponse failed with markdown: %v", err)
	}
}

func TestChatSchema_ValidateSample(t *testing.T) {
	s := NewChatSchema()

	// Valid sample (using []any as JSON deserialization produces)
	validSample := &Sample{
		Fields: map[string]any{
			"conversations": []any{
				map[string]any{"role": "user", "content": "A"},
				map[string]any{"role": "assistant", "content": "B"},
			},
		},
	}
	if err := s.ValidateSample(validSample); err != nil {
		t.Errorf("ValidateSample failed for valid sample: %v", err)
	}

	// Invalid: missing conversations
	invalidSample := &Sample{
		Fields: map[string]any{},
	}
	if err := s.ValidateSample(invalidSample); err == nil {
		t.Error("ValidateSample should fail for missing conversations")
	}

	// Valid: ParseResponse produces []ChatMessage, ValidateSample must accept it
	parseSample, err := s.ParseResponse(`{
		"conversations": [
			{"role": "user", "content": "Hello"},
			{"role": "assistant", "content": "Hi there"}
		]
	}`)
	if err != nil {
		t.Fatalf("ParseResponse failed: %v", err)
	}
	if err := s.ValidateSample(parseSample); err != nil {
		t.Errorf("ValidateSample failed for ParseResponse output: %v", err)
	}

	// Invalid: not enough messages
	shortSample := &Sample{
		Fields: map[string]any{
			"conversations": []any{
				map[string]any{"role": "user", "content": "Only one"},
			},
		},
	}
	if err := s.ValidateSample(shortSample); err == nil {
		t.Error("ValidateSample should fail for < 2 messages")
	}
}
