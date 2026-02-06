// Package schema implements dataset schema definitions for KothaSet.
package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// DatasetStyle represents the type of dataset
type DatasetStyle string

const (
	StyleInstruction    DatasetStyle = "instruction"    // Instruction-response pairs
	StyleChat           DatasetStyle = "chat"           // Multi-turn conversations
	StylePreference     DatasetStyle = "preference"     // Chosen/rejected for DPO/RLHF
	StyleClassification DatasetStyle = "classification" // Text + label(s)
	StyleCompletion     DatasetStyle = "completion"     // Prompt + completion
)

// Schema defines the interface for dataset schemas
type Schema interface {
	// Metadata
	Name() string
	Style() DatasetStyle
	Description() string
	Version() string

	// Field information
	Fields() []FieldDefinition
	RequiredFields() []string

	// Generation
	GeneratePrompt(ctx context.Context, opts PromptOptions) (string, error)
	ParseResponse(raw string) (*Sample, error)

	// Validation
	ValidateSample(sample *Sample) error

	// Serialization
	ToJSON(sample *Sample) ([]byte, error)
	ToJSONL(sample *Sample) ([]byte, error)
}

// FieldDefinition describes a field in the schema
type FieldDefinition struct {
	Name        string    `json:"name" yaml:"name"`
	Type        FieldType `json:"type" yaml:"type"`
	Description string    `json:"description" yaml:"description"`
	Required    bool      `json:"required" yaml:"required"`
	Default     any       `json:"default,omitempty" yaml:"default,omitempty"`
}

// FieldType represents the data type of a field
type FieldType string

const (
	FieldTypeString  FieldType = "string"
	FieldTypeInt     FieldType = "int"
	FieldTypeFloat   FieldType = "float"
	FieldTypeBool    FieldType = "bool"
	FieldTypeList    FieldType = "list"
	FieldTypeObject  FieldType = "object"
	FieldTypeMessage FieldType = "message" // For chat messages
)

// PromptOptions contains options for prompt generation
type PromptOptions struct {
	// Topic or seed for the sample
	Topic string

	// Category or domain
	Category string

	// Language for the sample
	Language string

	// Complexity level (1-5)
	Complexity int

	// Custom variables for template
	Variables map[string]any

	// Number of examples for few-shot (0 = zero-shot)
	NumExamples int

	// Examples to include
	Examples []*Sample

	// Custom system prompt override
	SystemPrompt string
}

// Sample represents a single dataset sample
type Sample struct {
	// ID is a unique identifier for the sample
	ID string `json:"id"`

	// Fields contains the schema-specific data
	Fields map[string]any `json:"fields"`

	// Metadata about the generation
	Metadata SampleMetadata `json:"metadata,omitempty"`
}

// SampleMetadata contains generation metadata
type SampleMetadata struct {
	// GeneratedAt is when the sample was created
	GeneratedAt time.Time `json:"generated_at"`

	// Provider used for generation
	Provider string `json:"provider"`

	// Model used for generation
	Model string `json:"model"`

	// Temperature used
	Temperature float64 `json:"temperature"`

	// Seed for reproducibility
	Seed int64 `json:"seed,omitempty"`

	// TokensUsed in generation
	TokensUsed int `json:"tokens_used"`

	// Latency of generation
	Latency time.Duration `json:"latency"`

	// Topic/seed used
	Topic string `json:"topic,omitempty"`

	// Custom metadata
	Custom map[string]any `json:"custom,omitempty"`
}

// Get retrieves a field value by name
func (s *Sample) Get(field string) (any, bool) {
	val, ok := s.Fields[field]
	return val, ok
}

// GetString retrieves a string field
func (s *Sample) GetString(field string) string {
	val, ok := s.Fields[field]
	if !ok {
		return ""
	}
	str, _ := val.(string)
	return str
}

// GetStrings retrieves a string slice field
func (s *Sample) GetStrings(field string) []string {
	val, ok := s.Fields[field]
	if !ok {
		return nil
	}
	switch v := val.(type) {
	case []string:
		return v
	case []any:
		strs := make([]string, len(v))
		for i, item := range v {
			strs[i], _ = item.(string)
		}
		return strs
	}
	return nil
}

// Set sets a field value
func (s *Sample) Set(field string, value any) {
	if s.Fields == nil {
		s.Fields = make(map[string]any)
	}
	s.Fields[field] = value
}

// ToJSON converts the sample to JSON
func (s *Sample) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// ToJSONL converts the sample to JSONL format (single line)
func (s *Sample) ToJSONL() ([]byte, error) {
	data, err := json.Marshal(s.Fields)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// SchemaError represents a schema-related error
type SchemaError struct {
	Schema  string
	Field   string
	Message string
}

func (e *SchemaError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("schema %s: field %s: %s", e.Schema, e.Field, e.Message)
	}
	return fmt.Sprintf("schema %s: %s", e.Schema, e.Message)
}

// NewSchemaError creates a new schema error
func NewSchemaError(schema, field, message string) *SchemaError {
	return &SchemaError{
		Schema:  schema,
		Field:   field,
		Message: message,
	}
}
