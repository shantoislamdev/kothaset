package schema

import (
	"testing"
	"time"
)

func TestSample_GettersAndSetters(t *testing.T) {
	sample := &Sample{
		Fields: make(map[string]any),
	}

	// Test Set and Get
	sample.Set("test_string", "value")
	if val, ok := sample.Get("test_string"); !ok || val != "value" {
		t.Errorf("Get failed or value mismatch: expected 'value', got %v", val)
	}

	// Test GetString
	if str := sample.GetString("test_string"); str != "value" {
		t.Errorf("GetString failed: expected 'value', got %s", str)
	}

	// Test GetString with non-string value
	sample.Set("test_int", 123)
	if str := sample.GetString("test_int"); str != "" {
		t.Errorf("GetString should return empty string for non-string value, got %s", str)
	}

	// Test GetStrings
	strs := []string{"a", "b", "c"}
	sample.Set("test_strings", strs)
	if res := sample.GetStrings("test_strings"); len(res) != 3 || res[0] != "a" {
		t.Errorf("GetStrings failed for []string input")
	}

	// Test GetStrings with []any
	anyStrs := []any{"x", "y", "z"}
	sample.Set("test_anys", anyStrs)
	if res := sample.GetStrings("test_anys"); len(res) != 3 || res[0] != "x" {
		t.Errorf("GetStrings failed for []any input")
	}
}

func TestSampleMetadata(t *testing.T) {
	meta := SampleMetadata{
		GeneratedAt: time.Now(),
		Provider:    "test-provider",
		Model:       "test-model",
		TokensUsed:  100,
	}

	sample := &Sample{
		ID:       "test-id",
		Metadata: meta,
	}

	if sample.Metadata.Provider != "test-provider" {
		t.Errorf("Metadata not set correctly")
	}
}

func TestStripCodeBlock_Nested(t *testing.T) {
	raw := "```json\n{\n  \"text\": \"example with nested fence ``` inside\",\n  \"ok\": true\n}\n```"

	got := StripCodeBlock(raw)

	if got == raw {
		t.Fatalf("expected code fences to be stripped")
	}
	if got == "" {
		t.Fatalf("expected non-empty stripped content")
	}
	if got[0] != '{' {
		t.Fatalf("expected JSON object start, got: %q", got)
	}
	if got[len(got)-1] != '}' {
		t.Fatalf("expected JSON object end, got: %q", got)
	}
}
