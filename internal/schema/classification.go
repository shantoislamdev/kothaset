package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ClassificationSchema implements text classification format
type ClassificationSchema struct{}

// NewClassificationSchema creates a new classification schema
func NewClassificationSchema() *ClassificationSchema {
	return &ClassificationSchema{}
}

func (s *ClassificationSchema) Name() string        { return "classification" }
func (s *ClassificationSchema) Style() DatasetStyle { return StyleClassification }
func (s *ClassificationSchema) Version() string     { return "1.0" }

func (s *ClassificationSchema) Description() string {
	return "Text classification with labels for training classifiers"
}

func (s *ClassificationSchema) Fields() []FieldDefinition {
	return []FieldDefinition{
		{
			Name:        "text",
			Type:        FieldTypeString,
			Description: "The text to classify",
			Required:    true,
		},
		{
			Name:        "label",
			Type:        FieldTypeString,
			Description: "The classification label",
			Required:    true,
		},
		{
			Name:        "labels",
			Type:        FieldTypeList,
			Description: "Multiple labels for multi-label classification",
			Required:    false,
		},
		{
			Name:        "confidence",
			Type:        FieldTypeFloat,
			Description: "Confidence score (0-1)",
			Required:    false,
		},
	}
}

func (s *ClassificationSchema) RequiredFields() []string {
	return []string{"text", "label"}
}

func (s *ClassificationSchema) GeneratePrompt(ctx context.Context, opts PromptOptions) (string, error) {
	var sb strings.Builder

	sb.WriteString("Generate a text classification example.\n\n")

	if opts.Topic != "" {
		sb.WriteString(fmt.Sprintf("Category/Domain: %s\n", opts.Topic))
	}

	// Get labels from variables if provided
	var labels []string
	if opts.Variables != nil {
		if labelList, ok := opts.Variables["labels"].([]string); ok {
			labels = labelList
		}
	}

	sb.WriteString("\n")

	if len(labels) > 0 {
		sb.WriteString(fmt.Sprintf("Available labels: %s\n\n", strings.Join(labels, ", ")))
		sb.WriteString(`Generate a text sample and assign the most appropriate label:

{
  "text": "The text content to classify",
  "label": "one_of_the_available_labels"
}`)
	} else {
		sb.WriteString(`Generate a text classification example with an appropriate label:

{
  "text": "The text content to classify",
  "label": "an_appropriate_category_label"
}

Common classification types:
- Sentiment: positive, negative, neutral
- Topic: sports, politics, technology, entertainment, etc.
- Intent: question, request, complaint, feedback, etc.
- Toxicity: toxic, non-toxic
- Language: en, es, fr, de, etc.`)
	}

	sb.WriteString("\n\nRespond with ONLY the JSON object, no additional text.")

	return sb.String(), nil
}

func (s *ClassificationSchema) ParseResponse(raw string) (*Sample, error) {
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
		Text       string   `json:"text"`
		Label      string   `json:"label"`
		Labels     []string `json:"labels,omitempty"`
		Confidence float64  `json:"confidence,omitempty"`
	}

	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	fields := map[string]any{
		"text":  data.Text,
		"label": data.Label,
	}
	if len(data.Labels) > 0 {
		fields["labels"] = data.Labels
	}
	if data.Confidence > 0 {
		fields["confidence"] = data.Confidence
	}

	sample := &Sample{
		Fields: fields,
	}

	return sample, nil
}

func (s *ClassificationSchema) ValidateSample(sample *Sample) error {
	text := sample.GetString("text")
	if text == "" {
		return NewSchemaError(s.Name(), "text", "text is required")
	}

	label := sample.GetString("label")
	if label == "" {
		return NewSchemaError(s.Name(), "label", "label is required")
	}

	if len(text) < 5 {
		return NewSchemaError(s.Name(), "text", "text is too short")
	}

	return nil
}

func (s *ClassificationSchema) ToJSON(sample *Sample) ([]byte, error) {
	return json.MarshalIndent(sample.Fields, "", "  ")
}

func (s *ClassificationSchema) ToJSONL(sample *Sample) ([]byte, error) {
	data, err := json.Marshal(sample.Fields)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
