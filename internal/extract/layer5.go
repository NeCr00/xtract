package extract

import "github.com/pentester/xtract/internal/model"

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
func extractSubdomains(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	// Match fully-qualified domain names: at least one subdomain level (a.b.tld)
	re := model.GetRegex(`(?i)([a-z0-9](?:[a-z0-9\-]{0,61}[a-z0-9])?\.){2,}[a-z]{2,63}`)
	seen := make(map[string]bool)
	matches := re.FindAllStringIndex(ctx.Content, -1)
	for _, loc := range matches {
		domain := ctx.Content[loc[0]:loc[1]]
		lower := toLower(domain)
		if seen[lower] {
			continue
		}
		seen[lower] = true
		results = append(results, model.Result{
			URL:             domain,
			SourceFile:      ctx.FileName,
			SourceLine:      model.LineNumber(ctx.Content, loc[0]),
			DetectionMethod: "subdomain_extract",
			Category:        model.CategorizeURL(domain),
			Confidence:      model.ConfMedium,
			TechniqueID:     50,
		})
	}
	return results
}

// technique 51: CDN/asset domain references
func extractCDNAssetDomains(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	patterns := []struct {
		regex string
		name  string
	}{
		// S3 bucket styles
		{`(?i)https?://[a-z0-9][a-z0-9.\-]*\.s3[.\-][a-z0-9.\-]*\.amazonaws\.com[^\s"'<>]*`, "s3_bucket"},
		{`(?i)https?://s3[.\-][a-z0-9.\-]*\.amazonaws\.com/[^\s"'<>]*`, "s3_path_style"},
		{`(?i)https?://s3\.amazonaws\.com/[^\s"'<>]*`, "s3_global"},
		// CloudFront
		{`(?i)https?://[a-z0-9][a-z0-9\-]*\.cloudfront\.net[^\s"'<>]*`, "cloudfront"},
		// Azure Blob Storage
		{`(?i)https?://[a-z0-9][a-z0-9\-]*\.blob\.core\.windows\.net[^\s"'<>]*`, "azure_blob"},
		// Google Cloud Storage
		{`(?i)https?://storage\.googleapis\.com/[^\s"'<>]*`, "gcs"},
		// Firebase Storage
		{`(?i)https?://firebasestorage\.googleapis\.com/[^\s"'<>]*`, "firebase_storage"},
		// Akamai
		{`(?i)https?://[a-z0-9][a-z0-9\-]*\.akamaihd\.net[^\s"'<>]*`, "akamai_hd"},
		{`(?i)https?://[a-z0-9][a-z0-9\-]*\.akamaized\.net[^\s"'<>]*`, "akamai_ized"},
		{`(?i)https?://[a-z0-9][a-z0-9\-]*\.akamai\.net[^\s"'<>]*`, "akamai"},
		// Fastly
		{`(?i)https?://[a-z0-9][a-z0-9\-]*\.fastly\.net[^\s"'<>]*`, "fastly"},
		{`(?i)https?://[a-z0-9][a-z0-9\-]*\.fastlylb\.net[^\s"'<>]*`, "fastly_lb"},
		// Cloudflare CDN
		{`(?i)https?://[a-z0-9][a-z0-9\-]*\.cdninstagram\.com[^\s"'<>]*`, "cloudflare_cdn"},
		{`(?i)https?://cdnjs\.cloudflare\.com/[^\s"'<>]*`, "cloudflare_cdnjs"},
		{`(?i)https?://[a-z0-9][a-z0-9\-]*\.r2\.cloudflarestorage\.com[^\s"'<>]*`, "cloudflare_r2"},
	}

	seen := make(map[string]bool)
	for _, p := range patterns {
		re := model.GetRegex(p.regex)
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
				SourceLine:      model.LineNumber(ctx.Content, loc[0]),
				DetectionMethod: "cdn_asset_domains",
				Category:        model.CatCloudResource,
				Confidence:      model.ConfHigh,
				TechniqueID:     51,
			})
		}
	}
	return results
}

// technique 52: internal hostname patterns
func extractInternalHostnames(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	patterns := []string{
		// Internal TLD-style domains
		`(?i)https?://[a-z0-9][a-z0-9.\-]*\.internal(?:[:/\?#][^\s"'<>]*|)`,
		`(?i)https?://[a-z0-9][a-z0-9.\-]*\.local(?:[:/\?#][^\s"'<>]*|)`,
		`(?i)https?://[a-z0-9][a-z0-9.\-]*\.staging(?:[:/\?#][^\s"'<>]*|)`,
		`(?i)https?://[a-z0-9][a-z0-9.\-]*\.dev(?:[:/\?#][^\s"'<>]*|)`,
		`(?i)https?://[a-z0-9][a-z0-9.\-]*\.test(?:[:/\?#][^\s"'<>]*|)`,
		`(?i)https?://[a-z0-9][a-z0-9.\-]*\.corp(?:[:/\?#][^\s"'<>]*|)`,
		`(?i)https?://[a-z0-9][a-z0-9.\-]*\.private(?:[:/\?#][^\s"'<>]*|)`,
		// Hostname-only matches (without protocol, in string literals)
		`(?i)["'][a-z0-9][a-z0-9.\-]*\.internal["']`,
		`(?i)["'][a-z0-9][a-z0-9.\-]*\.local["']`,
		`(?i)["'][a-z0-9][a-z0-9.\-]*\.staging["']`,
		`(?i)["'][a-z0-9][a-z0-9.\-]*\.corp["']`,
		`(?i)["'][a-z0-9][a-z0-9.\-]*\.private["']`,
		// Private IPs in URLs
		`https?://10\.\d{1,3}\.\d{1,3}\.\d{1,3}(?::\d+)?[^\s"'<>]*`,
		`https?://172\.(?:1[6-9]|2\d|3[01])\.\d{1,3}\.\d{1,3}(?::\d+)?[^\s"'<>]*`,
		`https?://192\.168\.\d{1,3}\.\d{1,3}(?::\d+)?[^\s"'<>]*`,
		`https?://127\.0\.0\.1(?::\d+)?[^\s"'<>]*`,
		`https?://localhost(?::\d+)?[^\s"'<>]*`,
		// Private IPs in string literals (without protocol)
		`["']10\.\d{1,3}\.\d{1,3}\.\d{1,3}(?::\d+)?[^\s"'<>]*["']`,
		`["']172\.(?:1[6-9]|2\d|3[01])\.\d{1,3}\.\d{1,3}(?::\d+)?[^\s"'<>]*["']`,
		`["']192\.168\.\d{1,3}\.\d{1,3}(?::\d+)?[^\s"'<>]*["']`,
		`["']127\.0\.0\.1(?::\d+)?[^\s"'<>]*["']`,
		`["']localhost(?::\d+)?[^\s"'<>]*["']`,
	}

	seen := make(map[string]bool)
	for _, pattern := range patterns {
		re := model.GetRegex(pattern)
		matches := re.FindAllStringIndex(ctx.Content, -1)
		for _, loc := range matches {
			url := ctx.Content[loc[0]:loc[1]]
			// Strip surrounding quotes if present
			url = stripQuotes(url)
			if seen[url] {
				continue
			}
			seen[url] = true
			results = append(results, model.Result{
				URL:             url,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(ctx.Content, loc[0]),
				DetectionMethod: "internal_hostnames",
				Category:        model.CatInternalInfra,
				Confidence:      model.ConfHigh,
				TechniqueID:     52,
			})
		}
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
			SourceLine:      model.LineNumber(ctx.Content, loc[0]),
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
