package extract

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"

	"github.com/NeCr00/xtract/internal/model"
)

// ExtractLayer1 runs all 14 regex-based extraction techniques and returns combined results.
func ExtractLayer1(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	results = append(results, extractAbsoluteURLs(ctx)...)
	results = append(results, extractRelativePaths(ctx)...)
	results = append(results, extractAPIPatterns(ctx)...)
	results = append(results, extractQueryStrings(ctx)...)
	results = append(results, extractHashFragments(ctx)...)
	results = append(results, extractTemplateLiterals(ctx)...)
	results = append(results, extractStringConcatenation(ctx)...)
	results = append(results, extractDataURIs(ctx)...)
	results = append(results, extractBlobJavascriptURIs(ctx)...)
	results = append(results, extractMailtoLinks(ctx)...)
	results = append(results, extractIPBasedURLs(ctx)...)
	results = append(results, extractEncodedURLs(ctx)...)
	results = append(results, extractCustomProtocols(ctx)...)
	results = append(results, extractStringLiterals(ctx)...)
	return results
}

// technique1: Absolute URLs - http(s), ftp, wss, protocol-relative
func extractAbsoluteURLs(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result

	// Match absolute URLs: http://, https://, ftp://, ws://, wss://
	re := model.GetRegex(`(?i)((?:https?|ftp|wss?)://[^\s"'\x60<>{}\[\]|\\^` + "`" + `\x00-\x1f]{3,})`)

	for _, match := range re.FindAllStringIndex(ctx.Content, -1) {
		raw := ctx.Content[match[0]:match[1]]
		url := cleanTrailingURL(raw)
		if url == "" {
			continue
		}
		results = append(results, model.Result{
			URL:             url,
			SourceFile:      ctx.FileName,
			SourceLine:      model.LineNumber(ctx.Content, match[0]),
			DetectionMethod: "absolute_urls",
			Category:        model.CategorizeURL(url),
			Confidence:      model.ConfMedium,
			TechniqueID:     1,
		})
	}

	// Protocol-relative URLs: //example.com/path
	reProto := model.GetRegex(`(?:^|[=("'\x60\s])(//((?:[a-zA-Z0-9](?:[a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,})[^\s"'\x60<>)]*)`)
	for _, match := range reProto.FindAllStringSubmatchIndex(ctx.Content, -1) {
		if match[2] < 0 {
			continue
		}
		url := cleanTrailingURL(ctx.Content[match[2]:match[3]])
		if url == "" || len(url) < 5 {
			continue
		}
		results = append(results, model.Result{
			URL:             url,
			SourceFile:      ctx.FileName,
			SourceLine:      model.LineNumber(ctx.Content, match[2]),
			DetectionMethod: "absolute_urls",
			Category:        model.CategorizeURL(url),
			Confidence:      model.ConfMedium,
			TechniqueID:     1,
		})
	}

	return results
}

// technique2: Relative paths inside quotes
func extractRelativePaths(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result

	// Relative paths starting with /, ./, or ../ inside quotes
	re := model.GetRegex(`(?:"|'|` + "`" + `)((?:\.{1,2})?/(?:[a-zA-Z0-9_\-~.%@:]+/)*[a-zA-Z0-9_\-~.%@:]*(?:\?[^"'\x60\s]*)?(?:#[^"'\x60\s]*)?)(?:"|'|` + "`" + `)`)

	for _, match := range re.FindAllStringSubmatchIndex(ctx.Content, -1) {
		if match[2] < 0 {
			continue
		}
		url := ctx.Content[match[2]:match[3]]
		// Skip overly short or obviously non-path patterns
		if len(url) < 2 || url == "/" || url == "./" || url == "../" {
			continue
		}
		// Skip things that look like regex or file system noise
		if strings.HasPrefix(url, "//") {
			continue
		}
		results = append(results, model.Result{
			URL:             url,
			SourceFile:      ctx.FileName,
			SourceLine:      model.LineNumber(ctx.Content, match[2]),
			DetectionMethod: "relative_paths",
			Category:        model.CategorizeURL(url),
			Confidence:      model.ConfMedium,
			TechniqueID:     2,
		})
	}

	return results
}

// technique3: API patterns
func extractAPIPatterns(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result

	// Match API-style paths. Each entry has a quick-check literal that must
	// appear in the content before we bother running the expensive regex.
	type apiPattern struct {
		marker string // literal that must be present for this pattern to fire
		regex  string
	}
	patterns := []apiPattern{
		// /api/v1/..., /api/... (combined into one regex to avoid two 10MB scans)
		{"/api/", `(?:"|'|` + "`" + `|=\s*)((?:/api(?:/v\d+)?(?:/[a-zA-Z0-9_\-~.%{}:]+)*)(?:\?[^"'\x60\s<>]*)?)(?:"|'|` + "`" + `|\s|$|;|,)`},
		// /graphql
		{"/graphql", `(?:"|'|` + "`" + `|=\s*)((?:/graphql(?:/[a-zA-Z0-9_\-~.%{}:]*)*)(?:\?[^"'\x60\s<>]*)?)(?:"|'|` + "`" + `|\s|$|;|,)`},
		// /rest/...
		{"/rest/", `(?:"|'|` + "`" + `|=\s*)((?:/rest(?:/[a-zA-Z0-9_\-~.%{}:]+)+)(?:\?[^"'\x60\s<>]*)?)(?:"|'|` + "`" + `|\s|$|;|,)`},
		// /rpc/...
		{"/rpc/", `(?:"|'|` + "`" + `|=\s*)((?:/rpc(?:/[a-zA-Z0-9_\-~.%{}:]+)+)(?:\?[^"'\x60\s<>]*)?)(?:"|'|` + "`" + `|\s|$|;|,)`},
		// /ws/... (websocket paths)
		{"/ws/", `(?:"|'|` + "`" + `|=\s*)((?:/ws(?:/[a-zA-Z0-9_\-~.%{}:]+)*)(?:\?[^"'\x60\s<>]*)?)(?:"|'|` + "`" + `|\s|$|;|,)`},
		// /v1/..., /v2/... etc (versioned API patterns) — use empty marker; checked below
		{"", `(?:"|'|` + "`" + `|=\s*)((?:/v\d+(?:/[a-zA-Z0-9_\-~.%{}:]+)+)(?:\?[^"'\x60\s<>]*)?)(?:"|'|` + "`" + `|\s|$|;|,)`},
	}

	// Pre-check: does content contain /v followed by a digit?
	hasVersionedAPI := false
	if idx := strings.Index(ctx.Content, "/v"); idx >= 0 {
		// Scan for /v\d pattern
		for i := idx; i < len(ctx.Content)-2; i++ {
			if ctx.Content[i] == '/' && ctx.Content[i+1] == 'v' && ctx.Content[i+2] >= '0' && ctx.Content[i+2] <= '9' {
				hasVersionedAPI = true
				break
			}
			next := strings.Index(ctx.Content[i+1:], "/v")
			if next < 0 {
				break
			}
			i += next
		}
	}

	seen := make(map[string]bool)
	for _, ap := range patterns {
		if ap.marker == "" {
			// Special case for versioned API pattern
			if !hasVersionedAPI {
				continue
			}
		} else if !strings.Contains(ctx.Content, ap.marker) {
			continue
		}
		re := model.GetRegex(ap.regex)
		for _, match := range re.FindAllStringSubmatchIndex(ctx.Content, -1) {
			if match[2] < 0 {
				continue
			}
			url := ctx.Content[match[2]:match[3]]
			if seen[url] {
				continue
			}
			seen[url] = true
			results = append(results, model.Result{
				URL:             url,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(ctx.Content, match[2]),
				DetectionMethod: "api_patterns",
				Category:        model.CatAPIEndpoint,
				Confidence:      model.ConfMedium,
				TechniqueID:     3,
			})
		}
	}

	return results
}

// technique4: Query strings and parameter extraction
func extractQueryStrings(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result

	// Full URLs with query parameters
	re := model.GetRegex(`(?i)((?:https?://[^\s"'\x60<>]+|/[a-zA-Z0-9_\-/.]+)\?[a-zA-Z0-9_\-=&%.+*!~]+(?:#[^\s"'\x60<>]*)?)`)
	paramRe := model.GetRegex(`[?&]([a-zA-Z_][a-zA-Z0-9_\-.]*)=`)

	seen := make(map[string]bool)
	for _, match := range re.FindAllStringSubmatchIndex(ctx.Content, -1) {
		if match[2] < 0 {
			continue
		}
		raw := ctx.Content[match[2]:match[3]]
		url := cleanTrailingURL(raw)
		if url == "" || seen[url] {
			continue
		}
		seen[url] = true

		// Extract parameter names
		var params []string
		paramSeen := make(map[string]bool)
		for _, pm := range paramRe.FindAllStringSubmatch(url, -1) {
			if len(pm) > 1 && !paramSeen[pm[1]] {
				params = append(params, pm[1])
				paramSeen[pm[1]] = true
			}
		}

		results = append(results, model.Result{
			URL:             url,
			SourceFile:      ctx.FileName,
			SourceLine:      model.LineNumber(ctx.Content, match[2]),
			DetectionMethod: "query_strings",
			QueryParams:     params,
			Category:        model.CategorizeURL(url),
			Confidence:      model.ConfMedium,
			TechniqueID:     4,
		})
	}

	// Also look for standalone query params in assignment patterns: param=value&param2=
	// Skip for large files — these are caught by the main pattern above.
	if len(ctx.Content) > 2*1024*1024 {
		return results
	}
	reStandalone := model.GetRegex(`(?:"|'|` + "`" + `)(\?[a-zA-Z_][a-zA-Z0-9_\-]*=[^"'\x60\s]+)(?:"|'|` + "`" + `)`)
	for _, match := range reStandalone.FindAllStringSubmatchIndex(ctx.Content, -1) {
		if match[2] < 0 {
			continue
		}
		raw := ctx.Content[match[2]:match[3]]
		if seen[raw] {
			continue
		}
		seen[raw] = true

		var params []string
		paramSeen := make(map[string]bool)
		for _, pm := range paramRe.FindAllStringSubmatch(raw, -1) {
			if len(pm) > 1 && !paramSeen[pm[1]] {
				params = append(params, pm[1])
				paramSeen[pm[1]] = true
			}
		}

		results = append(results, model.Result{
			URL:             raw,
			SourceFile:      ctx.FileName,
			SourceLine:      model.LineNumber(ctx.Content, match[2]),
			DetectionMethod: "query_strings",
			QueryParams:     params,
			Category:        model.CategorizeURL(raw),
			Confidence:      model.ConfLow,
			TechniqueID:     4,
		})
	}

	return results
}

// technique5: Hash fragments (SPA routing)
func extractHashFragments(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result

	// Quick check: skip if no hash-route marker present
	if !strings.Contains(ctx.Content, "#/") {
		return results
	}

	// SPA hash routes: #/ followed by path segments
	re := model.GetRegex(`(?:"|'|` + "`" + `)(#/[a-zA-Z0-9_\-/.~%:]+(?:\?[^"'\x60\s<>]*)?)(?:"|'|` + "`" + `)`)

	for _, match := range re.FindAllStringSubmatchIndex(ctx.Content, -1) {
		if match[2] < 0 {
			continue
		}
		url := ctx.Content[match[2]:match[3]]
		results = append(results, model.Result{
			URL:             url,
			SourceFile:      ctx.FileName,
			SourceLine:      model.LineNumber(ctx.Content, match[2]),
			DetectionMethod: "hash_fragments",
			Category:        model.CatPageRoute,
			Confidence:      model.ConfMedium,
			TechniqueID:     5,
		})
	}

	// Also match hash routes in href attributes: href="#/..."
	// Skip for large files — the pattern above already catches these.
	if len(ctx.Content) > 2*1024*1024 {
		return results
	}
	reAttr := model.GetRegex(`(?i)href\s*=\s*(?:"|')(#/[a-zA-Z0-9_\-/.~%:]+(?:\?[^"'\s<>]*)?)(?:"|')`)
	for _, match := range reAttr.FindAllStringSubmatchIndex(ctx.Content, -1) {
		if match[2] < 0 {
			continue
		}
		url := ctx.Content[match[2]:match[3]]
		results = append(results, model.Result{
			URL:             url,
			SourceFile:      ctx.FileName,
			SourceLine:      model.LineNumber(ctx.Content, match[2]),
			DetectionMethod: "hash_fragments",
			Category:        model.CatPageRoute,
			Confidence:      model.ConfHigh,
			TechniqueID:     5,
		})
	}

	return results
}

// technique6: Template literals with ${...} interpolation
func extractTemplateLiterals(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result

	// Skip for large files — template literal scanning is expensive and largely
	// redundant with techniques 1-3 which already extract URLs from all contexts.
	if len(ctx.Content) > 2*1024*1024 {
		return results
	}

	// Match backtick-delimited strings containing ${ } and path-like content
	// Use [^}]* inside ${} to match the expression, then allow more content after
	re := model.GetRegex("`" + `([^` + "`" + `]*\$\{[^}]*\}[^` + "`" + `]*)` + "`")

	for _, match := range re.FindAllStringSubmatchIndex(ctx.Content, -1) {
		if match[2] < 0 {
			continue
		}
		raw := ctx.Content[match[2]:match[3]]
		// Must contain a path-like character sequence
		if !strings.Contains(raw, "/") {
			continue
		}
		// Replace ${...} with {{DYNAMIC}}
		dynRe := model.GetRegex(`\$\{[^}]*\}`)
		url := dynRe.ReplaceAllString(raw, "{{DYNAMIC}}")

		// Skip if the result is just {{DYNAMIC}} with no path content
		stripped := strings.ReplaceAll(url, "{{DYNAMIC}}", "")
		if !strings.Contains(stripped, "/") {
			continue
		}

		results = append(results, model.Result{
			URL:             url,
			SourceFile:      ctx.FileName,
			SourceLine:      model.LineNumber(ctx.Content, match[2]),
			DetectionMethod: "template_literals",
			Category:        model.CategorizeURL(url),
			Confidence:      model.ConfMedium,
			TechniqueID:     6,
		})
	}

	return results
}

// technique7: String concatenation path assembly
func extractStringConcatenation(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result

	// Skip for large files — concatenation regex can backtrack on big inputs
	// and results are mostly redundant with other techniques.
	if len(ctx.Content) > 2*1024*1024 {
		return results
	}

	// Match patterns like "path/" + variable + "/more" or 'path/' + variable + '/more'
	// Also handles: baseUrl + "/api/users"
	rePats := []string{
		// "string" + var + "string" patterns
		`((?:(?:"[^"]*"|'[^']*')\s*\+\s*)+(?:[a-zA-Z_$][a-zA-Z0-9_$.]*\s*\+\s*)*(?:"[^"]*"|'[^']*'))`,
		// var + "string" patterns
		`([a-zA-Z_$][a-zA-Z0-9_$.]*\s*\+\s*(?:(?:"[^"]*"|'[^']*')\s*\+?\s*)+)`,
	}

	seen := make(map[string]bool)
	for _, pat := range rePats {
		re := model.GetRegex(pat)
		for _, match := range re.FindAllStringSubmatchIndex(ctx.Content, -1) {
			if match[2] < 0 {
				continue
			}
			raw := ctx.Content[match[2]:match[3]]
			url := reassembleConcatenation(raw)
			if url == "" || seen[url] {
				continue
			}
			// Must contain a path separator to be interesting
			if !strings.Contains(url, "/") {
				continue
			}
			seen[url] = true
			results = append(results, model.Result{
				URL:             url,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(ctx.Content, match[2]),
				DetectionMethod: "string_concatenation",
				Category:        model.CategorizeURL(url),
				Confidence:      model.ConfLow,
				TechniqueID:     7,
			})
		}
	}

	return results
}

// reassembleConcatenation takes a concatenation expression and reassembles the static parts.
func reassembleConcatenation(expr string) string {
	// Split on +
	parts := strings.Split(expr, "+")
	var assembled strings.Builder
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) == 0 {
			continue
		}
		// If it's a string literal, extract the content
		if (part[0] == '"' && part[len(part)-1] == '"') || (part[0] == '\'' && part[len(part)-1] == '\'') {
			if len(part) > 2 {
				assembled.WriteString(part[1 : len(part)-1])
			}
		} else {
			// It's a variable reference
			assembled.WriteString("{{DYNAMIC}}")
		}
	}
	result := assembled.String()
	if result == "" || result == "{{DYNAMIC}}" {
		return ""
	}
	return result
}

// technique8: Data URIs
func extractDataURIs(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result

	// Quick check: skip if no data: marker present
	if !strings.Contains(ctx.Content, "data:") {
		return results
	}

	re := model.GetRegex(`(?i)(data:[a-zA-Z0-9]+/[a-zA-Z0-9+.\-]+(?:;[a-zA-Z0-9\-]+=?[a-zA-Z0-9\-]*)*(?:;base64)?,(?:[A-Za-z0-9+/=]|%[0-9A-Fa-f][0-9A-Fa-f])*)`)

	for _, match := range re.FindAllStringSubmatchIndex(ctx.Content, -1) {
		if match[2] < 0 {
			continue
		}
		url := ctx.Content[match[2]:match[3]]
		// Truncate very long data URIs but keep the type info
		displayURL := url
		if len(displayURL) > 120 {
			displayURL = displayURL[:120] + "..."
		}
		results = append(results, model.Result{
			URL:             displayURL,
			SourceFile:      ctx.FileName,
			SourceLine:      model.LineNumber(ctx.Content, match[2]),
			DetectionMethod: "data_uris",
			Category:        model.CatDataURI,
			Confidence:      model.ConfMedium,
			TechniqueID:     8,
		})
	}

	return results
}

// technique9: Blob and JavaScript URIs
func extractBlobJavascriptURIs(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result

	// Quick check: skip entirely if neither marker is present
	hasBlob := strings.Contains(ctx.Content, "blob:")
	hasJavascript := strings.Contains(ctx.Content, "javascript:")
	if !hasBlob && !hasJavascript {
		return results
	}

	// blob: URIs
	if hasBlob {
		reBlob := model.GetRegex(`(?i)(blob:(?:https?://)?[^\s"'\x60<>]{3,})`)
		for _, match := range reBlob.FindAllStringSubmatchIndex(ctx.Content, -1) {
			if match[2] < 0 {
				continue
			}
			url := cleanTrailingURL(ctx.Content[match[2]:match[3]])
			results = append(results, model.Result{
				URL:             url,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(ctx.Content, match[2]),
				DetectionMethod: "blob_javascript_uris",
				Category:        model.CategorizeURL(url),
				Confidence:      model.ConfMedium,
				TechniqueID:     9,
			})
		}
	}

	// javascript: URIs
	if hasJavascript {
		reJS := model.GetRegex(`(?i)(javascript:[^\s"'\x60<>]+)`)
		for _, match := range reJS.FindAllStringSubmatchIndex(ctx.Content, -1) {
			if match[2] < 0 {
				continue
			}
			url := cleanTrailingURL(ctx.Content[match[2]:match[3]])
			results = append(results, model.Result{
				URL:             url,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(ctx.Content, match[2]),
				DetectionMethod: "blob_javascript_uris",
				Category:        model.CategorizeURL(url),
				Confidence:      model.ConfMedium,
				TechniqueID:     9,
			})
		}
	}

	return results
}

// technique10: Mailto links
func extractMailtoLinks(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result

	// Quick check: skip if no mailto: marker
	if !strings.Contains(ctx.Content, "mailto:") {
		return results
	}

	re := model.GetRegex(`(?i)(mailto:[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}(?:\?[^\s"'\x60<>]*)?)`)

	for _, match := range re.FindAllStringSubmatchIndex(ctx.Content, -1) {
		if match[2] < 0 {
			continue
		}
		url := cleanTrailingURL(ctx.Content[match[2]:match[3]])
		results = append(results, model.Result{
			URL:             url,
			SourceFile:      ctx.FileName,
			SourceLine:      model.LineNumber(ctx.Content, match[2]),
			DetectionMethod: "mailto_links",
			Category:        model.CatMailto,
			Confidence:      model.ConfHigh,
			TechniqueID:     10,
		})
	}

	return results
}

// technique11: IP-based URLs (IPv4 and IPv6)
func extractIPBasedURLs(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result

	// For large files, skip this technique entirely if content doesn't contain
	// any IP-like markers. The absolute URL technique already catches http://IP URLs.
	if len(ctx.Content) > 2*1024*1024 {
		return results
	}

	// IPv4: http(s)://N.N.N.N[:port][/path]
	reIPv4 := model.GetRegex(`(?i)(https?://\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}(?::\d{1,5})?(?:/[^\s"'\x60<>]*)?)`)
	for _, match := range reIPv4.FindAllStringSubmatchIndex(ctx.Content, -1) {
		if match[2] < 0 {
			continue
		}
		url := cleanTrailingURL(ctx.Content[match[2]:match[3]])
		results = append(results, model.Result{
			URL:             url,
			SourceFile:      ctx.FileName,
			SourceLine:      model.LineNumber(ctx.Content, match[2]),
			DetectionMethod: "ip_based_urls",
			Category:        model.CategorizeURL(url),
			Confidence:      model.ConfHigh,
			TechniqueID:     11,
		})
	}

	// IPv6: http(s)://[::1][:port][/path]
	reIPv6 := model.GetRegex(`(?i)(https?://\[[0-9a-fA-F:]+\](?::\d{1,5})?(?:/[^\s"'\x60<>]*)?)`)
	for _, match := range reIPv6.FindAllStringSubmatchIndex(ctx.Content, -1) {
		if match[2] < 0 {
			continue
		}
		url := cleanTrailingURL(ctx.Content[match[2]:match[3]])
		results = append(results, model.Result{
			URL:             url,
			SourceFile:      ctx.FileName,
			SourceLine:      model.LineNumber(ctx.Content, match[2]),
			DetectionMethod: "ip_based_urls",
			Category:        model.CategorizeURL(url),
			Confidence:      model.ConfHigh,
			TechniqueID:     11,
		})
	}

	// Bare IP addresses in quotes that look like hosts (without protocol)
	reBareIP := model.GetRegex(`(?:"|'|` + "`" + `)(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}(?::\d{1,5})?)(?:"|'|` + "`" + `)`)
	for _, match := range reBareIP.FindAllStringSubmatchIndex(ctx.Content, -1) {
		if match[2] < 0 {
			continue
		}
		ip := ctx.Content[match[2]:match[3]]
		// Validate it looks like a real IP (each octet 0-255)
		if !isValidIPv4(ip) {
			continue
		}
		results = append(results, model.Result{
			URL:             ip,
			SourceFile:      ctx.FileName,
			SourceLine:      model.LineNumber(ctx.Content, match[2]),
			DetectionMethod: "ip_based_urls",
			Category:        model.CatInternalInfra,
			Confidence:      model.ConfLow,
			TechniqueID:     11,
		})
	}

	return results
}

// isValidIPv4 checks if a string like "192.168.1.1" or "192.168.1.1:8080" has valid octets.
func isValidIPv4(s string) bool {
	host := s
	if idx := strings.IndexByte(s, ':'); idx >= 0 {
		host = s[:idx]
	}
	parts := strings.Split(host, ".")
	if len(parts) != 4 {
		return false
	}
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 || n > 255 {
			return false
		}
	}
	return true
}

// technique12: Encoded URLs (URL-encoded, Unicode, hex, HTML entities)
func extractEncodedURLs(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result

	// Skip entirely for large files — encoded URL scanning is expensive and
	// largely redundant with the absolute URL technique which already catches
	// most encoded URLs. The rare encoded paths in large bundles are not worth
	// the O(n) regex scans.
	if len(ctx.Content) > 2*1024*1024 {
		return results
	}

	// For smaller files, only run sub-patterns when their encoding markers are present.
	hasPercent := strings.Contains(ctx.Content, "%2F") || strings.Contains(ctx.Content, "%2f") || strings.Contains(ctx.Content, "%3A") || strings.Contains(ctx.Content, "%3a")
	hasUnicode := strings.Contains(ctx.Content, "\\u00")
	hasHexEsc := strings.Contains(ctx.Content, "\\x2")
	hasHTMLEnt := strings.Contains(ctx.Content, "&#")

	// URL-encoded strings containing %2F (/) which indicates encoded paths
	if hasPercent {
		reURLEnc := model.GetRegex(`(?i)((?:https?(?:%3A|:)(?:%2F|/){2}|%2F)[^\s"'\x60<>]*%[0-9A-Fa-f]{2}[^\s"'\x60<>]*)`)
		for _, match := range reURLEnc.FindAllStringSubmatchIndex(ctx.Content, -1) {
			if match[2] < 0 {
				continue
			}
			raw := ctx.Content[match[2]:match[3]]
			decoded := decodeURLEncoding(raw)
			if decoded != raw {
				// First pass decoded something; try double decode
				decoded = decodeURLEncoding(decoded)
			}
			decoded = cleanTrailingURL(decoded)
			if decoded == "" {
				continue
			}
			results = append(results, model.Result{
				URL:             decoded,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(ctx.Content, match[2]),
				DetectionMethod: "encoded_urls",
				Category:        model.CategorizeURL(decoded),
				Confidence:      model.ConfMedium,
				TechniqueID:     12,
			})
		}
	}

	// Unicode-escaped strings: \u002F = /
	if hasUnicode {
		reUnicode := model.GetRegex(`((?:\\u[0-9A-Fa-f]{4}){3,})`)
		for _, match := range reUnicode.FindAllStringSubmatchIndex(ctx.Content, -1) {
			if match[2] < 0 {
				continue
			}
			raw := ctx.Content[match[2]:match[3]]
			decoded := decodeUnicodeEscapes(raw)
			if !strings.Contains(decoded, "/") && !strings.Contains(decoded, ".") {
				continue
			}
			decoded = cleanTrailingURL(decoded)
			if decoded == "" || len(decoded) < 3 {
				continue
			}
			results = append(results, model.Result{
				URL:             decoded,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(ctx.Content, match[2]),
				DetectionMethod: "encoded_urls",
				Category:        model.CategorizeURL(decoded),
				Confidence:      model.ConfMedium,
				TechniqueID:     12,
			})
		}
	}

	// Hex-escaped strings: \x2F = /
	if hasHexEsc {
		reHex := model.GetRegex(`((?:\\x[0-9A-Fa-f]{2}){3,})`)
		for _, match := range reHex.FindAllStringSubmatchIndex(ctx.Content, -1) {
			if match[2] < 0 {
				continue
			}
			raw := ctx.Content[match[2]:match[3]]
			decoded := decodeHexEscapes(raw)
			if !strings.Contains(decoded, "/") && !strings.Contains(decoded, ".") {
				continue
			}
			decoded = cleanTrailingURL(decoded)
			if decoded == "" || len(decoded) < 3 {
				continue
			}
			results = append(results, model.Result{
				URL:             decoded,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(ctx.Content, match[2]),
				DetectionMethod: "encoded_urls",
				Category:        model.CategorizeURL(decoded),
				Confidence:      model.ConfMedium,
				TechniqueID:     12,
			})
		}
	}

	// HTML entity-encoded: &#x2F; = /, &#47; = /
	if hasHTMLEnt {
		reHTML := model.GetRegex(`((?:&#x?[0-9A-Fa-f]+;){3,})`)
		for _, match := range reHTML.FindAllStringSubmatchIndex(ctx.Content, -1) {
			if match[2] < 0 {
				continue
			}
			raw := ctx.Content[match[2]:match[3]]
			decoded := decodeHTMLEntities(raw)
			if !strings.Contains(decoded, "/") && !strings.Contains(decoded, ".") {
				continue
			}
			decoded = cleanTrailingURL(decoded)
			if decoded == "" || len(decoded) < 3 {
				continue
			}
			results = append(results, model.Result{
				URL:             decoded,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(ctx.Content, match[2]),
				DetectionMethod: "encoded_urls",
				Category:        model.CategorizeURL(decoded),
				Confidence:      model.ConfMedium,
				TechniqueID:     12,
			})
		}
	}

	// Also look for full strings that contain a mix of encoded and literal characters
	if hasPercent || hasUnicode || hasHexEsc || hasHTMLEnt {
		reMixed := model.GetRegex(`(?:"|'|` + "`" + `)([^"'\x60]*(?:%2[Ff]|\\u002[Ff]|\\x2[Ff]|&#(?:x2[Ff]|47);)[^"'\x60]*)(?:"|'|` + "`" + `)`)
		seen := make(map[string]bool)
		for _, match := range reMixed.FindAllStringSubmatchIndex(ctx.Content, -1) {
			if match[2] < 0 {
				continue
			}
			raw := ctx.Content[match[2]:match[3]]
			decoded := decodeAllEncodings(raw)
			decoded = cleanTrailingURL(decoded)
			if decoded == "" || len(decoded) < 3 || seen[decoded] {
				continue
			}
			seen[decoded] = true
			results = append(results, model.Result{
				URL:             decoded,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(ctx.Content, match[2]),
				DetectionMethod: "encoded_urls",
				Category:        model.CategorizeURL(decoded),
				Confidence:      model.ConfMedium,
			TechniqueID:     12,
			})
		}
	}

	return results
}

// decodeURLEncoding decodes %XX sequences in a string.
func decodeURLEncoding(s string) string {
	var buf strings.Builder
	buf.Grow(len(s))
	i := 0
	for i < len(s) {
		if i+2 < len(s) && s[i] == '%' && isHexDigit(s[i+1]) && isHexDigit(s[i+2]) {
			hi := unhex(s[i+1])
			lo := unhex(s[i+2])
			buf.WriteByte(byte(hi<<4 | lo))
			i += 3
		} else {
			buf.WriteByte(s[i])
			i++
		}
	}
	return buf.String()
}

// decodeUnicodeEscapes decodes \uXXXX sequences.
func decodeUnicodeEscapes(s string) string {
	re := model.GetRegex(`\\u([0-9A-Fa-f]{4})`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		hex := match[2:]
		codePoint, err := strconv.ParseInt(hex, 16, 32)
		if err != nil {
			return match
		}
		return string(rune(codePoint))
	})
}

// decodeHexEscapes decodes \xXX sequences.
func decodeHexEscapes(s string) string {
	re := model.GetRegex(`\\x([0-9A-Fa-f]{2})`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		hex := match[2:]
		val, err := strconv.ParseInt(hex, 16, 32)
		if err != nil {
			return match
		}
		return string(rune(val))
	})
}

// decodeHTMLEntities decodes &#xHH; and &#DD; sequences (and named entities).
func decodeHTMLEntities(s string) string {
	// First use Go's built-in HTML unescaper for named entities
	result := html.UnescapeString(s)
	// Then handle numeric entities that may remain
	reHex := model.GetRegex(`&#[xX]([0-9A-Fa-f]+);`)
	result = reHex.ReplaceAllStringFunc(result, func(match string) string {
		// Extract hex value between &#x and ;
		inner := match[3 : len(match)-1]
		val, err := strconv.ParseInt(inner, 16, 32)
		if err != nil {
			return match
		}
		return string(rune(val))
	})
	reDec := model.GetRegex(`&#(\d+);`)
	result = reDec.ReplaceAllStringFunc(result, func(match string) string {
		inner := match[2 : len(match)-1]
		val, err := strconv.ParseInt(inner, 10, 32)
		if err != nil {
			return match
		}
		return string(rune(val))
	})
	return result
}

// decodeAllEncodings applies all decoding functions in sequence.
func decodeAllEncodings(s string) string {
	result := s
	result = decodeURLEncoding(result)
	result = decodeUnicodeEscapes(result)
	result = decodeHexEscapes(result)
	result = decodeHTMLEntities(result)
	// Double-decode URL encoding in case of double-encoding
	if strings.Contains(result, "%") {
		result = decodeURLEncoding(result)
	}
	return result
}

func isHexDigit(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func unhex(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}

// technique13: Custom protocol handlers
func extractCustomProtocols(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result

	// Quick check: only run the expensive alternation regex if at least one
	// of these protocol prefixes appears in the content. This avoids a 2.5s
	// regex scan on large files that contain none of these schemes.
	// Use a fast literal scan rather than ToLower on the whole content.
	protocolMarkers := []string{
		"ndroid-app://", "os-app://", "ntent://", "eeplink://",
		"fb://", "witter://", "hatsapp://", "tg://", "lack://",
		"iber://", "line://", "tel://", "sms://", "geo://",
		"aps://", "arket://", "tms-app", "ebcal://",
		"vn+ssh://", "ssh://", "s3://", "gs://", "az://",
	}
	hasAny := false
	for _, p := range protocolMarkers {
		if strings.Contains(ctx.Content, p) {
			hasAny = true
			break
		}
	}
	if !hasAny {
		return results
	}

	// Match android-app://, ios-app://, intent://, deeplink://, and other custom schemes
	re := model.GetRegex(`(?i)((?:android-app|ios-app|intent|deeplink|fb|twitter|whatsapp|tg|slack|viber|line|tel|sms|geo|maps|market|itms-apps?|webcal|svn\+ssh|ssh|s3|gs|az)://[^\s"'\x60<>)]+)`)

	for _, match := range re.FindAllStringSubmatchIndex(ctx.Content, -1) {
		if match[2] < 0 {
			continue
		}
		url := cleanTrailingURL(ctx.Content[match[2]:match[3]])
		results = append(results, model.Result{
			URL:             url,
			SourceFile:      ctx.FileName,
			SourceLine:      model.LineNumber(ctx.Content, match[2]),
			DetectionMethod: "custom_protocols",
			Category:        model.CatCustomProtocol,
			Confidence:      model.ConfMedium,
			TechniqueID:     13,
		})
	}

	// Intent URIs with more complex syntax: intent:#Intent;scheme=...;end
	reIntent := model.GetRegex(`(?i)(intent://[^\s"'\x60<>]*(?:;[^\s"'\x60<>]*)*;end)`)
	for _, match := range reIntent.FindAllStringSubmatchIndex(ctx.Content, -1) {
		if match[2] < 0 {
			continue
		}
		url := ctx.Content[match[2]:match[3]]
		results = append(results, model.Result{
			URL:             url,
			SourceFile:      ctx.FileName,
			SourceLine:      model.LineNumber(ctx.Content, match[2]),
			DetectionMethod: "custom_protocols",
			Category:        model.CatCustomProtocol,
			Confidence:      model.ConfHigh,
			TechniqueID:     13,
		})
	}

	return results
}

// technique14: Path patterns inside string literals
func extractStringLiterals(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result

	// Skip for large files — this scans every quoted string in the entire file
	// and is highly redundant with techniques 1 (absolute URLs), 2 (relative paths),
	// and 3 (API patterns) which already extract paths from quoted contexts.
	if len(ctx.Content) > 2*1024*1024 {
		return results
	}

	// Match paths inside double-quoted strings
	reDouble := model.GetRegex(`"((?:/[a-zA-Z0-9_\-~.%@:{}]+)+(?:\.[a-zA-Z0-9]+)?(?:\?[^"\s]*)?)"`)
	// Match paths inside single-quoted strings
	reSingle := model.GetRegex(`'((?:/[a-zA-Z0-9_\-~.%@:{}]+)+(?:\.[a-zA-Z0-9]+)?(?:\?[^'\s]*)?)'`)
	// Match paths inside backtick strings (excluding template literal expressions already caught)
	reBacktick := model.GetRegex("`" + `((?:/[a-zA-Z0-9_\-~.%@:{}]+)+(?:\.[a-zA-Z0-9]+)?(?:\?[^` + "`" + `\s]*)?)` + "`")

	seen := make(map[string]bool)
	extractFromMatches := func(re *regexp.Regexp) {
		for _, match := range re.FindAllStringSubmatchIndex(ctx.Content, -1) {
			if match[2] < 0 {
				continue
			}
			url := ctx.Content[match[2]:match[3]]
			// Must contain at least one / and look like a path
			if !strings.Contains(url, "/") || len(url) < 2 {
				continue
			}
			// Skip common false positives
			if isCommonFalsePositive(url) {
				continue
			}
			if seen[url] {
				continue
			}
			seen[url] = true
			results = append(results, model.Result{
				URL:             url,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(ctx.Content, match[2]),
				DetectionMethod: "string_literals",
				Category:        model.CategorizeURL(url),
				Confidence:      model.ConfLow,
				TechniqueID:     14,
			})
		}
	}

	extractFromMatches(reDouble)
	extractFromMatches(reSingle)
	extractFromMatches(reBacktick)

	return results
}

// isCommonFalsePositive filters out strings that look like paths but aren't URLs.
func isCommonFalsePositive(s string) bool {
	// Skip MIME types
	if model.GetRegex(`^[a-zA-Z]+/[a-zA-Z0-9.+\-]+$`).MatchString(s) {
		return true
	}
	// Skip date patterns
	if model.GetRegex(`^\d{1,4}/\d{1,2}/\d{1,4}$`).MatchString(s) {
		return true
	}
	// Skip regex-like patterns
	if strings.HasPrefix(s, "/^") || strings.HasSuffix(s, "$/") {
		return true
	}
	// Skip pure file system paths on Windows
	if model.GetRegex(`^[A-Z]:\\`).MatchString(s) {
		return true
	}
	// Skip math expressions
	if model.GetRegex(`^\d+/\d+$`).MatchString(s) {
		return true
	}
	// Skip XML/HTML closing tags
	if strings.HasPrefix(s, "/") && model.GetRegex(`^/[a-zA-Z]+$`).MatchString(s) {
		return true
	}
	return false
}

// cleanTrailingURL removes trailing characters that are commonly not part of a URL.
func cleanTrailingURL(url string) string {
	if len(url) == 0 {
		return ""
	}

	// Characters that commonly appear after a URL but aren't part of it
	trailingChars := []byte{'.', ',', ';', ':', '!', '?', ')', ']', '}', '>', '\'', '"', '`'}
	changed := true
	for changed {
		changed = false
		if len(url) == 0 {
			break
		}
		last := url[len(url)-1]
		for _, c := range trailingChars {
			if last == c {
				// Special case: don't strip ) if there's a matching ( in the URL
				if c == ')' && strings.Count(url, "(") > strings.Count(url, ")")-1 {
					continue
				}
				url = url[:len(url)-1]
				changed = true
				break
			}
		}
	}

	// Remove trailing whitespace
	url = strings.TrimRight(url, " \t\n\r")

	return url
}

// Ensure fmt is used (for potential future debug logging).
var _ = fmt.Sprintf
