package parser

import (
	"strings"

	"github.com/Necr00/xtract/internal/model"
)

// ParseCSS extracts URLs from CSS content including url(), @import, and @font-face.
func ParseCSS(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	// Match url(...) - handles quoted and unquoted, single and double quotes.
	urlRe := model.GetRegex(`url\(\s*["']?([^"')]+?)["']?\s*\)`)
	matches := urlRe.FindAllStringSubmatchIndex(content, -1)
	for _, m := range matches {
		if len(m) >= 4 {
			u := strings.TrimSpace(content[m[2]:m[3]])
			if isValidExtractedURL(u) && !strings.HasPrefix(u, "data:") {
				results = append(results, model.Result{
					URL:             u,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(content, m[0]) + ctx.SourceLine,
					DetectionMethod: "css_url",
					Category:        model.CategorizeURL(u),
					Confidence:      model.ConfHigh,
					TechniqueID:     0,
				})
			}
		}
	}

	// Match @import "..." and @import '...' (without url())
	importRe := model.GetRegex(`@import\s+["']([^"']+)["']`)
	matches = importRe.FindAllStringSubmatchIndex(content, -1)
	for _, m := range matches {
		if len(m) >= 4 {
			u := strings.TrimSpace(content[m[2]:m[3]])
			if isValidExtractedURL(u) {
				results = append(results, model.Result{
					URL:             u,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(content, m[0]) + ctx.SourceLine,
					DetectionMethod: "css_import",
					Category:        model.CategorizeURL(u),
					Confidence:      model.ConfHigh,
					TechniqueID:     0,
				})
			}
		}
	}

	return results
}
