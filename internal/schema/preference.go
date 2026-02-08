package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// PreferenceSchema implements the DPO/RLHF preference format
type PreferenceSchema struct{}

// NewPreferenceSchema creates a new preference schema
func NewPreferenceSchema() *PreferenceSchema {
	return &PreferenceSchema{}
}

func (s *PreferenceSchema) Name() string        { return "preference" }
func (s *PreferenceSchema) Style() DatasetStyle { return StylePreference }
func (s *PreferenceSchema) Version() string     { return "1.0" }

func (s *PreferenceSchema) Description() string {
	return "DPO/RLHF preference pairs with chosen and rejected responses"
}

func (s *PreferenceSchema) Fields() []FieldDefinition {
	return []FieldDefinition{
		{
			Name:        "prompt",
			Type:        FieldTypeString,
			Description: "The instruction or question",
			Required:    true,
		},
		{
			Name:        "chosen",
			Type:        FieldTypeString,
			Description: "The preferred/better response",
			Required:    true,
		},
		{
			Name:        "rejected",
			Type:        FieldTypeString,
			Description: "The less preferred/worse response",
			Required:    true,
		},
	}
}

func (s *PreferenceSchema) RequiredFields() []string {
	return []string{"prompt", "chosen", "rejected"}
}

func (s *PreferenceSchema) GeneratePrompt(ctx context.Context, opts PromptOptions) (string, error) {
	var sb strings.Builder

	// Inject user context first (from context.yaml)
	if opts.UserContext != "" {
		sb.WriteString(opts.UserContext)
		sb.WriteString("\n\n")
	} else {
		// Default context if none provided
		sb.WriteString("Generate a preference pair for training AI alignment.\n\n")
	}

	if opts.Topic != "" {
		sb.WriteString(fmt.Sprintf("Topic: %s\n", opts.Topic))
	}
	if opts.Category != "" {
		sb.WriteString(fmt.Sprintf("Category: %s\n", opts.Category))
	}

	sb.WriteString("\n")

	sb.WriteString(`Generate a prompt with two responses - one better (chosen) and one worse (rejected):

{
  "prompt": "A clear question or instruction",
  "chosen": "The preferred response - helpful, accurate, safe, and well-written",
  "rejected": "A less preferred response - could be less helpful, less accurate, less safe, or lower quality"
}

The difference between chosen and rejected should represent clear quality distinctions:
- Accuracy: chosen is factually correct, rejected has minor errors
- Helpfulness: chosen directly addresses the need, rejected is vague
- Safety: chosen avoids harmful content, rejected may be borderline
- Clarity: chosen is well-organized, rejected is confusing
- Completeness: chosen is thorough, rejected is incomplete`)

	// Inject user instructions (from context.yaml)
	if opts.UserInstruction != "" {
		sb.WriteString("\n\nAdditional Instructions:\n")
		sb.WriteString(opts.UserInstruction)
	}

	sb.WriteString("\n\nRespond with ONLY the JSON object, no additional text.")

	return sb.String(), nil
}

func (s *PreferenceSchema) ParseResponse(raw string) (*Sample, error) {
	raw = strings.TrimSpace(raw)

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
		Prompt   string `json:"prompt"`
		Chosen   string `json:"chosen"`
		Rejected string `json:"rejected"`
	}

	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	sample := &Sample{
		Fields: map[string]any{
			"prompt":   data.Prompt,
			"chosen":   data.Chosen,
			"rejected": data.Rejected,
		},
	}

	return sample, nil
}

func (s *PreferenceSchema) ValidateSample(sample *Sample) error {
	prompt := sample.GetString("prompt")
	if prompt == "" {
		return NewSchemaError(s.Name(), "prompt", "prompt is required")
	}

	chosen := sample.GetString("chosen")
	if chosen == "" {
		return NewSchemaError(s.Name(), "chosen", "chosen is required")
	}

	rejected := sample.GetString("rejected")
	if rejected == "" {
		return NewSchemaError(s.Name(), "rejected", "rejected is required")
	}

	// Quality checks
	if len(prompt) < 10 {
		return NewSchemaError(s.Name(), "prompt", "prompt is too short")
	}
	if chosen == rejected {
		return NewSchemaError(s.Name(), "chosen", "chosen and rejected should be different")
	}

	return nil
}

func (s *PreferenceSchema) ToJSON(sample *Sample) ([]byte, error) {
	return json.MarshalIndent(sample.Fields, "", "  ")
}

func (s *PreferenceSchema) ToJSONL(sample *Sample) ([]byte, error) {
	data, err := json.Marshal(sample.Fields)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
