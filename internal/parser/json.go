package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Necr00/xtract/internal/model"
)

// ParseJSON recursively walks JSON content and extracts URLs from string values.
func ParseJSON(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	var data interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		// If JSON is invalid, fall back to regex extraction.
		return extractURLsFromRawJSON(ctx)
	}

	// Detect if this is a source map.
	if m, ok := data.(map[string]interface{}); ok {
		if _, hasSources := m["sources"]; hasSources {
			if _, hasMapping := m["mappings"]; hasMapping {
				return ParseSourceMap(ctx)
			}
		}
	}

	// Detect if this is a web manifest.
	if m, ok := data.(map[string]interface{}); ok {
		if _, hasStartURL := m["start_url"]; hasStartURL {
			results = append(results, extractManifestURLs(ctx, m)...)
		}
	}

	// URL-bearing key names.
	urlKeys := map[string]bool{
		"url": true, "href": true, "src": true, "endpoint": true,
		"path": true, "link": true, "uri": true, "redirect": true,
		"callback": true, "webhook": true, "origin": true, "host": true,
		"action": true, "target": true, "location": true, "base_url": true,
		"baseurl": true, "baseURL": true, "apiUrl": true, "api_url": true,
	}

	var walk func(v interface{}, keyHint string)
	walk = func(v interface{}, keyHint string) {
		switch val := v.(type) {
		case map[string]interface{}:
			for k, child := range val {
				walk(child, strings.ToLower(k))
			}
		case []interface{}:
			for _, child := range val {
				walk(child, keyHint)
			}
		case string:
			isURLKey := urlKeys[keyHint]
			if isURLKey || looksLikeURLForJSON(val) {
				u := strings.TrimSpace(val)
				if isValidExtractedURL(u) && len(u) > 1 {
					method := "json_value"
					conf := model.ConfMedium
					if isURLKey {
						method = "json_url_key"
						conf = model.ConfHigh
					}
					results = append(results, model.Result{
						URL:             u,
						SourceFile:      ctx.FileName,
						SourceLine:      1,
						DetectionMethod: method,
						Category:        model.CategorizeURL(u),
						Confidence:      conf,
						TechniqueID:     0,
					})
				}
			}
		}
	}

	walk(data, "")
	return results
}

// ParseSourceMap parses source map JSON and extracts source file paths,
// the file field, and sourceRoot.
func ParseSourceMap(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return results
	}

	// Extract "file" field.
	if file, ok := data["file"]; ok {
		if s, ok := file.(string); ok && s != "" {
			results = append(results, model.Result{
				URL:             s,
				SourceFile:      ctx.FileName,
				SourceLine:      1,
				DetectionMethod: "sourcemap_file",
				Category:        model.CatSourceMap,
				Confidence:      model.ConfHigh,
				TechniqueID:     44,
			})
		}
	}

	// Extract "sourceRoot" field.
	sourceRoot := ""
	if sr, ok := data["sourceRoot"]; ok {
		if s, ok := sr.(string); ok && s != "" {
			sourceRoot = s
			results = append(results, model.Result{
				URL:             s,
				SourceFile:      ctx.FileName,
				SourceLine:      1,
				DetectionMethod: "sourcemap_sourceRoot",
				Category:        model.CatSourceMap,
				Confidence:      model.ConfHigh,
				TechniqueID:     44,
			})
		}
	}

	// Extract "sources" array.
	if sources, ok := data["sources"]; ok {
		if sourceList, ok := sources.([]interface{}); ok {
			for _, src := range sourceList {
				if s, ok := src.(string); ok && s != "" {
					fullPath := s
					if sourceRoot != "" && !strings.HasPrefix(s, "/") &&
						!strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
						fullPath = strings.TrimRight(sourceRoot, "/") + "/" + s
					}
					results = append(results, model.Result{
						URL:             fullPath,
						SourceFile:      ctx.FileName,
						SourceLine:      1,
						DetectionMethod: "sourcemap_source",
						Category:        model.CatSourceMap,
						Confidence:      model.ConfHigh,
						TechniqueID:     44,
					})
				}
			}
		}
	}

	// Scan "sourcesContent" entries for URLs.
	if sourcesContent, ok := data["sourcesContent"]; ok {
		if contentList, ok := sourcesContent.([]interface{}); ok {
			// Determine source names for labeling.
			var sourceNames []string
			if sources, ok := data["sources"]; ok {
				if sourceList, ok := sources.([]interface{}); ok {
					for _, src := range sourceList {
						if s, ok := src.(string); ok {
							sourceNames = append(sourceNames, s)
						} else {
							sourceNames = append(sourceNames, "")
						}
					}
				}
			}
			for i, entry := range contentList {
				if s, ok := entry.(string); ok && s != "" {
					sourceName := fmt.Sprintf("%s:[sourcesContent:%d]", ctx.FileName, i)
					if i < len(sourceNames) && sourceNames[i] != "" {
						sourceName = sourceNames[i]
					}
					subCtx := &model.ExtractionContext{
						Content:  s,
						FileName: sourceName,
						FileType: "js",
						BaseURL:  ctx.BaseURL,
					}
					// Extract URLs from source content using regex.
					urlRe := model.GetRegex(`["']((?:https?://|/)[^"'\s]{2,})["']`)
					for _, m := range urlRe.FindAllStringSubmatchIndex(subCtx.Content, -1) {
						if len(m) >= 4 {
							u := subCtx.Content[m[2]:m[3]]
							if isValidExtractedURL(u) {
								results = append(results, model.Result{
									URL:             u,
									SourceFile:      sourceName,
									SourceLine:      model.LineNumber(subCtx.Content, m[0]),
									DetectionMethod: "sourcemap_content",
									Category:        model.CategorizeURL(u),
									Confidence:      model.ConfMedium,
									TechniqueID:     44,
								})
							}
						}
					}
				}
			}
		}
	}

	return results
}
