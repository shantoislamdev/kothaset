package schema

import (
	"context"
	"strings"
	"testing"
)

func TestClassificationSchema_GeneratePrompt(t *testing.T) {
	s := NewClassificationSchema()
	ctx := context.Background()

	opts := PromptOptions{
		Topic: "Movies",
		Variables: map[string]any{
			"labels": []string{"positive", "negative"},
		},
	}

	prompt, err := s.GeneratePrompt(ctx, opts)
	if err != nil {
		t.Fatalf("GeneratePrompt failed: %v", err)
	}

	if !strings.Contains(prompt, "Category/Domain: Movies") {
		t.Error("Prompt missing topic")
	}
	if !strings.Contains(prompt, "positive, negative") {
		t.Error("Prompt missing labels")
	}
}

func TestClassificationSchema_ParseResponse(t *testing.T) {
	s := NewClassificationSchema()

	validJSON := `{
		"text": "This movie was great!",
		"label": "positive",
		"confidence": 0.95
	}`

	sample, err := s.ParseResponse(validJSON)
	if err != nil {
		t.Fatalf("ParseResponse failed: %v", err)
	}

	if sample.GetString("text") != "This movie was great!" {
		t.Error("Incorrect text parsed")
	}
	if sample.GetString("label") != "positive" {
		t.Error("Incorrect label parsed")
	}

	conf, ok := sample.Fields["confidence"].(float64)
	if !ok || conf != 0.95 {
		t.Error("Incorrect confidence parsed")
	}
}

func TestClassificationSchema_ValidateSample(t *testing.T) {
	s := NewClassificationSchema()

	tests := []struct {
		name    string
		sample  *Sample
		wantErr bool
	}{
		{
			name: "valid sample",
			sample: &Sample{
				Fields: map[string]any{
					"text":  "valid text content",
					"label": "label1",
				},
			},
			wantErr: false,
		},
		{
			name: "missing text",
			sample: &Sample{
				Fields: map[string]any{
					"label": "label1",
				},
			},
			wantErr: true,
		},
		{
			name: "missing label",
			sample: &Sample{
				Fields: map[string]any{
					"text": "valid text content",
				},
			},
			wantErr: true,
		},
		{
			name: "short text",
			sample: &Sample{
				Fields: map[string]any{
					"text":  "hi",
					"label": "label1",
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
