package input

import "testing"

func TestDetectFileType(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"app.js", "js"},
		{"module.mjs", "js"},
		{"component.ts", "ts"},
		{"page.tsx", "ts"},
		{"app.jsx", "jsx"},
		{"index.html", "html"},
		{"page.htm", "html"},
		{"config.json", "json"},
		{"data.xml", "xml"},
		{"icon.svg", "svg"},
		{"styles.css", "css"},
		{"app.js.map", "sourcemap"},
		{"App.vue", "vue"},
		{"Page.svelte", "svelte"},
		{"readme.md", "unknown"},
	}

	for _, tt := range tests {
		got := DetectFileType(tt.filename)
		if got != tt.expected {
			t.Errorf("DetectFileType(%q) = %q, want %q", tt.filename, got, tt.expected)
		}
	}
}

func TestDetectFileType_URLPaths(t *testing.T) {
	// URLs without file extensions should return "unknown"
	tests := []struct {
		filename string
		expected string
	}{
		{"https://example.com/api/endpoint", "unknown"},
		{"https://example.com/", "unknown"},
		{"https://example.com", "unknown"},
		{"https://example.com/page.html", "html"},
		{"https://cdn.example.com/bundle.js", "js"},
		{"https://cdn.example.com/bundle.js?v=1.2.3", "unknown"}, // query string confuses Ext()
	}

	for _, tt := range tests {
		got := DetectFileType(tt.filename)
		if got != tt.expected {
			t.Errorf("DetectFileType(%q) = %q, want %q", tt.filename, got, tt.expected)
		}
	}
}

func TestDetectFileTypeFromContentType(t *testing.T) {
	tests := []struct {
		ct       string
		expected string
	}{
		{"text/html", "html"},
		{"text/javascript", "js"},
		{"application/javascript", "js"},
		{"application/json", "json"},
		{"application/xml", "xml"},
		{"text/xml", "xml"},
		{"image/svg+xml", "svg"},  // Note: contains "xml" but svg check comes first
		{"text/css", "css"},
		{"text/plain", "unknown"},
		{"application/octet-stream", "unknown"},
		{"", "unknown"},
	}

	for _, tt := range tests {
		got := DetectFileTypeFromContentType(tt.ct)
		if got != tt.expected {
			t.Errorf("DetectFileTypeFromContentType(%q) = %q, want %q", tt.ct, got, tt.expected)
		}
	}
}

func TestSniffFileType(t *testing.T) {
	tests := []struct {
		content  string
		expected string
	}{
		{"<!DOCTYPE html><html><body></body></html>", "html"},
		{"<html><head></head></html>", "html"},
		{"  <!doctype HTML>\n<html>", "html"},
		{`{"key": "value"}`, "json"},
		{`[1, 2, 3]`, "json"},
		{`<?xml version="1.0"?>`, "xml"},
		{`<svg xmlns="http://www.w3.org/2000/svg">`, "xml"},
		{`var x = 1;`, "unknown"},
		{`function hello() {}`, "unknown"},
		{"", "unknown"},
	}

	for _, tt := range tests {
		got := SniffFileType([]byte(tt.content))
		if got != tt.expected {
			t.Errorf("SniffFileType(%q...) = %q, want %q", tt.content[:min(len(tt.content), 30)], got, tt.expected)
		}
	}
}

func TestIsBinary(t *testing.T) {
	text := []byte("Hello, this is text content with URLs like /api/test")
	if IsBinary(text) {
		t.Error("Expected text content to not be detected as binary")
	}

	binary := []byte{0x89, 0x50, 0x4E, 0x47, 0x00, 0x00, 0x00}
	if !IsBinary(binary) {
		t.Error("Expected binary content (with null bytes) to be detected as binary")
	}

	// Empty content should not be binary
	if IsBinary([]byte{}) {
		t.Error("Expected empty content to not be binary")
	}

	// UTF-8 text with multibyte chars should not be binary
	utf8 := []byte("Hello 世界 /api/users")
	if IsBinary(utf8) {
		t.Error("Expected UTF-8 text to not be binary")
	}
}
