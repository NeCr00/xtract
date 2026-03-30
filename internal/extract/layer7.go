package extract

import (
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/Necr00/xtract/internal/model"
)

// ExtractLayer7 runs Layer 7: Encoded & Obfuscated Recovery (techniques 60-67).
func ExtractLayer7(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	results = append(results, extractBase64Decode(ctx)...)
	results = append(results, extractHexDecode(ctx)...)
	results = append(results, extractUnicodeDecode(ctx)...)
	results = append(results, extractArrayObfuscation(ctx)...)
	results = append(results, extractFromCharCode(ctx)...)
	results = append(results, extractAtobCalls(ctx)...)
	results = append(results, extractReverseSplitJoin(ctx)...)
	results = append(results, extractURLConstructor(ctx)...)
	return results
}

// decodeBase64 attempts to decode a base64 string, returning the decoded string and success flag.
func decodeBase64(s string) (string, bool) {
	// Try standard encoding first
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err == nil && isPrintableASCII(decoded) {
		return string(decoded), true
	}
	// Try URL-safe encoding
	decoded, err = base64.URLEncoding.DecodeString(s)
	if err == nil && isPrintableASCII(decoded) {
		return string(decoded), true
	}
	// Try without padding (RawStdEncoding)
	decoded, err = base64.RawStdEncoding.DecodeString(s)
	if err == nil && isPrintableASCII(decoded) {
		return string(decoded), true
	}
	// Try URL-safe without padding
	decoded, err = base64.RawURLEncoding.DecodeString(s)
	if err == nil && isPrintableASCII(decoded) {
		return string(decoded), true
	}
	return "", false
}

// isPrintableASCII checks whether the byte slice consists mostly of printable characters.
func isPrintableASCII(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	nonPrintable := 0
	for _, c := range b {
		if c < 0x20 && c != '\n' && c != '\r' && c != '\t' {
			nonPrintable++
		}
	}
	// Allow up to 10% non-printable characters
	return nonPrintable*10 <= len(b)
}

// decodeHexEscapeSeq decodes a string of \xNN hex escape sequences, returning
// the decoded string and a success flag.
func decodeHexEscapeSeq(s string) (string, bool) {
	re := model.GetRegex(`\\x([0-9a-fA-F]{2})`)
	matches := re.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return "", false
	}
	var buf []byte
	for _, m := range matches {
		val, err := strconv.ParseUint(m[1], 16, 8)
		if err != nil {
			return "", false
		}
		buf = append(buf, byte(val))
	}
	result := string(buf)
	if !isPrintableASCII([]byte(result)) {
		return "", false
	}
	return result, true
}

// decodeUnicodeEscapeSeq decodes a string of \uNNNN unicode escape sequences,
// returning the decoded string and a success flag.
func decodeUnicodeEscapeSeq(s string) (string, bool) {
	re := model.GetRegex(`\\u([0-9a-fA-F]{4})`)
	matches := re.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return "", false
	}
	var buf strings.Builder
	for _, m := range matches {
		val, err := strconv.ParseUint(m[1], 16, 32)
		if err != nil {
			return "", false
		}
		buf.WriteRune(rune(val))
	}
	result := buf.String()
	if !isPrintableASCII([]byte(result)) {
		return "", false
	}
	return result, true
}

// findURLsInDecoded extracts URL-like patterns from decoded strings.
func findURLsInDecoded(s string) []string {
	var urls []string
	seen := make(map[string]bool)

	// Absolute URLs
	absRe := model.GetRegex(`(?i)(?:https?|ftp|wss?)://[^\s"'<>` + "`" + `]+`)
	for _, m := range absRe.FindAllString(s, -1) {
		m = strings.TrimRight(m, ".,;:!?)")
		if !seen[m] {
			seen[m] = true
			urls = append(urls, m)
		}
	}

	// Relative paths
	relRe := model.GetRegex(`(?:^|[\s"'=])(/[a-zA-Z0-9_\-][a-zA-Z0-9_\-./]*[a-zA-Z0-9_\-])`)
	for _, match := range relRe.FindAllStringSubmatch(s, -1) {
		if len(match) > 1 {
			p := match[1]
			if len(p) > 2 && !seen[p] {
				seen[p] = true
				urls = append(urls, p)
			}
		}
	}

	return urls
}

// technique 60: Base64-encoded string detection and decoding
func extractBase64Decode(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	// Match base64 strings in string literals (min 20 chars, valid base64 charset)
	re := model.GetRegex(`["']([A-Za-z0-9+/=_\-]{20,})["']`)
	seen := make(map[string]bool)

	for _, loc := range re.FindAllStringSubmatchIndex(ctx.Content, -1) {
		encoded := ctx.Content[loc[2]:loc[3]]
		decoded, ok := decodeBase64(encoded)
		if !ok {
			continue
		}
		urls := findURLsInDecoded(decoded)
		for _, u := range urls {
			if !seen[u] {
				seen[u] = true
				results = append(results, model.Result{
					URL:             u,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(ctx.Content, loc[0]),
					DetectionMethod: "base64_decode",
					Category:        model.CategorizeURL(u),
					Confidence:      model.ConfLow,
					TechniqueID:     60,
				})
			}
		}
	}
	return results
}

// technique 61: Hex-encoded string detection and decoding
func extractHexDecode(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	// Match sequences of \xNN (at least 4 consecutive)
	re := model.GetRegex(`(?:\\x[0-9a-fA-F]{2}){4,}`)
	seen := make(map[string]bool)

	for _, loc := range re.FindAllStringIndex(ctx.Content, -1) {
		hexStr := ctx.Content[loc[0]:loc[1]]
		decoded, ok := decodeHexEscapeSeq(hexStr)
		if !ok {
			continue
		}
		urls := findURLsInDecoded(decoded)
		for _, u := range urls {
			if !seen[u] {
				seen[u] = true
				results = append(results, model.Result{
					URL:             u,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(ctx.Content, loc[0]),
					DetectionMethod: "hex_decode",
					Category:        model.CategorizeURL(u),
					Confidence:      model.ConfLow,
					TechniqueID:     61,
				})
			}
		}
	}
	return results
}

// technique 62: Unicode-escaped string decoding
func extractUnicodeDecode(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	// Match sequences of \uNNNN (at least 4 consecutive)
	re := model.GetRegex(`(?:\\u[0-9a-fA-F]{4}){4,}`)
	seen := make(map[string]bool)

	for _, loc := range re.FindAllStringIndex(ctx.Content, -1) {
		uniStr := ctx.Content[loc[0]:loc[1]]
		decoded, ok := decodeUnicodeEscapeSeq(uniStr)
		if !ok {
			continue
		}
		urls := findURLsInDecoded(decoded)
		for _, u := range urls {
			if !seen[u] {
				seen[u] = true
				results = append(results, model.Result{
					URL:             u,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(ctx.Content, loc[0]),
					DetectionMethod: "unicode_decode",
					Category:        model.CategorizeURL(u),
					Confidence:      model.ConfLow,
					TechniqueID:     62,
				})
			}
		}
	}
	return results
}

// technique 63: Array-based string obfuscation recovery
func extractArrayObfuscation(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	// Match var _0x... = ['...', '...'] or var a = ["...", "..."]
	re := model.GetRegex(`(?:var|let|const)\s+[a-zA-Z_$][a-zA-Z0-9_$]*\s*=\s*\[([^\]]{10,})\]`)
	seen := make(map[string]bool)

	for _, loc := range re.FindAllStringSubmatchIndex(ctx.Content, -1) {
		arrayContent := ctx.Content[loc[2]:loc[3]]
		// Extract individual string elements from the array
		strRe := model.GetRegex(`["']([^"']+)["']`)
		for _, strMatch := range strRe.FindAllStringSubmatch(arrayContent, -1) {
			if len(strMatch) < 2 {
				continue
			}
			element := strMatch[1]
			urls := findURLsInDecoded(element)
			for _, u := range urls {
				if !seen[u] {
					seen[u] = true
					results = append(results, model.Result{
						URL:             u,
						SourceFile:      ctx.FileName,
						SourceLine:      model.LineNumber(ctx.Content, loc[0]),
						DetectionMethod: "array_obfuscation",
						Category:        model.CategorizeURL(u),
						Confidence:      model.ConfLow,
						TechniqueID:     63,
					})
				}
			}
			// Also check if the element itself looks like a path
			if isURLLike(element) && !seen[element] {
				seen[element] = true
				results = append(results, model.Result{
					URL:             element,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(ctx.Content, loc[0]),
					DetectionMethod: "array_obfuscation",
					Category:        model.CategorizeURL(element),
					Confidence:      model.ConfLow,
					TechniqueID:     63,
				})
			}
		}
	}
	return results
}

// isURLLike checks whether a string looks like a URL or path.
func isURLLike(s string) bool {
	if len(s) < 2 {
		return false
	}
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") ||
		strings.HasPrefix(s, "ftp://") || strings.HasPrefix(s, "ws://") ||
		strings.HasPrefix(s, "wss://") || strings.HasPrefix(s, "//") {
		return true
	}
	if s[0] == '/' && len(s) > 1 && s[1] != '/' {
		return true
	}
	// Domain-like: contains at least one dot and looks like a hostname
	dotRe := model.GetRegex(`^[a-zA-Z0-9][a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}(?:/|$)`)
	return dotRe.MatchString(s)
}

// technique 64: String.fromCharCode() resolution
func extractFromCharCode(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	re := model.GetRegex(`String\.fromCharCode\(([0-9,\s]+)\)`)
	seen := make(map[string]bool)

	for _, loc := range re.FindAllStringSubmatchIndex(ctx.Content, -1) {
		argsStr := ctx.Content[loc[2]:loc[3]]
		parts := strings.Split(argsStr, ",")
		var buf strings.Builder
		valid := true
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			n, err := strconv.Atoi(p)
			if err != nil || n < 0 || n > 0x10FFFF {
				valid = false
				break
			}
			buf.WriteRune(rune(n))
		}
		if !valid {
			continue
		}
		decoded := buf.String()
		urls := findURLsInDecoded(decoded)
		for _, u := range urls {
			if !seen[u] {
				seen[u] = true
				results = append(results, model.Result{
					URL:             u,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(ctx.Content, loc[0]),
					DetectionMethod: "fromcharcode",
					Category:        model.CategorizeURL(u),
					Confidence:      model.ConfLow,
					TechniqueID:     64,
				})
			}
		}
	}
	return results
}

// technique 65: atob() Base64 decoding
func extractAtobCalls(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	re := model.GetRegex(`atob\(["']([A-Za-z0-9+/=_\-]+)["']\)`)
	seen := make(map[string]bool)

	for _, loc := range re.FindAllStringSubmatchIndex(ctx.Content, -1) {
		encoded := ctx.Content[loc[2]:loc[3]]
		decoded, ok := decodeBase64(encoded)
		if !ok {
			continue
		}
		urls := findURLsInDecoded(decoded)
		for _, u := range urls {
			if !seen[u] {
				seen[u] = true
				results = append(results, model.Result{
					URL:             u,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(ctx.Content, loc[0]),
					DetectionMethod: "atob_calls",
					Category:        model.CategorizeURL(u),
					Confidence:      model.ConfMedium,
					TechniqueID:     65,
				})
			}
		}
		// If the decoded string itself looks like a URL/path, emit it directly
		if isURLLike(decoded) && !seen[decoded] {
			seen[decoded] = true
			results = append(results, model.Result{
				URL:             decoded,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(ctx.Content, loc[0]),
				DetectionMethod: "atob_calls",
				Category:        model.CategorizeURL(decoded),
				Confidence:      model.ConfMedium,
				TechniqueID:     65,
			})
		}
	}
	return results
}

// technique 66: Reverse/split/join obfuscation patterns
func extractReverseSplitJoin(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	seen := make(map[string]bool)

	// Pattern: "string".split('').reverse().join('')
	reverseRe := model.GetRegex(`["']([^"']+)["']\.split\(["']["']\)\.reverse\(\)\.join\(["']["']\)`)
	for _, loc := range reverseRe.FindAllStringSubmatchIndex(ctx.Content, -1) {
		original := ctx.Content[loc[2]:loc[3]]
		reversed := reverseString(original)
		urls := findURLsInDecoded(reversed)
		for _, u := range urls {
			if !seen[u] {
				seen[u] = true
				results = append(results, model.Result{
					URL:             u,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(ctx.Content, loc[0]),
					DetectionMethod: "reverse_split_join",
					Category:        model.CategorizeURL(u),
					Confidence:      model.ConfLow,
					TechniqueID:     66,
				})
			}
		}
		if isURLLike(reversed) && !seen[reversed] {
			seen[reversed] = true
			results = append(results, model.Result{
				URL:             reversed,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(ctx.Content, loc[0]),
				DetectionMethod: "reverse_split_join",
				Category:        model.CategorizeURL(reversed),
				Confidence:      model.ConfLow,
				TechniqueID:     66,
			})
		}
	}

	// Pattern: "string".split('x').join('y') (string replace)
	splitJoinRe := model.GetRegex(`["']([^"']+)["']\.split\(["']([^"']*)["']\)\.join\(["']([^"']*)["']\)`)
	for _, loc := range splitJoinRe.FindAllStringSubmatchIndex(ctx.Content, -1) {
		original := ctx.Content[loc[2]:loc[3]]
		splitStr := ctx.Content[loc[4]:loc[5]]
		joinStr := ctx.Content[loc[6]:loc[7]]
		// Perform the split/join (equivalent to string replace)
		result := strings.ReplaceAll(original, splitStr, joinStr)
		urls := findURLsInDecoded(result)
		for _, u := range urls {
			if !seen[u] {
				seen[u] = true
				results = append(results, model.Result{
					URL:             u,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(ctx.Content, loc[0]),
					DetectionMethod: "reverse_split_join",
					Category:        model.CategorizeURL(u),
					Confidence:      model.ConfLow,
					TechniqueID:     66,
				})
			}
		}
		if isURLLike(result) && !seen[result] {
			seen[result] = true
			results = append(results, model.Result{
				URL:             result,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(ctx.Content, loc[0]),
				DetectionMethod: "reverse_split_join",
				Category:        model.CategorizeURL(result),
				Confidence:      model.ConfLow,
				TechniqueID:     66,
			})
		}
	}

	return results
}

// reverseString reverses a string rune-by-rune.
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// technique 67: new URL() constructor resolution
func extractURLConstructor(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	seen := make(map[string]bool)

	// Pattern: new URL('/path', 'https://base.com') - two arguments
	twoArgRe := model.GetRegex(`new\s+URL\(\s*["']([^"']+)["']\s*,\s*["']([^"']+)["']\s*\)`)
	for _, loc := range twoArgRe.FindAllStringSubmatchIndex(ctx.Content, -1) {
		path := ctx.Content[loc[2]:loc[3]]
		baseURL := ctx.Content[loc[4]:loc[5]]
		resolved := resolveURL(baseURL, path)
		if !seen[resolved] {
			seen[resolved] = true
			results = append(results, model.Result{
				URL:             resolved,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(ctx.Content, loc[0]),
				DetectionMethod: "url_constructor",
				Category:        model.CategorizeURL(resolved),
				Confidence:      model.ConfHigh,
				TechniqueID:     67,
			})
		}
	}

	// Pattern: new URL('https://full.url') - single argument with absolute URL
	oneArgRe := model.GetRegex(`new\s+URL\(\s*["'](https?://[^"']+)["']\s*\)`)
	for _, loc := range oneArgRe.FindAllStringSubmatchIndex(ctx.Content, -1) {
		url := ctx.Content[loc[2]:loc[3]]
		if !seen[url] {
			seen[url] = true
			results = append(results, model.Result{
				URL:             url,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(ctx.Content, loc[0]),
				DetectionMethod: "url_constructor",
				Category:        model.CategorizeURL(url),
				Confidence:      model.ConfHigh,
				TechniqueID:     67,
			})
		}
	}

	return results
}

// resolveURL resolves a relative path against a base URL.
func resolveURL(baseURL, path string) string {
	// If path is already absolute, return it directly
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}

	// Remove trailing slash from base
	base := strings.TrimRight(baseURL, "/")

	// If path starts with /, it's relative to the origin
	if strings.HasPrefix(path, "/") {
		// Extract origin from base URL (scheme + host)
		origin := extractOrigin(base)
		return origin + path
	}

	// Otherwise, relative to the base path
	return base + "/" + path
}

// extractOrigin returns the scheme + host portion of a URL.
func extractOrigin(url string) string {
	// Find the scheme
	schemeEnd := strings.Index(url, "://")
	if schemeEnd == -1 {
		return url
	}
	// Find the first / after the scheme
	rest := url[schemeEnd+3:]
	slashIdx := strings.Index(rest, "/")
	if slashIdx == -1 {
		return url
	}
	return url[:schemeEnd+3+slashIdx]
}
