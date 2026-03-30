package extract

import "github.com/NeCr00/xtract/internal/model"

// lineNum returns the line number for offset using the fast binary-search index
// when available, falling back to the O(n) linear scan otherwise.
func lineNum(ctx *model.ExtractionContext, offset int) int {
	if ctx.Lines != nil {
		return ctx.Lines.Line(offset)
	}
	return model.LineNumber(ctx.Content, offset)
}

// ExtractLayer5 runs Layer 5: Subdomain & Infrastructure Discovery (techniques 50-53).
func ExtractLayer5(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	results = append(results, extractSubdomains(ctx)...)
	results = append(results, extractCDNAssetDomains(ctx)...)
	results = append(results, extractInternalHostnames(ctx)...)
	results = append(results, extractWebSocketEndpoints(ctx)...)
	return results
}

// technique 50: subdomain extraction
//
// Instead of running a bare domain regex against the full content (which matches
// nearly every dot-separated identifier in JS code), we anchor to URL/string
// context first:
//   1. Extract hostnames from URLs that start with a protocol (http://, https://, //)
//   2. Extract hostnames from quoted strings that look like domains
//
// Then we validate that the extracted hostname has at least two dot-separated
// labels (i.e. is a subdomain of something).
func extractSubdomains(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	seen := make(map[string]bool)

	// Pattern 1: domains from protocol-prefixed URLs  (http://x.y.z/... or //x.y.z/...)
	// Captures the hostname portion after the protocol.
	reURL := model.GetRegex(`(?i)(?:https?://|//)([a-z0-9](?:[a-z0-9\-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9\-]{0,61}[a-z0-9])?)+\.[a-z]{2,63})`)
	for _, m := range reURL.FindAllStringSubmatchIndex(ctx.Content, -1) {
		if m[2] < 0 {
			continue
		}
		domain := ctx.Content[m[2]:m[3]]
		lower := toLower(domain)
		// Must have at least 2 dots to be a subdomain (a.b.tld)
		if dotCount(lower) < 2 {
			continue
		}
		if seen[lower] {
			continue
		}
		seen[lower] = true
		results = append(results, model.Result{
			URL:             domain,
			SourceFile:      ctx.FileName,
			SourceLine:      lineNum(ctx, m[2]),
			DetectionMethod: "subdomain_extract",
			Category:        model.CategorizeURL(domain),
			Confidence:      model.ConfMedium,
			TechniqueID:     50,
		})
	}

	// Pattern 2: domains in quoted strings or preceded by common delimiters.
	// Uses a broad FQDN regex but anchored to string/URL context to avoid
	// matching every dot-separated JS identifier.
	reQuoted := model.GetRegex(`(?i)(?:["'\x60= ]|^)([a-z0-9](?:[a-z0-9\-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9\-]{0,61}[a-z0-9])?)+\.[a-z]{2,63})(?:[/"'\x60\s:,;)\]}>]|$)`)
	for _, m := range reQuoted.FindAllStringSubmatchIndex(ctx.Content, -1) {
		if m[2] < 0 {
			continue
		}
		domain := ctx.Content[m[2]:m[3]]
		lower := toLower(domain)
		if dotCount(lower) < 1 {
			continue
		}
		// Skip common JS false positives: Object.prototype, module.exports, etc.
		if isJSIdentifier(lower) {
			continue
		}
		if seen[lower] {
			continue
		}
		seen[lower] = true
		results = append(results, model.Result{
			URL:             domain,
			SourceFile:      ctx.FileName,
			SourceLine:      lineNum(ctx, m[2]),
			DetectionMethod: "subdomain_extract",
			Category:        model.CategorizeURL(domain),
			Confidence:      model.ConfMedium,
			TechniqueID:     50,
		})
	}

	return results
}

// isJSIdentifier returns true if the domain-like string is actually a common
// JavaScript identifier pattern (e.g., Object.keys, module.exports, console.log).
func isJSIdentifier(s string) bool {
	jsIdents := []string{
		"object.", "array.", "string.", "number.", "boolean.", "function.",
		"math.", "date.", "regexp.", "json.", "promise.", "symbol.",
		"error.", "map.", "set.", "weakmap.", "weakset.", "proxy.",
		"reflect.", "intl.", "console.", "module.", "exports.", "require.",
		"process.", "window.", "document.", "navigator.", "global.",
		"self.", "this.", "proto.", "prototype.", "constructor.",
		"__webpack", "__esmodule", "e.exports", "t.exports", "n.exports",
		"r.exports", "a.exports",
	}
	for _, id := range jsIdents {
		if len(s) >= len(id) && s[:len(id)] == id {
			return true
		}
	}
	return false
}

// technique 51: CDN/asset domain references
//
// All CDN patterns are combined into a single regex with alternation so the
// engine makes one pass over the content instead of 15 separate passes.
func extractCDNAssetDomains(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result

	// Single combined regex for all CDN providers.
	// Each branch starts with https?:// so the engine can skip non-URL text quickly.
	re := model.GetRegex(`(?i)https?://(?:` +
		// S3 bucket styles
		`[a-z0-9][a-z0-9.\-]*\.s3[.\-][a-z0-9.\-]*\.amazonaws\.com|` +
		`s3[.\-][a-z0-9.\-]*\.amazonaws\.com/|` +
		`s3\.amazonaws\.com/|` +
		// CloudFront
		`[a-z0-9][a-z0-9\-]*\.cloudfront\.net|` +
		// Azure Blob
		`[a-z0-9][a-z0-9\-]*\.blob\.core\.windows\.net|` +
		// Google Cloud Storage
		`storage\.googleapis\.com/|` +
		`firebasestorage\.googleapis\.com/|` +
		// Akamai
		`[a-z0-9][a-z0-9\-]*\.akamaihd\.net|` +
		`[a-z0-9][a-z0-9\-]*\.akamaized\.net|` +
		`[a-z0-9][a-z0-9\-]*\.akamai\.net|` +
		// Fastly
		`[a-z0-9][a-z0-9\-]*\.fastly\.net|` +
		`[a-z0-9][a-z0-9\-]*\.fastlylb\.net|` +
		// Cloudflare
		`[a-z0-9][a-z0-9\-]*\.cdninstagram\.com|` +
		`cdnjs\.cloudflare\.com/|` +
		`[a-z0-9][a-z0-9\-]*\.r2\.cloudflarestorage\.com` +
		`)` +
		// After the domain, consume the rest of the URL
		`[^\s"'<>]*`)

	seen := make(map[string]bool)
	for _, loc := range re.FindAllStringIndex(ctx.Content, -1) {
		url := ctx.Content[loc[0]:loc[1]]
		if seen[url] {
			continue
		}
		seen[url] = true
		results = append(results, model.Result{
			URL:             url,
			SourceFile:      ctx.FileName,
			SourceLine:      lineNum(ctx, loc[0]),
			DetectionMethod: "cdn_asset_domains",
			Category:        model.CatCloudResource,
			Confidence:      model.ConfHigh,
			TechniqueID:     51,
		})
	}
	return results
}

// technique 52: internal hostname patterns
//
// Patterns are consolidated into a few combined regexes instead of 22 separate passes.
func extractInternalHostnames(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	seen := make(map[string]bool)

	addResult := func(url string, offset int) {
		url = stripQuotes(url)
		if seen[url] {
			return
		}
		seen[url] = true
		results = append(results, model.Result{
			URL:             url,
			SourceFile:      ctx.FileName,
			SourceLine:      lineNum(ctx, offset),
			DetectionMethod: "internal_hostnames",
			Category:        model.CatInternalInfra,
			Confidence:      model.ConfHigh,
			TechniqueID:     52,
		})
	}

	// Combined: internal TLD-style domains in URLs (with protocol)
	reURLInternal := model.GetRegex(`(?i)https?://[a-z0-9][a-z0-9.\-]*\.(?:internal|local|staging|dev|test|corp|private)(?:[:/\?#][^\s"'<>]*|)`)
	for _, loc := range reURLInternal.FindAllStringIndex(ctx.Content, -1) {
		addResult(ctx.Content[loc[0]:loc[1]], loc[0])
	}

	// Combined: internal TLD-style domains in quoted strings (without protocol)
	reQuotedInternal := model.GetRegex(`(?i)["'][a-z0-9][a-z0-9.\-]*\.(?:internal|local|staging|corp|private)["']`)
	for _, loc := range reQuotedInternal.FindAllStringIndex(ctx.Content, -1) {
		addResult(ctx.Content[loc[0]:loc[1]], loc[0])
	}

	// Combined: private IPs in URLs
	rePrivateIP := model.GetRegex(`https?://(?:10\.\d{1,3}\.\d{1,3}\.\d{1,3}|172\.(?:1[6-9]|2\d|3[01])\.\d{1,3}\.\d{1,3}|192\.168\.\d{1,3}\.\d{1,3}|127\.0\.0\.1|localhost)(?::\d+)?[^\s"'<>]*`)
	for _, loc := range rePrivateIP.FindAllStringIndex(ctx.Content, -1) {
		addResult(ctx.Content[loc[0]:loc[1]], loc[0])
	}

	// Combined: private IPs in quoted strings (without protocol)
	reQuotedIP := model.GetRegex(`["'](?:10\.\d{1,3}\.\d{1,3}\.\d{1,3}|172\.(?:1[6-9]|2\d|3[01])\.\d{1,3}\.\d{1,3}|192\.168\.\d{1,3}\.\d{1,3}|127\.0\.0\.1|localhost)(?::\d+)?[^\s"'<>]*["']`)
	for _, loc := range reQuotedIP.FindAllStringIndex(ctx.Content, -1) {
		addResult(ctx.Content[loc[0]:loc[1]], loc[0])
	}

	return results
}

// technique 53: WebSocket endpoint discovery
func extractWebSocketEndpoints(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	re := model.GetRegex(`wss?://[^\s"'<>` + "`" + `\)]+`)
	seen := make(map[string]bool)
	matches := re.FindAllStringIndex(ctx.Content, -1)
	for _, loc := range matches {
		url := ctx.Content[loc[0]:loc[1]]
		if seen[url] {
			continue
		}
		seen[url] = true
		results = append(results, model.Result{
			URL:             url,
			SourceFile:      ctx.FileName,
			SourceLine:      lineNum(ctx, loc[0]),
			DetectionMethod: "websocket_endpoints",
			Category:        model.CatWebSocket,
			Confidence:      model.ConfHigh,
			TechniqueID:     53,
		})
	}
	return results
}

// stripQuotes removes surrounding single or double quotes from a string.
func stripQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '\'' && s[len(s)-1] == '\'') || (s[0] == '"' && s[len(s)-1] == '"') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// toLower converts an ASCII string to lowercase without importing strings.
func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

// dotCount counts the number of '.' characters in a string.
func dotCount(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			n++
		}
	}
	return n
}
