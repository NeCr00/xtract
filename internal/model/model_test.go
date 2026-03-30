package model

import (
	"testing"
)

// ── CategorizeURL Tests ────────────────────────────────────────

func TestCategorizeURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"wss://ws.example.com/socket", CatWebSocket},
		{"ws://ws.example.com/socket", CatWebSocket},
		{"/api/v1/users", CatAPIEndpoint},
		{"/graphql", CatAPIEndpoint},
		{"https://bucket.s3.amazonaws.com/file", CatCloudResource},
		{"http://localhost:3000/dev", CatInternalInfra},
		{"http://192.168.1.1/admin", CatInternalInfra},
		{"/images/logo.png", CatStaticAsset},
		{"app.bundle.js.map", CatSourceMap},
		{"data:image/png;base64,abc", CatDataURI},
		{"mailto:admin@example.com", CatMailto},
		{"https://external.example.com/page", CatExternalSvc},
		{"/dashboard/settings", CatPageRoute},
	}

	for _, tt := range tests {
		got := CategorizeURL(tt.url)
		if got != tt.expected {
			t.Errorf("CategorizeURL(%q) = %q, want %q", tt.url, got, tt.expected)
		}
	}
}

func TestCategorizeURL_PrivateIPRanges(t *testing.T) {
	privateIPs := []string{
		"http://10.0.0.1/api",
		"http://10.255.255.255/api",
		"http://172.16.0.1/api",
		"http://172.20.0.1/api",
		"http://172.29.255.255/api",
		"http://172.30.0.1/api",
		"http://172.31.255.255/api",
		"http://192.168.0.1/api",
		"http://127.0.0.1:8080/api",
	}
	for _, url := range privateIPs {
		got := CategorizeURL(url)
		if got != CatInternalInfra {
			t.Errorf("CategorizeURL(%q) = %q, want %q (should be private)", url, got, CatInternalInfra)
		}
	}

	// Public IPs that previously matched the overly broad "172.2" pattern
	publicIPs := []string{
		"http://172.2.0.1/api",
		"http://172.200.0.1/api",
		"http://172.1.0.1/api",
		"http://172.32.0.1/api",
	}
	for _, url := range publicIPs {
		got := CategorizeURL(url)
		if got == CatInternalInfra {
			t.Errorf("CategorizeURL(%q) = %q, but this is a PUBLIC IP", url, got)
		}
	}
}

func TestCategorizeURL_WebSocketBothProtocols(t *testing.T) {
	for _, url := range []string{
		"ws://example.com/socket",
		"wss://example.com/socket",
		"ws://localhost:8080/ws",
		"wss://api.example.com:443/realtime",
	} {
		got := CategorizeURL(url)
		if got != CatWebSocket {
			t.Errorf("CategorizeURL(%q) = %q, want %q", url, got, CatWebSocket)
		}
	}
}

func TestCategorizeURL_EdgeCases(t *testing.T) {
	if got := CategorizeURL(""); got != CatPageRoute {
		t.Errorf("CategorizeURL(\"\") = %q, want page_route", got)
	}
	if got := CategorizeURL("/"); got != CatPageRoute {
		t.Errorf("CategorizeURL(\"/\") = %q, want page_route", got)
	}
}

// ── ResultSet Tests ────────────────────────────────────────────

func TestDeduplication(t *testing.T) {
	rs := NewResultSet()
	rs.Add(Result{URL: "/api/test"})
	rs.Add(Result{URL: "/api/test"})
	rs.Add(Result{URL: "/api/other"})

	if rs.Count() != 2 {
		t.Errorf("Expected 2, got %d", rs.Count())
	}
}

func TestDeduplication_MethodDistinction(t *testing.T) {
	rs := NewResultSet()
	rs.Add(Result{URL: "/api/users", HTTPMethod: "GET"})
	rs.Add(Result{URL: "/api/users", HTTPMethod: "POST"})
	rs.Add(Result{URL: "/api/users", HTTPMethod: "GET"})

	if rs.Count() != 2 {
		t.Errorf("Expected 2 (GET and POST), got %d", rs.Count())
	}
}

func TestDeduplication_ConcurrentSafety(t *testing.T) {
	rs := NewResultSet()
	done := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				rs.Add(Result{URL: "/api/test"})
			}
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	if rs.Count() != 1 {
		t.Errorf("Expected 1 after concurrent adds, got %d", rs.Count())
	}
}

// ── Technique Registry ─────────────────────────────────────────

func TestAllTechniquesCount(t *testing.T) {
	techniques := AllTechniques()
	if len(techniques) != 67 {
		t.Errorf("Expected 67, got %d", len(techniques))
	}
	for i, tech := range techniques {
		if tech.ID != i+1 {
			t.Errorf("Technique index %d has ID %d, expected %d", i, tech.ID, i+1)
		}
	}
}

// ── LineNumber Tests ───────────────────────────────────────────

func TestLineNumber(t *testing.T) {
	content := "line1\nline2\nline3\nline4"
	tests := []struct {
		offset   int
		expected int
	}{
		{0, 1},
		{4, 1},
		{5, 1},
		{6, 2},
		{12, 3},
		{18, 4},
	}
	for _, tt := range tests {
		got := LineNumber(content, tt.offset)
		if got != tt.expected {
			t.Errorf("LineNumber(content, %d) = %d, want %d", tt.offset, got, tt.expected)
		}
	}
}

func TestLineNumber_EdgeCases(t *testing.T) {
	if got := LineNumber("", 0); got != 1 {
		t.Errorf("LineNumber(\"\", 0) = %d, want 1", got)
	}
	if got := LineNumber("no newlines", 5); got != 1 {
		t.Errorf("LineNumber(\"no newlines\", 5) = %d, want 1", got)
	}
	if got := LineNumber("test", -1); got != 1 {
		t.Errorf("LineNumber(\"test\", -1) = %d, want 1", got)
	}
}

// ── GetRegex Tests ─────────────────────────────────────────────

func TestGetRegex_Caching(t *testing.T) {
	r1 := GetRegex(`\d+`)
	r2 := GetRegex(`\d+`)
	if r1 != r2 {
		t.Error("GetRegex should return same instance for identical patterns")
	}
}

func TestGetRegex_InvalidPattern(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("GetRegex with invalid pattern should panic")
		}
	}()
	GetRegex(`[invalid`)
}

// ── ContainsStr Tests ──────────────────────────────────────────

func TestContainsStr(t *testing.T) {
	if !ContainsStr("hello world", "world") {
		t.Error("Expected true")
	}
	if ContainsStr("hello", "world") {
		t.Error("Expected false")
	}
	if ContainsStr("", "x") {
		t.Error("Expected false for empty string")
	}
	if !ContainsStr("x", "x") {
		t.Error("Expected true for exact match")
	}
}
