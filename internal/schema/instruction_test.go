package schema

import (
	"context"
	"strings"
	"testing"
)

func TestInstructionSchema_GeneratePrompt(t *testing.T) {
	s := NewInstructionSchema()
	ctx := context.Background()

	// Test few-shot examples
	examples := []*Sample{
		{
			Fields: map[string]any{
				"instruction": "Q1",
				"output":      "A1",
			},
		},
	}

	opts := PromptOptions{
		Topic:       "Coding",
		NumExamples: 1,
		Examples:    examples,
	}

	prompt, err := s.GeneratePrompt(ctx, opts)
	if err != nil {
		t.Fatalf("GeneratePrompt failed: %v", err)
	}

	if !strings.Contains(prompt, "Topic/Seed: Coding") {
		t.Error("Prompt missing topic")
	}
	if !strings.Contains(prompt, "Example 1") {
		t.Error("Prompt missing example header")
	}
	if !strings.Contains(prompt, "Instruction: Q1") {
		t.Error("Prompt missing example content")
	}
}

func TestInstructionSchema_ParseResponse(t *testing.T) {
	s := NewInstructionSchema()

	validJSON := `{
		"instruction": "Write a hello world program",
		"input": "python",
		"output": "print('Hello, World!')"
	}`

	sample, err := s.ParseResponse(validJSON)
	if err != nil {
		t.Fatalf("ParseResponse failed: %v", err)
	}

	if sample.GetString("instruction") != "Write a hello world program" {
		t.Error("Incorrect instruction parsed")
	}
	if sample.GetString("input") != "python" {
		t.Error("Incorrect input parsed")
	}
}

func TestInstructionSchema_ValidateSample(t *testing.T) {
	s := NewInstructionSchema()

	tests := []struct {
		name    string
		sample  *Sample
		wantErr bool
	}{
		{
			name: "valid sample",
			sample: &Sample{
				Fields: map[string]any{
					"instruction": "Write a function to add two numbers",
					"output":      "def add(a, b): return a + b",
				},
			},
			wantErr: false,
		},
		{
			name: "missing instruction",
			sample: &Sample{
				Fields: map[string]any{
					"output": "result",
				},
			},
			wantErr: true,
		},
		{
			name: "short instruction",
			sample: &Sample{
				Fields: map[string]any{
					"instruction": "short",
					"output":      "valid output length here",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateSample(tt.sample)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSample() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
