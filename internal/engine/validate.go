package engine

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/NeCr00/xtract/internal/model"
)

// validateResults removes false positives from extracted results.
// Each rule is conservative — it only rejects URLs that are clearly
// invalid or noise. When in doubt, keep the result.
func validateResults(results []model.Result) []model.Result {
	filtered := make([]model.Result, 0, len(results))
	for i := range results {
		if isValid(&results[i]) {
			filtered = append(filtered, results[i])
		}
	}
	return filtered
}

// isValid returns true if a result looks like a genuine URL/endpoint.
func isValid(r *model.Result) bool {
	u := r.URL

	// ── Rule 1: Length bounds ────────────────────────────────
	// A valid URL/path must be at least 2 chars (e.g. "/x").
	// Reject anything over 2048 chars (RFC 2616 practical limit,
	// also browser max). Data URIs are exempt from the upper limit.
	if len(u) < 2 {
		return false
	}
	if len(u) > 2048 && !strings.HasPrefix(u, "data:") {
		return false
	}

	// ── Rule 2: No whitespace ───────────────────────────────
	// URLs cannot contain unencoded spaces, tabs, or newlines.
	// Exception: data URIs may contain encoded whitespace.
	if !strings.HasPrefix(u, "data:") {
		for _, c := range u {
			if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
				return false
			}
		}
	}

	// ── Rule 3: No control characters ───────────────────────
	// Reject URLs containing ASCII control chars (0x00-0x1F, 0x7F).
	for _, c := range u {
		if c < 0x20 || c == 0x7F {
			return false
		}
	}

	// ── Rule 4: Not a bare wildcard or garbage token ────────
	if u == "*" || u == "." || u == ".." || u == "#" || u == "?" || u == ":" || u == "//" || u == "//:" {
		return false
	}

	// ── Rule 5: MIME type false positives ────────────────────
	// Strings like "application/json", "text/html", "image/png"
	// are not URLs but frequently extracted from Content-Type headers.
	if isMIMEType(u) {
		return false
	}

	// ── Rule 6: Pure numeric/version strings ────────────────
	// "1.2.3", "2.0", "10.0.0" — version numbers, not paths.
	// Only reject if there's no path separator after a domain-like part.
	if isPureVersionString(u) {
		return false
	}

	// ── Rule 7: Encoded remnants that weren't decoded ───────
	// Strings like "2Fapi.example.com", "252Ftriple.example.com"
	// are partially-decoded URL fragments. The leading hex digits
	// are remnants of %2F or %252F encoding.
	if isEncodedRemnant(u) {
		return false
	}

	// ── Rule 8: Code fragments ──────────────────────────────
	// Reject strings that look like code, not URLs.
	// E.g. "String.fromCharCode.apply", "function()", "var x".
	if isCodeFragment(u) {
		return false
	}

	// ── Rule 9: Template artifacts ──────────────────────────
	// Reject URLs that contain unresolved template junk that
	// leaked past the template literal handler.
	// E.g. "{{DYNAMIC}};// comment text..."
	if strings.Contains(u, "{{DYNAMIC}}") && (strings.Contains(u, "//") || strings.Contains(u, ";") || len(u) > 100) {
		return false
	}

	// ── Rule 10: Bare escaped sequences ─────────────────────
	// If the URL still contains raw \x or \u escape sequences,
	// it wasn't properly decoded and is likely noise.
	if strings.Contains(u, "\\x") || strings.Contains(u, "\\u") {
		return false
	}

	// ── Rule 11: Relative path validation ───────────────────
	// Paths starting with / must have at least one valid path char
	// after the slash. Reject bare "/" which adds no value.
	if u == "/" {
		return false
	}

	// ── Rule 12: Reject common JS/CSS tokens misidentified as paths ─
	if isCommonToken(u) {
		return false
	}

	// ── Rule 13: Path-like strings need path structure ──────
	// If a string has no protocol and doesn't start with / or ./,
	// it might be a bare hostname (valid) or garbage. Bare hostnames
	// are valid for security testing, so we keep those. But reject
	// strings that don't look like hostnames or paths.
	if !hasValidStructure(u) {
		return false
	}

	return true
}

// isMIMEType returns true if the string looks like a MIME type.
var mimeTypePrefixes = []string{
	"application/", "text/", "image/", "audio/", "video/",
	"font/", "multipart/", "message/", "model/", "chemical/",
}

func isMIMEType(s string) bool {
	lower := strings.ToLower(s)
	for _, prefix := range mimeTypePrefixes {
		if strings.HasPrefix(lower, prefix) {
			// Make sure it's a simple MIME type, not a URL that
			// happens to start with "text/" etc.
			if !strings.Contains(s, "//") && !strings.Contains(s, "?") {
				return true
			}
		}
	}
	return false
}

// isPureVersionString returns true for strings that are only digits and dots.
// E.g. "1.2.3", "10.0.0.1" (but NOT "10.0.0.1/api" which has a path).
func isPureVersionString(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c != '.' && (c < '0' || c > '9') {
			return false
		}
	}
	return true
}

// isEncodedRemnant detects partially-decoded URL fragments where leading
// hex chars are leftover from URL encoding (e.g. "2Fapi" from "%2Fapi").
func isEncodedRemnant(s string) bool {
	if len(s) < 3 {
		return false
	}
	// Pattern: 2-6 hex digits followed by a letter/domain.
	// E.g. "2F" from %2F, "252F" from %252F, "25252F" from %25252F
	hexPrefixes := []string{"2F", "252F", "25252F", "3A", "253A"}
	for _, prefix := range hexPrefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

// isCodeFragment detects strings that look like code rather than URLs.
func isCodeFragment(s string) bool {
	codePatterns := []string{
		"String.fromCharCode",
		"function(",
		"function ",
		"var ",
		"let ",
		"const ",
		"return ",
		"typeof ",
		"undefined",
		"null",
		"true",
		"false",
		"NaN",
		"Infinity",
		"Object.",
		"Array.",
		"Math.",
		"Date.",
		"RegExp(",
		"JSON.",
		"Promise.",
		"Symbol(",
		"Error(",
	}
	for _, pat := range codePatterns {
		if strings.HasPrefix(s, pat) {
			return true
		}
	}

	// Reject if it contains assignment operators or semicolons
	// (strong indicator of code, not a URL).
	if (strings.Contains(s, " = ") || strings.Contains(s, " => ")) && !strings.Contains(s, "?") {
		return false // let the URL-structure check handle it
	}

	return false
}

// isCommonToken checks for common JavaScript/CSS tokens that get
// misidentified as paths due to containing a forward slash.
var commonTokens = map[string]bool{
	"true/false": true,
	"yes/no":     true,
	"on/off":     true,
	"0/0":        true,
	"n/a":        true,
	"N/A":        true,
	"/v1":        false, // valid API version path, keep it
}

func isCommonToken(s string) bool {
	return commonTokens[s]
}

// hasValidStructure checks that a URL has valid structural characteristics.
func hasValidStructure(s string) bool {
	// Absolute URLs with known protocols are structurally valid.
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") ||
		strings.HasPrefix(s, "ftp://") || strings.HasPrefix(s, "ws://") ||
		strings.HasPrefix(s, "wss://") || strings.HasPrefix(s, "//") ||
		strings.HasPrefix(s, "data:") || strings.HasPrefix(s, "blob:") ||
		strings.HasPrefix(s, "javascript:") || strings.HasPrefix(s, "mailto:") ||
		strings.HasPrefix(s, "android-app://") || strings.HasPrefix(s, "ios-app://") ||
		strings.HasPrefix(s, "intent://") || strings.HasPrefix(s, "deeplink://") {
		return true
	}

	// Paths starting with / or ./ or ../ are structurally valid.
	if strings.HasPrefix(s, "/") || strings.HasPrefix(s, "./") || strings.HasPrefix(s, "../") {
		return true
	}

	// Hash routes (#/path) are valid.
	if strings.HasPrefix(s, "#/") {
		return true
	}

	// GraphQL operation names (graphql:OperationName) are valid.
	if strings.HasPrefix(s, "graphql:") {
		return true
	}

	// Bare hostnames: must have at least one dot and look like a domain.
	// E.g. "api.example.com", "cdn.example.com/path"
	if strings.Contains(s, ".") {
		// Must start with alphanumeric (not a dot or special char).
		first, _ := utf8.DecodeRuneInString(s)
		if unicode.IsLetter(first) || unicode.IsDigit(first) {
			return true
		}
	}

	// Env variable references (process.env.X) are valid metadata.
	if strings.HasPrefix(s, "process.env.") || strings.HasPrefix(s, "import.meta.env.") {
		return true
	}

	// If none of the above, it's likely a false positive.
	return false
}
