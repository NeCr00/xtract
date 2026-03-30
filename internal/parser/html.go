package parser

import (
	"strings"

	"github.com/pentester/xtract/internal/model"
	"golang.org/x/net/html"
)

// ParseHTML parses HTML content and extracts URLs from tags, attributes,
// inline event handlers, and meta refresh directives.
func ParseHTML(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	tokenizer := html.NewTokenizer(strings.NewReader(content))

	// Track byte position by counting consumed bytes.
	var consumed int

	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}

		raw := tokenizer.Raw()
		// Find offset of this token in the original content for line tracking.
		idx := strings.Index(content[consumed:], string(raw))
		var tokenOffset int
		if idx >= 0 {
			tokenOffset = consumed + idx
			consumed = tokenOffset + len(raw)
		} else {
			tokenOffset = consumed
		}
		srcLine := model.LineNumber(content, tokenOffset) + ctx.SourceLine

		switch tt {
		case html.StartTagToken, html.SelfClosingTagToken:
			tn, hasAttr := tokenizer.TagName()
			tagName := string(tn)

			if !hasAttr {
				continue
			}

			attrs := make(map[string]string)
			for {
				key, val, more := tokenizer.TagAttr()
				k := string(key)
				v := string(val)
				attrs[k] = v
				if !more {
					break
				}
			}

			// Standard URL-bearing attributes by tag.
			urlAttrs := tagURLAttributes(tagName)
			for _, attr := range urlAttrs {
				if v, ok := attrs[attr]; ok && v != "" {
					u := strings.TrimSpace(v)
					if isValidExtractedURL(u) {
						results = append(results, model.Result{
							URL:             u,
							SourceFile:      ctx.FileName,
							SourceLine:      srcLine,
							DetectionMethod: "html_" + tagName + "_" + attr,
							Category:        model.CategorizeURL(u),
							Confidence:      model.ConfHigh,
							TechniqueID:     0,
						})
					}
				}
			}

			// Handle <meta http-equiv="refresh" content="...;url=...">
			if tagName == "meta" {
				if strings.EqualFold(attrs["http-equiv"], "refresh") {
					if ct, ok := attrs["content"]; ok {
						u := extractMetaRefreshURL(ct)
						if u != "" {
							results = append(results, model.Result{
								URL:             u,
								SourceFile:      ctx.FileName,
								SourceLine:      srcLine,
								DetectionMethod: "html_meta_refresh",
								Category:        model.CategorizeURL(u),
								Confidence:      model.ConfHigh,
								TechniqueID:     0,
							})
						}
					}
				}
			}

			// Handle srcset attribute (can appear on img, source, etc.)
			if srcset, ok := attrs["srcset"]; ok && srcset != "" {
				urls := parseSrcset(srcset)
				for _, u := range urls {
					if isValidExtractedURL(u) {
						results = append(results, model.Result{
							URL:             u,
							SourceFile:      ctx.FileName,
							SourceLine:      srcLine,
							DetectionMethod: "html_srcset",
							Category:        model.CategorizeURL(u),
							Confidence:      model.ConfHigh,
							TechniqueID:     0,
						})
					}
				}
			}

			// Handle data-* attributes containing URLs
			for k, v := range attrs {
				if strings.HasPrefix(k, "data-") && v != "" {
					u := strings.TrimSpace(v)
					if looksLikeURL(u) {
						results = append(results, model.Result{
							URL:             u,
							SourceFile:      ctx.FileName,
							SourceLine:      srcLine,
							DetectionMethod: "html_data_attr",
							Category:        model.CategorizeURL(u),
							Confidence:      model.ConfMedium,
							TechniqueID:     0,
						})
					}
				}
			}

			// Handle inline event handlers (onclick, onload, etc.)
			eventAttrs := []string{
				"onclick", "onload", "onerror", "onsubmit", "onchange",
				"onmouseover", "onmouseout", "onfocus", "onblur", "onkeyup",
				"onkeydown", "onkeypress", "onresize", "onscroll",
				"ondblclick", "oncontextmenu", "oninput", "onreset",
			}
			for _, ea := range eventAttrs {
				if v, ok := attrs[ea]; ok && v != "" {
					urls := extractURLsFromJSSnippet(v)
					for _, u := range urls {
						results = append(results, model.Result{
							URL:             u,
							SourceFile:      ctx.FileName,
							SourceLine:      srcLine,
							DetectionMethod: "html_event_handler",
							Category:        model.CategorizeURL(u),
							Confidence:      model.ConfMedium,
							TechniqueID:     0,
						})
					}
				}
			}
		}
	}

	return results
}

// ExtractScriptContents extracts all inline <script> tag contents from HTML
// as separate ExtractionContext objects, each with FileType "js".
func ExtractScriptContents(ctx *model.ExtractionContext) []model.ExtractionContext {
	var contexts []model.ExtractionContext
	content := ctx.Content

	tokenizer := html.NewTokenizer(strings.NewReader(content))
	var consumed int
	inScript := false
	var scriptStart int

	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}

		raw := tokenizer.Raw()
		idx := strings.Index(content[consumed:], string(raw))
		var tokenOffset int
		if idx >= 0 {
			tokenOffset = consumed + idx
			consumed = tokenOffset + len(raw)
		} else {
			tokenOffset = consumed
		}

		switch tt {
		case html.StartTagToken:
			tn, hasAttr := tokenizer.TagName()
			if string(tn) == "script" {
				// Check if it has a src attribute; if so, skip its content.
				hasSrc := false
				if hasAttr {
					for {
						key, _, more := tokenizer.TagAttr()
						if string(key) == "src" {
							hasSrc = true
						}
						if !more {
							break
						}
					}
				}
				if !hasSrc {
					inScript = true
					scriptStart = consumed
				}
			}

		case html.TextToken:
			if inScript {
				text := string(tokenizer.Text())
				trimmed := strings.TrimSpace(text)
				if trimmed != "" {
					srcLine := model.LineNumber(content, scriptStart) + ctx.SourceLine
					contexts = append(contexts, model.ExtractionContext{
						Content:    text,
						FileName:   ctx.FileName,
						FileType:   "js",
						BaseURL:    ctx.BaseURL,
						SourceLine: srcLine,
					})
				}
			}

		case html.EndTagToken:
			tn, _ := tokenizer.TagName()
			if string(tn) == "script" {
				inScript = false
			}
		}
	}

	return contexts
}
