package schema

import (
	"context"
	"testing"
)

func TestPreferenceSchema_ValidateSample(t *testing.T) {
	s := NewPreferenceSchema()

	tests := []struct {
		name    string
		sample  *Sample
		wantErr bool
	}{
		{
			name: "valid sample",
			sample: &Sample{
				Fields: map[string]any{
					"prompt":   "What is 2+2?",
					"chosen":   "2+2 is 4",
					"rejected": "2+2 is 5",
				},
			},
			wantErr: false,
		},
		{
			name: "missing prompt",
			sample: &Sample{
				Fields: map[string]any{
					"chosen":   "A",
					"rejected": "B",
				},
			},
			wantErr: true,
		},
		{
			name: "identical chosen and rejected",
			sample: &Sample{
				Fields: map[string]any{
					"prompt":   "Question",
					"chosen":   "Answer",
					"rejected": "Answer",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.GeneratePrompt(context.Background(), PromptOptions{})
			if err != nil {
				t.Errorf("GeneratePrompt failed: %v", err)
			}

			err = s.ValidateSample(tt.sample)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSample() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
