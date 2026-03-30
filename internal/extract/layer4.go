package extract

import (
	"strings"

	"github.com/NeCr00/xtract/internal/model"
)

// ExtractLayer4 runs all Layer 4 (Configuration & Metadata) techniques and returns results.
func ExtractLayer4(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	results = append(results, extractEnvVariables(ctx)...)
	results = append(results, extractConfigObjects(ctx)...)
	results = append(results, extractWebpackPublicPath(ctx)...)
	results = append(results, extractSourceMapRefs(ctx)...)
	results = append(results, extractBaseTags(ctx)...)
	results = append(results, extractManifestFiles(ctx)...)
	results = append(results, extractSitemapRobots(ctx)...)
	results = append(results, extractOpenAPISwagger(ctx)...)
	results = append(results, extractFeatureFlags(ctx)...)
	return results
}

// extractEnvVariables detects environment variable references like process.env.*
// and import.meta.env.* patterns, extracting variable names and assigned values.
func extractEnvVariables(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	// Match process.env.VARIABLE_NAME patterns
	envPrefixes := []string{
		`process\.env\.REACT_APP_[A-Z_][A-Z0-9_]*`,
		`process\.env\.NEXT_PUBLIC_[A-Z_][A-Z0-9_]*`,
		`process\.env\.VITE_[A-Z_][A-Z0-9_]*`,
		`process\.env\.VUE_APP_[A-Z_][A-Z0-9_]*`,
		`process\.env\.API_URL`,
		`process\.env\.[A-Z_][A-Z0-9_]*`,
		`import\.meta\.env\.VITE_[A-Z_][A-Z0-9_]*`,
	}

	for _, pat := range envPrefixes {
		// Match the env reference, optionally followed by assignment to a string
		re := model.GetRegex(`(` + pat + `)(?:\s*(?:=|:|\|\|)\s*["']([^"']+)["'])?`)
		matches := re.FindAllStringSubmatchIndex(content, -1)
		for _, m := range matches {
			if len(m) >= 4 {
				varName := content[m[2]:m[3]]
				url := varName
				// If a string value was assigned, use that instead
				if len(m) >= 6 && m[4] >= 0 && m[5] >= 0 {
					url = content[m[4]:m[5]]
				}
				results = append(results, model.Result{
					URL:             url,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(content, m[0]),
					DetectionMethod: "env_variables",
					Category:        model.CategorizeURL(url),
					Confidence:      model.ConfMedium,
					TechniqueID:     41,
				})
			}
		}
	}

	return results
}

// extractConfigObjects detects configuration objects with URL-related property keys
// and extracts their string values.
func extractConfigObjects(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	// Config key names that commonly hold URLs
	configKeys := []string{
		"baseURL", "baseUrl", "apiUrl", "apiURL",
		"endpoint", "host", "hostname", "origin",
		"server", "serverUrl", "backend", "backendUrl",
		"gateway", "proxy", "webhook", "webhookUrl",
		"serviceUrl", "url",
	}

	for _, key := range configKeys {
		// Match key: "value" or key: 'value' in object notation
		// Also match "key": "value" (JSON-style)
		re := model.GetRegex(`(?:["']?` + escapeRegexLiteral(key) + `["']?\s*:\s*["'])((?:https?://|/)[^"']+)["']`)
		matches := re.FindAllStringSubmatchIndex(content, -1)
		for _, m := range matches {
			if len(m) >= 4 {
				value := content[m[2]:m[3]]
				results = append(results, model.Result{
					URL:             value,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(content, m[0]),
					DetectionMethod: "config_objects",
					Category:        model.CategorizeURL(value),
					Confidence:      model.ConfMedium,
					TechniqueID:     42,
				})
			}
		}
	}

	return results
}

// escapeRegexLiteral escapes special regex characters in a literal string.
func escapeRegexLiteral(s string) string {
	special := `\.+*?^${}()|[]`
	var b strings.Builder
	for _, c := range s {
		if strings.ContainsRune(special, c) {
			b.WriteRune('\\')
		}
		b.WriteRune(c)
	}
	return b.String()
}

// extractWebpackPublicPath detects Webpack publicPath configuration patterns.
func extractWebpackPublicPath(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	patterns := []string{
		// __webpack_public_path__ = "..."
		`__webpack_public_path__\s*=\s*["']([^"']+)["']`,
		// output.publicPath or publicPath: "..."
		`publicPath\s*:\s*["']([^"']+)["']`,
		// __webpack_require__.p = "..."
		`__webpack_require__\.p\s*=\s*["']([^"']+)["']`,
	}

	for _, pat := range patterns {
		re := model.GetRegex(pat)
		matches := re.FindAllStringSubmatchIndex(content, -1)
		for _, m := range matches {
			if len(m) >= 4 {
				path := content[m[2]:m[3]]
				results = append(results, model.Result{
					URL:             path,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(content, m[0]),
					DetectionMethod: "webpack_public_path",
					Category:        model.CategorizeURL(path),
					Confidence:      model.ConfMedium,
					TechniqueID:     43,
				})
			}
		}
	}

	return results
}

// extractSourceMapRefs detects source map references in JavaScript files.
func extractSourceMapRefs(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	re := model.GetRegex(`//[#@]\s*sourceMappingURL\s*=\s*(\S+)`)
	matches := re.FindAllStringSubmatchIndex(content, -1)
	for _, m := range matches {
		if len(m) >= 4 {
			mapURL := content[m[2]:m[3]]
			results = append(results, model.Result{
				URL:             mapURL,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(content, m[0]),
				DetectionMethod: "sourcemap_refs",
				Category:        model.CatSourceMap,
				Confidence:      model.ConfHigh,
				TechniqueID:     44,
			})
		}
	}

	return results
}

// extractBaseTags detects HTML <base href="..."> tags.
func extractBaseTags(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	re := model.GetRegex(`(?i)<base\s+[^>]*href\s*=\s*["']([^"']+)["']`)
	matches := re.FindAllStringSubmatchIndex(content, -1)
	for _, m := range matches {
		if len(m) >= 4 {
			href := content[m[2]:m[3]]
			results = append(results, model.Result{
				URL:             href,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(content, m[0]),
				DetectionMethod: "base_tags",
				Category:        model.CategorizeURL(href),
				Confidence:      model.ConfHigh,
				TechniqueID:     45,
			})
		}
	}

	return results
}

// extractManifestFiles detects URLs in web app manifest files (JSON with
// start_url, scope, icons, related_applications, shortcuts).
func extractManifestFiles(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	// Quick check: is this likely a manifest file?
	isManifest := strings.Contains(content, "start_url") ||
		strings.Contains(content, "\"scope\"") ||
		strings.Contains(content, "\"short_name\"")
	if !isManifest {
		return results
	}

	// Extract start_url
	startURLRe := model.GetRegex(`"start_url"\s*:\s*"([^"]+)"`)
	for _, m := range startURLRe.FindAllStringSubmatchIndex(content, -1) {
		if len(m) >= 4 {
			results = append(results, model.Result{
				URL:             content[m[2]:m[3]],
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(content, m[0]),
				DetectionMethod: "manifest_files",
				Category:        model.CatPageRoute,
				Confidence:      model.ConfHigh,
				TechniqueID:     46,
			})
		}
	}

	// Extract scope
	scopeRe := model.GetRegex(`"scope"\s*:\s*"([^"]+)"`)
	for _, m := range scopeRe.FindAllStringSubmatchIndex(content, -1) {
		if len(m) >= 4 {
			results = append(results, model.Result{
				URL:             content[m[2]:m[3]],
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(content, m[0]),
				DetectionMethod: "manifest_files",
				Category:        model.CatPageRoute,
				Confidence:      model.ConfHigh,
				TechniqueID:     46,
			})
		}
	}

	// Extract icons[].src
	iconSrcRe := model.GetRegex(`"src"\s*:\s*"([^"]+)"`)
	for _, m := range iconSrcRe.FindAllStringSubmatchIndex(content, -1) {
		if len(m) >= 4 {
			results = append(results, model.Result{
				URL:             content[m[2]:m[3]],
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(content, m[0]),
				DetectionMethod: "manifest_files",
				Category:        model.CatStaticAsset,
				Confidence:      model.ConfHigh,
				TechniqueID:     46,
			})
		}
	}

	// Extract related_applications[].url
	relatedURLRe := model.GetRegex(`"url"\s*:\s*"([^"]+)"`)
	for _, m := range relatedURLRe.FindAllStringSubmatchIndex(content, -1) {
		if len(m) >= 4 {
			results = append(results, model.Result{
				URL:             content[m[2]:m[3]],
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(content, m[0]),
				DetectionMethod: "manifest_files",
				Category:        model.CategorizeURL(content[m[2]:m[3]]),
				Confidence:      model.ConfHigh,
				TechniqueID:     46,
			})
		}
	}

	// Extract shortcuts[].url (use the same pattern but labeled differently)
	// Already captured above by relatedURLRe, so no additional extraction needed.

	return results
}

// extractSitemapRobots detects references to sitemap.xml and robots.txt,
// and URLs found within inline sitemap or robots content.
func extractSitemapRobots(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	// References to sitemap.xml or robots.txt as URLs
	refRe := model.GetRegex(`["']((?:https?://)?[^"'\s]*(?:sitemap[^"'\s]*\.xml|robots\.txt))["']`)
	for _, m := range refRe.FindAllStringSubmatchIndex(content, -1) {
		if len(m) >= 4 {
			url := content[m[2]:m[3]]
			results = append(results, model.Result{
				URL:             url,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(content, m[0]),
				DetectionMethod: "sitemap_robots",
				Category:        model.CategorizeURL(url),
				Confidence:      model.ConfMedium,
				TechniqueID:     47,
			})
		}
	}

	// Sitemap XML inline: <loc>...</loc>
	locRe := model.GetRegex(`<loc>\s*(https?://[^<\s]+)\s*</loc>`)
	for _, m := range locRe.FindAllStringSubmatchIndex(content, -1) {
		if len(m) >= 4 {
			url := content[m[2]:m[3]]
			results = append(results, model.Result{
				URL:             url,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(content, m[0]),
				DetectionMethod: "sitemap_robots",
				Category:        model.CategorizeURL(url),
				Confidence:      model.ConfMedium,
				TechniqueID:     47,
			})
		}
	}

	// Robots.txt inline: Sitemap: <url> and Disallow/Allow directives
	sitemapDirectiveRe := model.GetRegex(`(?im)^Sitemap:\s*(https?://\S+)`)
	for _, m := range sitemapDirectiveRe.FindAllStringSubmatchIndex(content, -1) {
		if len(m) >= 4 {
			url := content[m[2]:m[3]]
			results = append(results, model.Result{
				URL:             url,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(content, m[0]),
				DetectionMethod: "sitemap_robots",
				Category:        model.CategorizeURL(url),
				Confidence:      model.ConfMedium,
				TechniqueID:     47,
			})
		}
	}

	// Robots.txt Disallow/Allow paths
	robotsPathRe := model.GetRegex(`(?i)(?:Disallow|Allow):\s*(/\S+)`)
	for _, m := range robotsPathRe.FindAllStringSubmatchIndex(content, -1) {
		if len(m) >= 4 {
			path := content[m[2]:m[3]]
			results = append(results, model.Result{
				URL:             path,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(content, m[0]),
				DetectionMethod: "sitemap_robots",
				Category:        model.CatPageRoute,
				Confidence:      model.ConfMedium,
				TechniqueID:     47,
			})
		}
	}

	return results
}

// extractOpenAPISwagger detects OpenAPI/Swagger path definitions and
// swagger-ui references.
func extractOpenAPISwagger(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	// Quick check for OpenAPI/Swagger markers
	isOpenAPI := strings.Contains(content, "\"openapi\"") ||
		strings.Contains(content, "\"swagger\"") ||
		strings.Contains(content, "'openapi'") ||
		strings.Contains(content, "'swagger'") ||
		strings.Contains(content, "openapi:") ||
		strings.Contains(content, "swagger:")
	if !isOpenAPI && !strings.Contains(content, "swagger-ui") {
		return results
	}

	// Extract path definitions: "paths": { "/api/...": { ... } }
	// Match keys under "paths" that start with /
	pathRe := model.GetRegex(`"(\/[^"]+)"\s*:\s*\{`)
	for _, m := range pathRe.FindAllStringSubmatchIndex(content, -1) {
		if len(m) >= 4 {
			path := content[m[2]:m[3]]
			results = append(results, model.Result{
				URL:             path,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(content, m[0]),
				DetectionMethod: "openapi_swagger",
				Category:        model.CatAPIEndpoint,
				Confidence:      model.ConfHigh,
				TechniqueID:     48,
			})
		}
	}

	// Detect swagger-ui references
	swaggerUIRe := model.GetRegex(`["']((?:https?://)?[^"'\s]*swagger-ui[^"'\s]*)["']`)
	for _, m := range swaggerUIRe.FindAllStringSubmatchIndex(content, -1) {
		if len(m) >= 4 {
			url := content[m[2]:m[3]]
			results = append(results, model.Result{
				URL:             url,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(content, m[0]),
				DetectionMethod: "openapi_swagger",
				Category:        model.CatAPIEndpoint,
				Confidence:      model.ConfHigh,
				TechniqueID:     48,
			})
		}
	}

	// Detect swagger.json / openapi.json references
	specFileRe := model.GetRegex(`["']((?:https?://)?[^"'\s]*(?:swagger|openapi)\.(?:json|yaml|yml))["']`)
	for _, m := range specFileRe.FindAllStringSubmatchIndex(content, -1) {
		if len(m) >= 4 {
			url := content[m[2]:m[3]]
			results = append(results, model.Result{
				URL:             url,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(content, m[0]),
				DetectionMethod: "openapi_swagger",
				Category:        model.CatAPIEndpoint,
				Confidence:      model.ConfHigh,
				TechniqueID:     48,
			})
		}
	}

	return results
}

// extractFeatureFlags detects feature flag and A/B testing URL patterns.
func extractFeatureFlags(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	flagKeys := []string{
		"experiment_url", "variant_url", "flag_endpoint",
		"feature_url", "ab_test_url", "toggle_endpoint",
	}

	for _, key := range flagKeys {
		// Match key: "value" or "key": "value"
		re := model.GetRegex(`(?:["']?` + escapeRegexLiteral(key) + `["']?\s*[:=]\s*["'])([^"']+)["']`)
		matches := re.FindAllStringSubmatchIndex(content, -1)
		for _, m := range matches {
			if len(m) >= 4 {
				value := content[m[2]:m[3]]
				results = append(results, model.Result{
					URL:             value,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(content, m[0]),
					DetectionMethod: "feature_flags",
					Category:        model.CategorizeURL(value),
					Confidence:      model.ConfLow,
					TechniqueID:     49,
				})
			}
		}
	}

	return results
}
