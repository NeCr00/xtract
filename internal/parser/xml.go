package parser

import (
	"github.com/pentester/xtract/internal/model"
)

// ParseXML extracts URLs from XML/SVG content using regex-based matching.
func ParseXML(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	// href="..." attributes
	hrefRe := model.GetRegex(`(?i)\bhref\s*=\s*["']([^"']+)["']`)
	for _, m := range hrefRe.FindAllStringSubmatchIndex(content, -1) {
		if len(m) >= 4 {
			u := content[m[2]:m[3]]
			if isValidExtractedURL(u) {
				results = append(results, model.Result{
					URL:             u,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(content, m[0]) + ctx.SourceLine,
					DetectionMethod: "xml_href",
					Category:        model.CategorizeURL(u),
					Confidence:      model.ConfHigh,
					TechniqueID:     0,
				})
			}
		}
	}

	// xlink:href="..." attributes
	xlinkRe := model.GetRegex(`(?i)\bxlink:href\s*=\s*["']([^"']+)["']`)
	for _, m := range xlinkRe.FindAllStringSubmatchIndex(content, -1) {
		if len(m) >= 4 {
			u := content[m[2]:m[3]]
			if isValidExtractedURL(u) {
				results = append(results, model.Result{
					URL:             u,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(content, m[0]) + ctx.SourceLine,
					DetectionMethod: "xml_xlink_href",
					Category:        model.CategorizeURL(u),
					Confidence:      model.ConfHigh,
					TechniqueID:     0,
				})
			}
		}
	}

	// src="..." attributes
	srcRe := model.GetRegex(`(?i)\bsrc\s*=\s*["']([^"']+)["']`)
	for _, m := range srcRe.FindAllStringSubmatchIndex(content, -1) {
		if len(m) >= 4 {
			u := content[m[2]:m[3]]
			if isValidExtractedURL(u) {
				results = append(results, model.Result{
					URL:             u,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(content, m[0]) + ctx.SourceLine,
					DetectionMethod: "xml_src",
					Category:        model.CategorizeURL(u),
					Confidence:      model.ConfHigh,
					TechniqueID:     0,
				})
			}
		}
	}

	// action="..." attributes (for XForms etc.)
	actionRe := model.GetRegex(`(?i)\baction\s*=\s*["']([^"']+)["']`)
	for _, m := range actionRe.FindAllStringSubmatchIndex(content, -1) {
		if len(m) >= 4 {
			u := content[m[2]:m[3]]
			if isValidExtractedURL(u) {
				results = append(results, model.Result{
					URL:             u,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(content, m[0]) + ctx.SourceLine,
					DetectionMethod: "xml_action",
					Category:        model.CategorizeURL(u),
					Confidence:      model.ConfMedium,
					TechniqueID:     0,
				})
			}
		}
	}

	// Namespace URIs: xmlns="..." and xmlns:prefix="..."
	nsRe := model.GetRegex(`(?i)\bxmlns(?::\w+)?\s*=\s*["'](https?://[^"']+)["']`)
	for _, m := range nsRe.FindAllStringSubmatchIndex(content, -1) {
		if len(m) >= 4 {
			u := content[m[2]:m[3]]
			results = append(results, model.Result{
				URL:             u,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(content, m[0]) + ctx.SourceLine,
				DetectionMethod: "xml_namespace",
				Category:        model.CatExternalSvc,
				Confidence:      model.ConfLow,
				TechniqueID:     0,
			})
		}
	}

	// Inline <script> content within XML/SVG - extract URLs from CDATA or text.
	scriptRe := model.GetRegex(`(?is)<script[^>]*>(.*?)</script>`)
	for _, m := range scriptRe.FindAllStringSubmatchIndex(content, -1) {
		if len(m) >= 4 {
			scriptContent := content[m[2]:m[3]]
			scriptLine := model.LineNumber(content, m[2]) + ctx.SourceLine
			// Extract URLs from embedded script using JS-style string matching.
			jsURLs := extractURLsFromJSSnippet(scriptContent)
			for _, u := range jsURLs {
				results = append(results, model.Result{
					URL:             u,
					SourceFile:      ctx.FileName,
					SourceLine:      scriptLine,
					DetectionMethod: "xml_embedded_script",
					Category:        model.CategorizeURL(u),
					Confidence:      model.ConfMedium,
					TechniqueID:     0,
				})
			}
		}
	}

	return results
}
