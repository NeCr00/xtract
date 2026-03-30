package parser

import (
	"strings"

	"github.com/pentester/xtract/internal/model"
)

// tagURLAttributes returns the attributes that carry URLs for a given HTML tag.
func tagURLAttributes(tag string) []string {
	switch tag {
	case "a":
		return []string{"href"}
	case "link":
		return []string{"href"}
	case "script":
		return []string{"src"}
	case "img":
		return []string{"src"}
	case "iframe":
		return []string{"src"}
	case "source":
		return []string{"src"}
	case "embed":
		return []string{"src"}
	case "object":
		return []string{"data"}
	case "area":
		return []string{"href"}
	case "base":
		return []string{"href"}
	case "form":
		return []string{"action"}
	case "video":
		return []string{"src", "poster"}
	case "audio":
		return []string{"src"}
	case "track":
		return []string{"src"}
	case "input":
		return []string{"src"}
	default:
		return nil
	}
}

// extractMetaRefreshURL extracts the URL from a meta refresh content attribute.
// Format: "N;url=..." or "N; URL=..."
func extractMetaRefreshURL(content string) string {
	lower := strings.ToLower(content)
	idx := strings.Index(lower, "url=")
	if idx < 0 {
		return ""
	}
	u := strings.TrimSpace(content[idx+4:])
	u = strings.Trim(u, "'\"")
	return u
}

// parseSrcset parses an srcset attribute value and returns the URLs.
// Format: "url1 1x, url2 2x" or "url1 100w, url2 200w"
func parseSrcset(srcset string) []string {
	var urls []string
	candidates := strings.Split(srcset, ",")
	for _, c := range candidates {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		parts := strings.Fields(c)
		if len(parts) >= 1 {
			u := strings.TrimSpace(parts[0])
			if u != "" {
				urls = append(urls, u)
			}
		}
	}
	return urls
}

// extractURLsFromJSSnippet extracts URLs from a short JavaScript snippet
// (such as an inline event handler).
func extractURLsFromJSSnippet(js string) []string {
	var urls []string
	// Look for string literals containing URL-like values.
	re := model.GetRegex(`["']([^"']*(?:https?://|/)[^"']*)["']`)
	matches := re.FindAllStringSubmatch(js, -1)
	for _, m := range matches {
		if len(m) >= 2 {
			u := strings.TrimSpace(m[1])
			if isValidExtractedURL(u) {
				urls = append(urls, u)
			}
		}
	}
	return urls
}

// isValidExtractedURL checks if a string is a valid extracted URL
// (not empty, not just a hash, not just a protocol).
func isValidExtractedURL(u string) bool {
	if u == "" || u == "#" || u == "?" || u == "javascript:void(0)" || u == "javascript:;" {
		return false
	}
	if u == "about:blank" || u == "about:srcdoc" {
		return false
	}
	return true
}

// looksLikeURL checks if a string looks like it might be a URL.
func looksLikeURL(s string) bool {
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "//") {
		return true
	}
	if strings.HasPrefix(s, "/") && !strings.HasPrefix(s, "//") && len(s) > 1 {
		return true
	}
	return false
}

// looksLikeURLForJSON checks if a JSON string value looks like a URL.
func looksLikeURLForJSON(s string) bool {
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "//") {
		return true
	}
	if strings.HasPrefix(s, "/") && len(s) > 1 && !strings.ContainsAny(s[:2], " \t\n") {
		// Looks like an absolute path, but filter out things that are clearly not URLs.
		if strings.Contains(s, " ") {
			return false
		}
		return true
	}
	return false
}

// extractURLsFromRawJSON is a fallback regex-based extraction for malformed JSON.
func extractURLsFromRawJSON(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	re := model.GetRegex(`"((?:https?://|/)[^"]+)"`)
	matches := re.FindAllStringSubmatchIndex(content, -1)
	for _, m := range matches {
		if len(m) >= 4 {
			u := content[m[2]:m[3]]
			if isValidExtractedURL(u) {
				results = append(results, model.Result{
					URL:             u,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(content, m[0]),
					DetectionMethod: "json_regex_fallback",
					Category:        model.CategorizeURL(u),
					Confidence:      model.ConfLow,
					TechniqueID:     0,
				})
			}
		}
	}
	return results
}

// extractManifestURLs extracts URLs from web manifest JSON fields.
func extractManifestURLs(ctx *model.ExtractionContext, m map[string]interface{}) []model.Result {
	var results []model.Result

	stringFields := []string{"start_url", "scope", "sourceRoot", "file"}
	for _, field := range stringFields {
		if v, ok := m[field]; ok {
			if s, ok := v.(string); ok && s != "" {
				results = append(results, model.Result{
					URL:             s,
					SourceFile:      ctx.FileName,
					SourceLine:      1,
					DetectionMethod: "manifest_" + field,
					Category:        model.CatPageRoute,
					Confidence:      model.ConfHigh,
					TechniqueID:     46,
				})
			}
		}
	}

	// icons[].src
	if icons, ok := m["icons"]; ok {
		if iconList, ok := icons.([]interface{}); ok {
			for _, icon := range iconList {
				if iconMap, ok := icon.(map[string]interface{}); ok {
					if src, ok := iconMap["src"]; ok {
						if s, ok := src.(string); ok && s != "" {
							results = append(results, model.Result{
								URL:             s,
								SourceFile:      ctx.FileName,
								SourceLine:      1,
								DetectionMethod: "manifest_icon",
								Category:        model.CatStaticAsset,
								Confidence:      model.ConfHigh,
								TechniqueID:     46,
							})
						}
					}
				}
			}
		}
	}

	return results
}
