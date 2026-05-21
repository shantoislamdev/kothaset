package schema

import (
	"context"
	"testing"
)

// mockSchema implements the Schema interface for testing purposes.
type mockSchema struct {
	name string
}

func (m *mockSchema) Name() string { return m.name }
func (m *mockSchema) Style() DatasetStyle { return StyleInstruction }
func (m *mockSchema) Description() string { return "Mock Schema" }
func (m *mockSchema) Version() string { return "1.0" }
func (m *mockSchema) Fields() []FieldDefinition { return nil }
func (m *mockSchema) RequiredFields() []string { return nil }
func (m *mockSchema) GeneratePrompt(ctx context.Context, opts PromptOptions) (string, error) { return "", nil }
func (m *mockSchema) ParseResponse(raw string) (*Sample, error) { return nil, nil }
func (m *mockSchema) ValidateSample(sample *Sample) error { return nil }
func (m *mockSchema) ToJSON(sample *Sample) ([]byte, error) { return nil, nil }
func (m *mockSchema) ToJSONL(sample *Sample) ([]byte, error) { return nil, nil }

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry() returned nil")
	}

	// Should have registered built-ins
	schemas := r.List()
	if len(schemas) == 0 {
		t.Error("NewRegistry() should initialize with built-in schemas")
	}

	// Verify some expected built-in schemas are present
	expectedBuiltins := []string{"instruction", "chat", "preference", "classification"}
	for _, expected := range expectedBuiltins {
		_, err := r.Get(expected)
		if err != nil {
			t.Errorf("Expected built-in schema %q to be registered, but got error: %v", expected, err)
		}
	}
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()

	m1 := &mockSchema{name: "mock1"}

	// Test successful registration
	err := r.Register(m1)
	if err != nil {
		t.Errorf("Failed to register valid schema: %v", err)
	}

	// Test duplicate registration
	err = r.Register(m1)
	if err == nil {
		t.Error("Expected error when registering duplicate schema, got nil")
	}

	// Validate the schema was actually stored
	stored, err := r.Get("mock1")
	if err != nil {
		t.Errorf("Failed to retrieve registered schema: %v", err)
	}
	if stored.Name() != "mock1" {
		t.Errorf("Expected retrieved schema to have name 'mock1', got %q", stored.Name())
	}
}

func TestRegistry_Get(t *testing.T) {
	r := NewRegistry()

	m1 := &mockSchema{name: "mock_get"}
	r.Register(m1)

	// Test existing schema
	s, err := r.Get("mock_get")
	if err != nil {
		t.Errorf("Get failed for existing schema: %v", err)
	}
	if s == nil || s.Name() != "mock_get" {
		t.Errorf("Get returned unexpected schema")
	}

	// Test non-existent schema
	_, err = r.Get("non_existent_schema")
	if err == nil {
		t.Error("Expected error for non-existent schema, got nil")
	}
}

func TestRegistry_List(t *testing.T) {
	// Create an empty registry manually instead of NewRegistry() to avoid built-ins
	r := &Registry{
		schemas: make(map[string]Schema),
	}

	// Initial state
	if len(r.List()) != 0 {
		t.Errorf("Expected empty list, got length %d", len(r.List()))
	}

	// Add some schemas
	r.Register(&mockSchema{name: "a"})
	r.Register(&mockSchema{name: "b"})
	r.Register(&mockSchema{name: "c"})

	names := r.List()
	if len(names) != 3 {
		t.Errorf("Expected length 3, got %d", len(names))
	}

	// Check if all names are present
	nameMap := make(map[string]bool)
	for _, n := range names {
		nameMap[n] = true
	}

	for _, expected := range []string{"a", "b", "c"} {
		if !nameMap[expected] {
			t.Errorf("Expected schema %q in List() result", expected)
		}
	}
}

func TestGlobalRegistry(t *testing.T) {
	mGlobal := &mockSchema{name: "global_mock"}

	// Test Register
	err := Register(mGlobal)
	if err != nil {
		t.Errorf("Global Register failed: %v", err)
	}

	// Duplicate registration should fail
	err = Register(mGlobal)
	if err == nil {
		t.Error("Global Register should fail on duplicate")
	}

	// Test Get
	s, err := Get("global_mock")
	if err != nil {
		t.Errorf("Global Get failed: %v", err)
	}
	if s == nil || s.Name() != "global_mock" {
		t.Errorf("Global Get returned unexpected schema")
	}

	_, err = Get("non_existent_global_schema")
	if err == nil {
		t.Error("Global Get should fail for non-existent schema")
	}

	// Test List
	names := List()
	found := false
	for _, n := range names {
		if n == "global_mock" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Global List did not return the registered schema")
	}
}
