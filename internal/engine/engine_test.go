package engine

import (
	"testing"

	"github.com/NeCr00/xtract/internal/model"
)

func TestRunEngine_NoInputs(t *testing.T) {
	cfg := &model.Config{Threads: 1}
	results := RunEngine(cfg)
	if len(results) != 0 {
		t.Errorf("Expected 0 results for no inputs, got %d", len(results))
	}
}

func TestRunEngine_SingleFile(t *testing.T) {
	cfg := &model.Config{
		Files:     []string{"../../testdata/layer1_test.js"},
		Threads:   1,
		MaxSizeMB: 100,
	}
	results := RunEngine(cfg)
	if len(results) == 0 {
		t.Error("Expected results from layer1_test.js, got 0")
	}
}

func TestRunEngine_Directory(t *testing.T) {
	cfg := &model.Config{
		Dirs:      []string{"../../testdata"},
		Threads:   2,
		MaxSizeMB: 100,
	}
	results := RunEngine(cfg)
	if len(results) < 100 {
		t.Errorf("Expected 100+ results from testdata directory, got %d", len(results))
	}
}

func TestFilterResults_Scope(t *testing.T) {
	results := []model.Result{
		{URL: "https://target.com/api"},
		{URL: "https://other.com/api"},
		{URL: "/local/path"},
	}
	cfg := &model.Config{Scope: "target.com"}
	filtered := filterResults(results, cfg)
	if len(filtered) != 1 {
		t.Errorf("Scope filter: expected 1, got %d", len(filtered))
	}
	if filtered[0].URL != "https://target.com/api" {
		t.Errorf("Scope filter: expected target.com URL, got %q", filtered[0].URL)
	}
}

func TestFilterResults_Exclude(t *testing.T) {
	results := []model.Result{
		{URL: "/api/users"},
		{URL: "/images/logo.png"},
		{URL: "/api/posts"},
		{URL: "/styles/main.css"},
	}
	cfg := &model.Config{Exclude: `\.(png|css)$`}
	filtered := filterResults(results, cfg)
	if len(filtered) != 2 {
		t.Errorf("Exclude filter: expected 2, got %d", len(filtered))
	}
}

func TestFilterResults_Include(t *testing.T) {
	results := []model.Result{
		{URL: "/api/users"},
		{URL: "/images/logo.png"},
		{URL: "/api/posts"},
		{URL: "/styles/main.css"},
	}
	cfg := &model.Config{Include: `/api/`}
	filtered := filterResults(results, cfg)
	if len(filtered) != 2 {
		t.Errorf("Include filter: expected 2, got %d", len(filtered))
	}
}

func TestFilterResults_InvalidRegex(t *testing.T) {
	results := []model.Result{{URL: "/api/test"}}
	cfg := &model.Config{Exclude: `[invalid`}
	// Should not panic, should return all results (filter skipped)
	filtered := filterResults(results, cfg)
	if len(filtered) != 1 {
		t.Errorf("Invalid regex filter: expected 1 (filter skipped), got %d", len(filtered))
	}
}

func TestSortResults(t *testing.T) {
	results := []model.Result{
		{URL: "/z/path", Category: "page_route"},
		{URL: "/a/path", Category: "page_route"},
		{URL: "/api/test", Category: "api_endpoint"},
	}
	sorted := sortResults(results)
	if sorted[0].Category != "api_endpoint" {
		t.Error("Expected api_endpoint category first")
	}
	if sorted[1].URL != "/a/path" {
		t.Errorf("Expected /a/path second, got %q", sorted[1].URL)
	}
}
