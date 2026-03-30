package engine

import (
	"strings"
	"testing"

	"github.com/NeCr00/xtract/internal/model"
)

func makeResult(url string) model.Result {
	return model.Result{URL: url, Category: "page_route", Confidence: "medium"}
}

// ── Things that SHOULD be rejected ─────────────────────────────

func TestValidate_RejectsTooShort(t *testing.T) {
	rejects := []string{"", "/", "*", ".", "x"}
	for _, u := range rejects {
		r := makeResult(u)
		if isValid(&r) {
			t.Errorf("Should reject too-short URL %q", u)
		}
	}
}

func TestValidate_RejectsMIMETypes(t *testing.T) {
	rejects := []string{
		"application/json",
		"text/html",
		"image/png",
		"font/woff2",
		"video/mp4",
		"audio/mpeg",
		"multipart/form-data",
	}
	for _, u := range rejects {
		r := makeResult(u)
		if isValid(&r) {
			t.Errorf("Should reject MIME type %q", u)
		}
	}
}

func TestValidate_RejectsWhitespace(t *testing.T) {
	rejects := []string{
		"https://example.com/search?q=hello world",
		"/api/test path",
		"/api/\ttab",
	}
	for _, u := range rejects {
		r := makeResult(u)
		if isValid(&r) {
			t.Errorf("Should reject URL with whitespace: %q", u)
		}
	}
}

func TestValidate_RejectsControlChars(t *testing.T) {
	r := makeResult("/api/test\x00hidden")
	if isValid(&r) {
		t.Error("Should reject URL with null byte")
	}
}

func TestValidate_RejectsVersionStrings(t *testing.T) {
	rejects := []string{"1.2.3", "10.0", "0.0.1", "2.0.0"}
	for _, u := range rejects {
		r := makeResult(u)
		if isValid(&r) {
			t.Errorf("Should reject version string %q", u)
		}
	}
}

func TestValidate_RejectsEncodedRemnants(t *testing.T) {
	rejects := []string{
		"2Fapi.example.com",
		"252Fhidden.example.com",
		"25252Ftriple.example.com",
		"3A//example.com",
	}
	for _, u := range rejects {
		r := makeResult(u)
		if isValid(&r) {
			t.Errorf("Should reject encoded remnant %q", u)
		}
	}
}

func TestValidate_RejectsCodeFragments(t *testing.T) {
	rejects := []string{
		"String.fromCharCode.apply",
		"function(a,b)",
		"undefined",
		"null",
		"true",
		"false",
		"Object.keys",
		"Array.isArray",
		"JSON.stringify",
	}
	for _, u := range rejects {
		r := makeResult(u)
		if isValid(&r) {
			t.Errorf("Should reject code fragment %q", u)
		}
	}
}

func TestValidate_RejectsJunkTokens(t *testing.T) {
	rejects := []string{"*", ".", "..", "#", "?", ":", "//", "//:"}
	for _, u := range rejects {
		r := makeResult(u)
		if isValid(&r) {
			t.Errorf("Should reject junk token %q", u)
		}
	}
}

func TestValidate_RejectsTemplateJunk(t *testing.T) {
	r := makeResult("{{DYNAMIC}};// some long comment text that leaked through")
	if isValid(&r) {
		t.Error("Should reject template artifact with comment junk")
	}
}

func TestValidate_RejectsRawEscapes(t *testing.T) {
	rejects := []string{
		`\x2f\x61\x70\x69/mixed/endpoint`,
		`\u002Fpath\u002Ftest`,
	}
	for _, u := range rejects {
		r := makeResult(u)
		if isValid(&r) {
			t.Errorf("Should reject raw escape sequence %q", u)
		}
	}
}

func TestValidate_RejectsTooLong(t *testing.T) {
	longURL := "https://example.com/" + strings.Repeat("a", 2100)
	r := makeResult(longURL)
	if isValid(&r) {
		t.Error("Should reject URL longer than 2048 chars")
	}
}

func TestValidate_RejectsCommonTokens(t *testing.T) {
	rejects := []string{"true/false", "yes/no", "on/off", "n/a", "N/A"}
	for _, u := range rejects {
		r := makeResult(u)
		if isValid(&r) {
			t.Errorf("Should reject common token %q", u)
		}
	}
}

func TestValidate_RejectsNoStructure(t *testing.T) {
	// Strings that don't start with /, ./, http, etc. and don't
	// contain a dot (so not a hostname) are likely noise.
	rejects := []string{
		"randomword",
		"just-a-name",
		"kebab-case",
		"CamelCase",
	}
	for _, u := range rejects {
		r := makeResult(u)
		if isValid(&r) {
			t.Errorf("Should reject structureless string %q", u)
		}
	}
}

// ── Things that MUST be kept ───────────────────────────────────

func TestValidate_KeepsAbsoluteURLs(t *testing.T) {
	keeps := []string{
		"https://api.example.com/v1/users",
		"http://localhost:8080/api",
		"ftp://files.example.com/data",
		"wss://ws.example.com/socket",
		"ws://realtime.example.com/ws",
		"//cdn.example.com/lib.js",
	}
	for _, u := range keeps {
		r := makeResult(u)
		if !isValid(&r) {
			t.Errorf("Should keep absolute URL %q", u)
		}
	}
}

func TestValidate_KeepsRelativePaths(t *testing.T) {
	keeps := []string{
		"/api/v1/users",
		"/graphql",
		"/dashboard/settings",
		"./components/Header",
		"../utils/helpers",
		"/api/v1/users?page=1&limit=20",
		"/api/v1/users/{id}",
	}
	for _, u := range keeps {
		r := makeResult(u)
		if !isValid(&r) {
			t.Errorf("Should keep relative path %q", u)
		}
	}
}

func TestValidate_KeepsBareHostnames(t *testing.T) {
	keeps := []string{
		"api.example.com",
		"cdn.example.com/assets/bundle.js",
		"staging.api.example.com",
		"s3.amazonaws.com/bucket/key",
	}
	for _, u := range keeps {
		r := makeResult(u)
		if !isValid(&r) {
			t.Errorf("Should keep bare hostname %q", u)
		}
	}
}

func TestValidate_KeepsSpecialProtocols(t *testing.T) {
	keeps := []string{
		"mailto:admin@example.com",
		"data:image/png;base64,abc",
		"android-app://com.example.app/deep/link",
		"intent://scan/#Intent;scheme=zxing;end",
		"javascript:void(0)",
	}
	for _, u := range keeps {
		r := makeResult(u)
		if !isValid(&r) {
			t.Errorf("Should keep special protocol URL %q", u)
		}
	}
}

func TestValidate_KeepsAPIEndpoints(t *testing.T) {
	keeps := []string{
		"/api/v1/users",
		"/api/v2/admin/settings",
		"/rest/services/data",
		"/rpc/execute",
		"/ws/notifications",
	}
	for _, u := range keeps {
		r := makeResult(u)
		if !isValid(&r) {
			t.Errorf("Should keep API endpoint %q", u)
		}
	}
}

func TestValidate_KeepsHashRoutes(t *testing.T) {
	keeps := []string{
		"#/dashboard",
		"#/users/profile/edit",
	}
	for _, u := range keeps {
		r := makeResult(u)
		if !isValid(&r) {
			t.Errorf("Should keep hash route %q", u)
		}
	}
}

func TestValidate_KeepsTemplateLiterals(t *testing.T) {
	// Clean template URLs with {{DYNAMIC}} are valid — they represent
	// parameterized endpoints.
	keeps := []string{
		"/api/v1/users/{{DYNAMIC}}/profile",
		"/api/v2/items/{{DYNAMIC}}",
	}
	for _, u := range keeps {
		r := makeResult(u)
		if !isValid(&r) {
			t.Errorf("Should keep template literal URL %q", u)
		}
	}
}

func TestValidate_KeepsEnvVars(t *testing.T) {
	keeps := []string{
		"process.env.API_URL",
		"process.env.REACT_APP_API_BASE",
		"import.meta.env.VITE_BACKEND_URL",
	}
	for _, u := range keeps {
		r := makeResult(u)
		if !isValid(&r) {
			t.Errorf("Should keep env variable reference %q", u)
		}
	}
}

func TestValidate_KeepsGraphQLOps(t *testing.T) {
	r := makeResult("graphql:GetUsers")
	if !isValid(&r) {
		t.Error("Should keep GraphQL operation name")
	}
}

func TestValidate_KeepsLongDataURIs(t *testing.T) {
	longData := "data:image/png;base64," + strings.Repeat("A", 3000)
	r := makeResult(longData)
	if !isValid(&r) {
		t.Error("Should keep long data URIs (exempt from length limit)")
	}
}

func TestValidate_KeepsSourceMaps(t *testing.T) {
	keeps := []string{
		"app.bundle.js.map",
		"vendor.min.js.map",
	}
	for _, u := range keeps {
		r := makeResult(u)
		if !isValid(&r) {
			t.Errorf("Should keep source map reference %q", u)
		}
	}
}

// ── Integration: validateResults batch ─────────────────────────

func TestValidateResults_BatchFiltering(t *testing.T) {
	input := []model.Result{
		makeResult("/api/v1/users"),             // keep
		makeResult("*"),                         // reject
		makeResult("/"),                         // reject
		makeResult("application/json"),          // reject
		makeResult("https://example.com/valid"), // keep
		makeResult("1.2.3"),                     // reject
		makeResult("true/false"),                // reject
		makeResult("api.example.com/path"),      // keep
	}

	output := validateResults(input)

	if len(output) != 3 {
		urls := make([]string, len(output))
		for i, r := range output {
			urls[i] = r.URL
		}
		t.Errorf("Expected 3 valid results, got %d: %v", len(output), urls)
	}
}
