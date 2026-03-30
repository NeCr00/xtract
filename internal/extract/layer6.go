package extract

import (
	"strings"

	"github.com/pentester/xtract/internal/model"
)

// ExtractLayer6 runs Layer 6: Comments & Developer Artifacts (techniques 54-59).
func ExtractLayer6(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	results = append(results, extractSingleLineComments(ctx)...)
	results = append(results, extractMultiLineComments(ctx)...)
	results = append(results, extractHTMLComments(ctx)...)
	results = append(results, extractJSDocTags(ctx)...)
	results = append(results, extractDevAnnotations(ctx)...)
	results = append(results, extractCommentedCode(ctx)...)
	return results
}

// extractURLsFromText finds all URL-like and path-like patterns in free text.
func extractURLsFromText(text string) []string {
	var urls []string
	seen := make(map[string]bool)

	// Absolute URLs (http, https, ftp, ws, wss)
	absRe := model.GetRegex(`(?i)(?:https?|ftp|wss?)://[^\s"'<>\)` + "`" + `\]\}]+`)
	for _, m := range absRe.FindAllString(text, -1) {
		// Trim trailing punctuation that is likely not part of the URL
		m = strings.TrimRight(m, ".,;:!?)")
		if !seen[m] {
			seen[m] = true
			urls = append(urls, m)
		}
	}

	// Protocol-relative URLs
	protoRelRe := model.GetRegex(`//[a-zA-Z0-9][a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}[^\s"'<>\)` + "`" + `]*`)
	for _, m := range protoRelRe.FindAllString(text, -1) {
		m = strings.TrimRight(m, ".,;:!?)")
		if !seen[m] {
			seen[m] = true
			urls = append(urls, m)
		}
	}

	// Relative paths starting with /
	relRe := model.GetRegex(`(?:^|[\s"'=\(])(/[a-zA-Z0-9_\-][a-zA-Z0-9_\-./]*[a-zA-Z0-9_\-])`)
	for _, match := range relRe.FindAllStringSubmatch(text, -1) {
		if len(match) > 1 {
			p := match[1]
			// Skip overly short or file-system-like paths
			if len(p) > 2 && !seen[p] {
				seen[p] = true
				urls = append(urls, p)
			}
		}
	}

	// Domain names (at least subdomain.domain.tld)
	domainRe := model.GetRegex(`(?i)[a-z0-9][a-z0-9\-]*\.[a-z0-9][a-z0-9\-]*\.[a-z]{2,63}(?:/[^\s"'<>]*)?`)
	for _, m := range domainRe.FindAllString(text, -1) {
		m = strings.TrimRight(m, ".,;:!?)")
		if !seen[m] {
			seen[m] = true
			urls = append(urls, m)
		}
	}

	return urls
}

// technique 54: URLs in single-line comments (//)
func extractSingleLineComments(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	re := model.GetRegex(`(?m)(?:^|[^:])//(.*)$`)
	matches := re.FindAllStringSubmatchIndex(ctx.Content, -1)
	seen := make(map[string]bool)
	for _, loc := range matches {
		// loc[2]:loc[3] is the capture group (comment text after //)
		commentText := ctx.Content[loc[2]:loc[3]]
		commentOffset := loc[0]
		urls := extractURLsFromText(commentText)
		for _, u := range urls {
			if seen[u] {
				continue
			}
			seen[u] = true
			results = append(results, model.Result{
				URL:             u,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(ctx.Content, commentOffset),
				DetectionMethod: "single_line_comments",
				Category:        model.CategorizeURL(u),
				Confidence:      model.ConfLow,
				TechniqueID:     54,
			})
		}
	}
	return results
}

// technique 55: URLs in multi-line comments (/* ... */)
func extractMultiLineComments(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	re := model.GetRegex(`(?s)/\*(.+?)\*/`)
	matches := re.FindAllStringSubmatchIndex(ctx.Content, -1)
	seen := make(map[string]bool)
	for _, loc := range matches {
		commentText := ctx.Content[loc[2]:loc[3]]
		commentOffset := loc[0]
		urls := extractURLsFromText(commentText)
		for _, u := range urls {
			if seen[u] {
				continue
			}
			seen[u] = true
			results = append(results, model.Result{
				URL:             u,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(ctx.Content, commentOffset),
				DetectionMethod: "multi_line_comments",
				Category:        model.CategorizeURL(u),
				Confidence:      model.ConfLow,
				TechniqueID:     55,
			})
		}
	}
	return results
}

// technique 56: URLs in HTML comments (<!-- ... -->)
func extractHTMLComments(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	re := model.GetRegex(`(?s)<!--(.+?)-->`)
	matches := re.FindAllStringSubmatchIndex(ctx.Content, -1)
	seen := make(map[string]bool)
	for _, loc := range matches {
		commentText := ctx.Content[loc[2]:loc[3]]
		commentOffset := loc[0]
		urls := extractURLsFromText(commentText)
		for _, u := range urls {
			if seen[u] {
				continue
			}
			seen[u] = true
			results = append(results, model.Result{
				URL:             u,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(ctx.Content, commentOffset),
				DetectionMethod: "html_comments",
				Category:        model.CategorizeURL(u),
				Confidence:      model.ConfLow,
				TechniqueID:     56,
			})
		}
	}
	return results
}

// technique 57: JSDoc @link and @see URL references
func extractJSDocTags(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	seen := make(map[string]bool)

	// {@link url} inline tag
	inlineLinkRe := model.GetRegex(`\{@link\s+([^\s\}]+)\}`)
	for _, loc := range inlineLinkRe.FindAllStringSubmatchIndex(ctx.Content, -1) {
		url := ctx.Content[loc[2]:loc[3]]
		if !seen[url] {
			seen[url] = true
			results = append(results, model.Result{
				URL:             url,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(ctx.Content, loc[0]),
				DetectionMethod: "jsdoc_tags",
				Category:        model.CategorizeURL(url),
				Confidence:      model.ConfLow,
				TechniqueID:     57,
			})
		}
	}

	// @link url (block-level, inside comment)
	blockLinkRe := model.GetRegex(`@link\s+(https?://[^\s\}]+)`)
	for _, loc := range blockLinkRe.FindAllStringSubmatchIndex(ctx.Content, -1) {
		url := ctx.Content[loc[2]:loc[3]]
		if !seen[url] {
			seen[url] = true
			results = append(results, model.Result{
				URL:             url,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(ctx.Content, loc[0]),
				DetectionMethod: "jsdoc_tags",
				Category:        model.CategorizeURL(url),
				Confidence:      model.ConfLow,
				TechniqueID:     57,
			})
		}
	}

	// @see url
	seeRe := model.GetRegex(`@see\s+(https?://[^\s]+)`)
	for _, loc := range seeRe.FindAllStringSubmatchIndex(ctx.Content, -1) {
		url := ctx.Content[loc[2]:loc[3]]
		if !seen[url] {
			seen[url] = true
			results = append(results, model.Result{
				URL:             url,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(ctx.Content, loc[0]),
				DetectionMethod: "jsdoc_tags",
				Category:        model.CategorizeURL(url),
				Confidence:      model.ConfLow,
				TechniqueID:     57,
			})
		}
	}

	// @example lines containing URLs
	exampleRe := model.GetRegex(`(?m)@example\s+(.*)$`)
	for _, loc := range exampleRe.FindAllStringSubmatchIndex(ctx.Content, -1) {
		line := ctx.Content[loc[2]:loc[3]]
		for _, u := range extractURLsFromText(line) {
			if !seen[u] {
				seen[u] = true
				results = append(results, model.Result{
					URL:             u,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(ctx.Content, loc[0]),
					DetectionMethod: "jsdoc_tags",
					Category:        model.CategorizeURL(u),
					Confidence:      model.ConfLow,
					TechniqueID:     57,
				})
			}
		}
	}

	return results
}

// technique 58: developer annotations (// API: url, // endpoint: url, etc.)
func extractDevAnnotations(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	re := model.GetRegex(`(?im)//\s*(?:API|endpoint|url|route|URL|server|host|base|TODO)[:\s]+(.+)$`)
	seen := make(map[string]bool)
	for _, loc := range re.FindAllStringSubmatchIndex(ctx.Content, -1) {
		annotationText := ctx.Content[loc[2]:loc[3]]
		urls := extractURLsFromText(annotationText)
		for _, u := range urls {
			if !seen[u] {
				seen[u] = true
				results = append(results, model.Result{
					URL:             u,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(ctx.Content, loc[0]),
					DetectionMethod: "dev_annotations",
					Category:        model.CategorizeURL(u),
					Confidence:      model.ConfLow,
					TechniqueID:     58,
				})
			}
		}
	}
	return results
}

// technique 59: URLs from commented-out code
func extractCommentedCode(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	// Match single-line comments that look like commented-out code
	// Heuristic: line contains =, (, ;, or function-call-like pattern
	re := model.GetRegex(`(?m)//\s*(.+)$`)
	codeIndicator := model.GetRegex(`[=\(;]`)
	seen := make(map[string]bool)

	for _, loc := range re.FindAllStringSubmatchIndex(ctx.Content, -1) {
		commentText := ctx.Content[loc[2]:loc[3]]
		// Only process if the comment looks like code
		if !codeIndicator.MatchString(commentText) {
			continue
		}
		urls := extractURLsFromText(commentText)
		for _, u := range urls {
			if !seen[u] {
				seen[u] = true
				results = append(results, model.Result{
					URL:             u,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(ctx.Content, loc[0]),
					DetectionMethod: "commented_code",
					Category:        model.CategorizeURL(u),
					Confidence:      model.ConfLow,
					TechniqueID:     59,
				})
			}
		}
	}

	// Also check multi-line commented-out code blocks
	mlRe := model.GetRegex(`(?s)/\*(.+?)\*/`)
	for _, loc := range mlRe.FindAllStringSubmatchIndex(ctx.Content, -1) {
		blockText := ctx.Content[loc[2]:loc[3]]
		lines := strings.Split(blockText, "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			// Remove leading * from JSDoc-style comments
			trimmed = strings.TrimLeft(trimmed, "* ")
			if len(trimmed) == 0 {
				continue
			}
			if !codeIndicator.MatchString(trimmed) {
				continue
			}
			for _, u := range extractURLsFromText(trimmed) {
				if !seen[u] {
					seen[u] = true
					results = append(results, model.Result{
						URL:             u,
						SourceFile:      ctx.FileName,
						SourceLine:      model.LineNumber(ctx.Content, loc[0]),
						DetectionMethod: "commented_code",
						Category:        model.CategorizeURL(u),
						Confidence:      model.ConfLow,
						TechniqueID:     59,
					})
				}
			}
		}
	}

	return results
}
