package output

import (
	"testing"

	"github.com/shantoislamdev/kothaset/internal/schema"
)

func TestNewWriter(t *testing.T) {
	sch := schema.NewInstructionSchema()

	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{
			name:    "jsonl format",
			format:  "jsonl",
			wantErr: false,
		},
		{
			name:    "empty format defaults to jsonl",
			format:  "",
			wantErr: false,
		},
		{
			name:    "unsupported format",
			format:  "csv",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := NewWriter(tt.format, sch)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewWriter() error = nil, wantErr %v", tt.wantErr)
				}
				if w != nil {
					t.Errorf("NewWriter() writer = %v, expected nil", w)
				}
			} else {
				if err != nil {
					t.Errorf("NewWriter() unexpected error = %v", err)
				}
				if w == nil {
					t.Error("NewWriter() returned nil writer")
				}

				// Verify the returned writer is of the correct type and format
				if w != nil && w.Format() != "jsonl" {
					t.Errorf("NewWriter() returned format = %v, want jsonl", w.Format())
				}
			}
		})
	}
}
