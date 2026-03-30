package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/NeCr00/xtract/internal/model"
)

func testResults() []model.Result {
	return []model.Result{
		{
			URL:             "/api/v1/users",
			SourceFile:      "app.js",
			SourceLine:      42,
			DetectionMethod: "fetch_calls",
			HTTPMethod:      "GET",
			QueryParams:     []string{"page", "limit"},
			Category:        "api_endpoint",
			Confidence:      "high",
			TechniqueID:     15,
		},
		{
			URL:             "https://cdn.example.com/bundle.js",
			SourceFile:      "index.html",
			SourceLine:      10,
			DetectionMethod: "html_script_src",
			Category:        "static_asset",
			Confidence:      "high",
			TechniqueID:     0,
		},
	}
}

func TestWriteText_URLsOnly(t *testing.T) {
	var buf bytes.Buffer
	results := testResults()
	cfg := &model.Config{}
	writeText(&buf, results, cfg)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(lines))
	}
	if lines[0] != "/api/v1/users" {
		t.Errorf("Expected '/api/v1/users', got %q", lines[0])
	}
}

func TestWriteText_WithMethods(t *testing.T) {
	var buf bytes.Buffer
	results := testResults()
	cfg := &model.Config{WithMethods: true}
	writeText(&buf, results, cfg)

	output := buf.String()
	if !strings.Contains(output, "GET /api/v1/users") {
		t.Errorf("Expected 'GET /api/v1/users' in output, got %q", output)
	}
}

func TestWriteText_WithSource(t *testing.T) {
	var buf bytes.Buffer
	results := testResults()
	cfg := &model.Config{WithSource: true}
	writeText(&buf, results, cfg)

	output := buf.String()
	if !strings.Contains(output, "(app.js:42)") {
		t.Errorf("Expected '(app.js:42)' in output, got %q", output)
	}
}

func TestWriteText_WithParams(t *testing.T) {
	var buf bytes.Buffer
	results := testResults()
	cfg := &model.Config{WithParams: true}
	writeText(&buf, results, cfg)

	output := buf.String()
	if !strings.Contains(output, "[page,limit]") {
		t.Errorf("Expected '[page,limit]' in output, got %q", output)
	}
}

func TestWriteText_Debug(t *testing.T) {
	var buf bytes.Buffer
	results := testResults()
	cfg := &model.Config{Debug: true}
	writeText(&buf, results, cfg)

	output := buf.String()
	if !strings.Contains(output, "{fetch_calls #15}") {
		t.Errorf("Expected '{fetch_calls #15}' in output, got %q", output)
	}
}

func TestWriteText_NewlinesSanitized(t *testing.T) {
	var buf bytes.Buffer
	results := []model.Result{
		{URL: "/api/test\ninjected\nlines"},
	}
	cfg := &model.Config{}
	writeText(&buf, results, cfg)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Errorf("Newlines in URL should be sanitized, got %d lines", len(lines))
	}
}

func TestWriteJSONLines(t *testing.T) {
	var buf bytes.Buffer
	results := testResults()
	cfg := &model.Config{}
	writeJSONLines(&buf, results, cfg)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 JSON lines, got %d", len(lines))
	}

	var r model.Result
	if err := json.Unmarshal([]byte(lines[0]), &r); err != nil {
		t.Errorf("Failed to parse JSON line: %v", err)
	}
	if r.URL != "/api/v1/users" {
		t.Errorf("Expected '/api/v1/users', got %q", r.URL)
	}
}

func TestWriteCSV(t *testing.T) {
	var buf bytes.Buffer
	results := testResults()
	writeCSV(&buf, results)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 { // header + 2 data rows
		t.Errorf("Expected 3 CSV lines (header + 2 rows), got %d", len(lines))
	}
	if !strings.HasPrefix(lines[0], "url,") {
		t.Errorf("Expected CSV header starting with 'url,', got %q", lines[0])
	}
}
