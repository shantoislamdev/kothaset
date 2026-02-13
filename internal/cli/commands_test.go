package cli

import "testing"

func TestHasExtension_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name string
		path string
		ext  string
		want bool
	}{
		{name: "lowercase", path: "dataset.jsonl", ext: ".jsonl", want: true},
		{name: "uppercase", path: "dataset.JSONL", ext: ".jsonl", want: true},
		{name: "mixedcase", path: "dataset.JsOn", ext: ".json", want: true},
		{name: "non-match", path: "dataset.txt", ext: ".jsonl", want: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := hasExtension(tc.path, tc.ext); got != tc.want {
				t.Fatalf("hasExtension(%q, %q) = %v, want %v", tc.path, tc.ext, got, tc.want)
			}
		})
	}
}

func TestDetectFormat_CaseInsensitive(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{path: "sample.JSONL", want: "jsonl"},
		{path: "sample.JsOn", want: ""},
		{path: "sample.CSV", want: ""},
		{path: "sample.txt", want: ""},
	}

	for _, tc := range tests {
		if got := detectFormat(tc.path); got != tc.want {
			t.Fatalf("detectFormat(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}
