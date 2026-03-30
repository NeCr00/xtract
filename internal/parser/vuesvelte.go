package parser

import (
	"github.com/Necr00/xtract/internal/model"
)

// ParseVueSvelte parses Vue SFC (.vue) and Svelte (.svelte) files by extracting
// <template>, <script>, and <style> sections and processing each appropriately.
func ParseVueSvelte(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	// Extract <template> section and process as HTML.
	templateRe := model.GetRegex(`(?is)<template[^>]*>(.*?)</template>`)
	if m := templateRe.FindStringSubmatchIndex(content); m != nil && len(m) >= 4 {
		templateContent := content[m[2]:m[3]]
		templateLine := model.LineNumber(content, m[2]) + ctx.SourceLine
		subCtx := &model.ExtractionContext{
			Content:    templateContent,
			FileName:   ctx.FileName,
			FileType:   "html",
			BaseURL:    ctx.BaseURL,
			SourceLine: templateLine,
		}
		results = append(results, ParseHTML(subCtx)...)
	}

	// Extract <script> sections (may have multiple: <script> and <script setup>).
	scriptRe := model.GetRegex(`(?is)<script[^>]*>(.*?)</script>`)
	for _, m := range scriptRe.FindAllStringSubmatchIndex(content, -1) {
		if len(m) >= 4 {
			scriptContent := content[m[2]:m[3]]
			scriptLine := model.LineNumber(content, m[2]) + ctx.SourceLine
			// Extract URLs from script content using string literal matching.
			urlRe := model.GetRegex(`["']((?:https?://|/)[^"'\s]{2,})["']`)
			for _, um := range urlRe.FindAllStringSubmatchIndex(scriptContent, -1) {
				if len(um) >= 4 {
					u := scriptContent[um[2]:um[3]]
					if isValidExtractedURL(u) {
						results = append(results, model.Result{
							URL:             u,
							SourceFile:      ctx.FileName,
							SourceLine:      model.LineNumber(scriptContent, um[0]) + scriptLine,
							DetectionMethod: "vue_svelte_script",
							Category:        model.CategorizeURL(u),
							Confidence:      model.ConfHigh,
							TechniqueID:     0,
						})
					}
				}
			}
		}
	}

	// Extract <style> sections and process as CSS.
	styleRe := model.GetRegex(`(?is)<style[^>]*>(.*?)</style>`)
	for _, m := range styleRe.FindAllStringSubmatchIndex(content, -1) {
		if len(m) >= 4 {
			styleContent := content[m[2]:m[3]]
			styleLine := model.LineNumber(content, m[2]) + ctx.SourceLine
			subCtx := &model.ExtractionContext{
				Content:    styleContent,
				FileName:   ctx.FileName,
				FileType:   "css",
				BaseURL:    ctx.BaseURL,
				SourceLine: styleLine,
			}
			results = append(results, ParseCSS(subCtx)...)
		}
	}

	return results
}
