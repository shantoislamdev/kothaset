package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// InstructionSchema implements the Alpaca-style instruction-response format
type InstructionSchema struct{}

// NewInstructionSchema creates a new instruction schema
func NewInstructionSchema() *InstructionSchema {
	return &InstructionSchema{}
}

func (s *InstructionSchema) Name() string        { return "instruction" }
func (s *InstructionSchema) Style() DatasetStyle { return StyleInstruction }
func (s *InstructionSchema) Version() string     { return "1.0" }

func (s *InstructionSchema) Description() string {
	return "Alpaca-style instruction-response pairs for instruction following tasks"
}

func (s *InstructionSchema) Fields() []FieldDefinition {
	return []FieldDefinition{
		{
			Name:        "instruction",
			Type:        FieldTypeString,
			Description: "The task instruction or question",
			Required:    true,
		},
		{
			Name:        "input",
			Type:        FieldTypeString,
			Description: "Optional additional context or input for the task",
			Required:    false,
			Default:     "",
		},
		{
			Name:        "output",
			Type:        FieldTypeString,
			Description: "The expected response or answer",
			Required:    true,
		},
	}
}

func (s *InstructionSchema) RequiredFields() []string {
	return []string{"instruction", "output"}
}

func (s *InstructionSchema) GeneratePrompt(ctx context.Context, opts PromptOptions) (string, error) {
	var sb strings.Builder

	// Inject user context first (from context.yaml)
	if opts.UserContext != "" {
		sb.WriteString(opts.UserContext)
		sb.WriteString("\n\n")
	} else {
		// Default context if none provided
		sb.WriteString("Generate a high-quality instruction-response pair for training an AI assistant.\n\n")
	}

	// Add topic/category context if provided
	if opts.Topic != "" {
		sb.WriteString(fmt.Sprintf("Topic/Seed: %s\n", opts.Topic))
	}
	if opts.Category != "" {
		sb.WriteString(fmt.Sprintf("Category: %s\n", opts.Category))
	}
	if opts.Language != "" && opts.Language != "en" {
		sb.WriteString(fmt.Sprintf("Language: %s\n", opts.Language))
	}
	if opts.Complexity > 0 {
		sb.WriteString(fmt.Sprintf("Complexity level: %d/5\n", opts.Complexity))
	}

	sb.WriteString("\n")

	// Add few-shot examples if provided
	if opts.NumExamples > 0 && len(opts.Examples) > 0 {
		sb.WriteString("Here are some examples of the expected format:\n\n")
		for i, example := range opts.Examples {
			if i >= opts.NumExamples {
				break
			}
			sb.WriteString(fmt.Sprintf("Example %d:\n", i+1))
			sb.WriteString(fmt.Sprintf("Instruction: %s\n", example.GetString("instruction")))
			if input := example.GetString("input"); input != "" {
				sb.WriteString(fmt.Sprintf("Input: %s\n", input))
			}
			sb.WriteString(fmt.Sprintf("Output: %s\n\n", example.GetString("output")))
		}
	}

	// Output format instruction
	sb.WriteString(`Generate a new instruction-response pair in the following JSON format:
{
  "instruction": "A clear, specific instruction or question",
  "input": "Optional additional context (can be empty string)",
  "output": "A comprehensive, accurate response"
}

Requirements:
- The instruction should be clear and actionable
- The output should be helpful, accurate, and well-formatted
- Vary the style: questions, commands, requests, tasks
- Be creative and diverse in topics and approaches`)

	// Inject user instructions (from context.yaml)
	if opts.UserInstruction != "" {
		sb.WriteString("\n\nAdditional Instructions:\n")
		sb.WriteString(opts.UserInstruction)
	}

	sb.WriteString("\n\nRespond with ONLY the JSON object, no additional text.")

	return sb.String(), nil
}

func (s *InstructionSchema) ParseResponse(raw string) (*Sample, error) {
	raw = StripCodeBlock(raw)

	// Parse JSON
	var data struct {
		Instruction string `json:"instruction"`
		Input       string `json:"input"`
		Output      string `json:"output"`
	}

	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, fmt.Errorf("failed to parse response as JSON: %w", err)
	}

	// Create sample
	sample := &Sample{
		Fields: map[string]any{
			"instruction": data.Instruction,
			"input":       data.Input,
			"output":      data.Output,
		},
	}

	return sample, nil
}

func (s *InstructionSchema) ValidateSample(sample *Sample) error {
	instruction := sample.GetString("instruction")
	if instruction == "" {
		return NewSchemaError(s.Name(), "instruction", "instruction is required")
	}

	output := sample.GetString("output")
	if output == "" {
		return NewSchemaError(s.Name(), "output", "output is required")
	}

	// Quality checks
	if len(instruction) < 10 {
		return NewSchemaError(s.Name(), "instruction", "instruction is too short")
	}
	if len(output) < 10 {
		return NewSchemaError(s.Name(), "output", "output is too short")
	}

	return nil
}

func (s *InstructionSchema) ToJSON(sample *Sample) ([]byte, error) {
	return json.MarshalIndent(sample.Fields, "", "  ")
}

func (s *InstructionSchema) ToJSONL(sample *Sample) ([]byte, error) {
	data, err := json.Marshal(sample.Fields)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
