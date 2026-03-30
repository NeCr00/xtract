package extract

import (
	"regexp"
	"strings"

	"github.com/NeCr00/xtract/internal/model"
)

// ExtractLayer2 runs all AST-based extraction techniques (15-33) and returns
// the combined results.
func ExtractLayer2(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	results = append(results, extractFetchCalls(ctx)...)
	results = append(results, extractXMLHttpRequest(ctx)...)
	results = append(results, extractAxiosCalls(ctx)...)
	results = append(results, extractJQueryAjax(ctx)...)
	results = append(results, extractSendBeacon(ctx)...)
	results = append(results, extractEventSourceWebSocket(ctx)...)
	results = append(results, extractDynamicImport(ctx)...)
	results = append(results, extractRequireCalls(ctx)...)
	results = append(results, extractLocationAssign(ctx)...)
	results = append(results, extractWindowOpen(ctx)...)
	results = append(results, extractElementSrcHref(ctx)...)
	results = append(results, extractSetAttribute(ctx)...)
	results = append(results, extractInnerHTMLURLs(ctx)...)
	results = append(results, extractPostMessage(ctx)...)
	results = append(results, extractFormAction(ctx)...)
	results = append(results, extractServiceWorker(ctx)...)
	results = append(results, extractWebWorker(ctx)...)
	results = append(results, extractWebpackRequire(ctx)...)
	results = append(results, extractDynamicScript(ctx)...)
	return results
}

// ---------------------------------------------------------------------------
// Helpers for extracting string arguments from JS function calls
// ---------------------------------------------------------------------------

// stringArgRegex matches a quoted string (single, double, or backtick).
var stringArgRegex = regexp.MustCompile("(?:[\"'`])([^\"'`\\n]*?)(?:[\"'`])")

// extractStringArg extracts the string content from a quoted literal.
// It handles single quotes, double quotes, and backtick template literals.
func extractStringArg(s string) string {
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return ""
	}
	q := s[0]
	if q != '\'' && q != '"' && q != '`' {
		return ""
	}
	end := strings.IndexByte(s[1:], q)
	if end < 0 {
		return ""
	}
	return s[1 : 1+end]
}

// extractNthArg extracts the nth (0-based) argument from inside parentheses.
// It understands string delimiters and nested parens/braces/brackets.
func extractNthArg(argsStr string, n int) string {
	depth := 0
	argStart := 0
	argIndex := 0
	inStr := byte(0)

	for i := 0; i < len(argsStr); i++ {
		ch := argsStr[i]

		// Handle string delimiters
		if inStr != 0 {
			if ch == inStr && (i == 0 || argsStr[i-1] != '\\') {
				inStr = 0
			}
			continue
		}
		if ch == '\'' || ch == '"' || ch == '`' {
			inStr = ch
			continue
		}

		switch ch {
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
		case ',':
			if depth == 0 {
				if argIndex == n {
					return strings.TrimSpace(argsStr[argStart:i])
				}
				argIndex++
				argStart = i + 1
			}
		}
	}

	if argIndex == n {
		return strings.TrimSpace(argsStr[argStart:])
	}
	return ""
}

// extractStringValue strips quote characters from a raw argument.
func extractStringValue(raw string) string {
	raw = strings.TrimSpace(raw)
	if len(raw) < 2 {
		return ""
	}
	q := raw[0]
	if q == '\'' || q == '"' || q == '`' {
		if raw[len(raw)-1] == q {
			return raw[1 : len(raw)-1]
		}
		// Might have trailing junk; find the closing quote.
		end := strings.IndexByte(raw[1:], q)
		if end >= 0 {
			return raw[1 : 1+end]
		}
	}
	return ""
}

// isValidURL returns true if the string looks like it could be a URL or path.
func isValidURL(s string) bool {
	if s == "" {
		return false
	}
	// Must contain at least one path separator or protocol indicator
	if strings.HasPrefix(s, "/") ||
		strings.HasPrefix(s, "./") ||
		strings.HasPrefix(s, "../") ||
		strings.HasPrefix(s, "http://") ||
		strings.HasPrefix(s, "https://") ||
		strings.HasPrefix(s, "ws://") ||
		strings.HasPrefix(s, "wss://") ||
		strings.HasPrefix(s, "//") ||
		strings.Contains(s, "/") {
		return true
	}
	return false
}

// makeResult constructs a Result with common fields filled in.
func makeResult(ctx *model.ExtractionContext, url string, line int, method string, technique string, httpMethod string, confidence string, techID int) model.Result {
	cat := model.CategorizeURL(url)
	return model.Result{
		URL:             url,
		SourceFile:      ctx.FileName,
		SourceLine:      line,
		DetectionMethod: technique,
		HTTPMethod:      httpMethod,
		Category:        cat,
		Confidence:      confidence,
		TechniqueID:     techID,
	}
}

// ---------------------------------------------------------------------------
// Technique 15: fetch() calls
// ---------------------------------------------------------------------------

func extractFetchCalls(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	// Match fetch('url' ...) or fetch("url" ...) or fetch(`url` ...)
	re := model.GetRegex(`(?m)\bfetch\s*\(`)
	locs := re.FindAllStringIndex(content, -1)

	for _, loc := range locs {
		callStart := loc[1] // position right after the opening paren
		argsStr := extractParenContent(content, callStart-1)
		if argsStr == "" {
			continue
		}

		// First argument is the URL
		firstArg := extractNthArg(argsStr, 0)
		url := extractStringValue(firstArg)
		if url == "" || !isValidURL(url) {
			continue
		}

		line := model.LineNumber(content, loc[0])
		httpMethod := ""
		var queryParams []string
		var bodyParams []string

		// Try to extract options from second argument
		secondArg := extractNthArg(argsStr, 1)
		if secondArg != "" {
			httpMethod = extractMethodFromObject(secondArg)
			queryParams = extractQueryParams(url)
			bodyParams = extractBodyParams(secondArg)
		} else {
			queryParams = extractQueryParams(url)
		}

		r := makeResult(ctx, url, line, "fetch", "fetch_calls", strings.ToUpper(httpMethod), model.ConfHigh, 15)
		r.QueryParams = queryParams
		r.BodyParams = bodyParams
		results = append(results, r)
	}

	return results
}

// extractParenContent extracts the content between matching parentheses
// starting at position pos which should point to the '('.
func extractParenContent(s string, pos int) string {
	if pos < 0 || pos >= len(s) || s[pos] != '(' {
		return ""
	}
	depth := 0
	inStr := byte(0)
	for i := pos; i < len(s); i++ {
		ch := s[i]
		if inStr != 0 {
			if ch == inStr && (i == 0 || s[i-1] != '\\') {
				inStr = 0
			}
			continue
		}
		if ch == '\'' || ch == '"' || ch == '`' {
			inStr = ch
			continue
		}
		if ch == '(' {
			depth++
		} else if ch == ')' {
			depth--
			if depth == 0 {
				return s[pos+1 : i]
			}
		}
	}
	// Didn't find matching paren; return empty string
	return ""
}

// extractMethodFromObject tries to find method: 'GET' or method: "POST" in an
// options object string.
func extractMethodFromObject(obj string) string {
	re := model.GetRegex(`(?i)method\s*:\s*["'\x60](GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)["'\x60]`)
	m := re.FindStringSubmatch(obj)
	if len(m) > 1 {
		return strings.ToUpper(m[1])
	}
	return ""
}

// extractQueryParams pulls out query parameter names from a URL string.
func extractQueryParams(url string) []string {
	idx := strings.Index(url, "?")
	if idx < 0 {
		return nil
	}
	query := url[idx+1:]
	// Remove fragment
	if fragIdx := strings.Index(query, "#"); fragIdx >= 0 {
		query = query[:fragIdx]
	}
	var params []string
	for _, pair := range strings.Split(query, "&") {
		eqIdx := strings.Index(pair, "=")
		var name string
		if eqIdx >= 0 {
			name = pair[:eqIdx]
		} else {
			name = pair
		}
		name = strings.TrimSpace(name)
		if name != "" {
			params = append(params, name)
		}
	}
	return params
}

// extractBodyParams tries to find body parameter names from an options object
// containing a body property with JSON-like content.
func extractBodyParams(obj string) []string {
	re := model.GetRegex(`(?i)body\s*:\s*(?:JSON\.stringify\s*\()?\s*\{([^}]*)\}`)
	m := re.FindStringSubmatch(obj)
	if len(m) < 2 {
		return nil
	}
	inner := m[1]
	keyRe := model.GetRegex(`(?:["'\x60]?)(\w+)(?:["'\x60]?)\s*:`)
	matches := keyRe.FindAllStringSubmatch(inner, -1)
	var params []string
	for _, km := range matches {
		if len(km) > 1 && km[1] != "" {
			params = append(params, km[1])
		}
	}
	return params
}

// ---------------------------------------------------------------------------
// Technique 16: XMLHttpRequest
// ---------------------------------------------------------------------------

func extractXMLHttpRequest(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	// Match .open('GET', '/api/...')  or .open("POST", url)
	re := model.GetRegex(`\.open\s*\(\s*["'\x60](GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)["'\x60]\s*,\s*["'\x60]([^"'\x60\n]+)["'\x60]`)
	matches := re.FindAllStringSubmatchIndex(content, -1)

	for _, m := range matches {
		if len(m) < 6 {
			continue
		}
		method := strings.ToUpper(content[m[2]:m[3]])
		url := content[m[4]:m[5]]
		if !isValidURL(url) {
			continue
		}
		line := model.LineNumber(content, m[0])
		r := makeResult(ctx, url, line, "xmlhttprequest", "xmlhttprequest", method, model.ConfHigh, 16)
		r.QueryParams = extractQueryParams(url)
		results = append(results, r)
	}

	return results
}

// ---------------------------------------------------------------------------
// Technique 17: axios calls
// ---------------------------------------------------------------------------

func extractAxiosCalls(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	// axios.get(url), axios.post(url), etc.
	methodRe := model.GetRegex(`\baxios\s*\.\s*(get|post|put|delete|patch|head|options)\s*\(`)
	methodMatches := methodRe.FindAllStringSubmatchIndex(content, -1)

	for _, m := range methodMatches {
		if len(m) < 4 {
			continue
		}
		method := strings.ToUpper(content[m[2]:m[3]])
		callStart := m[1] // right after '('
		argsStr := extractParenContent(content, callStart-1)
		if argsStr == "" {
			continue
		}
		firstArg := extractNthArg(argsStr, 0)
		url := extractStringValue(firstArg)
		if url == "" || !isValidURL(url) {
			continue
		}
		line := model.LineNumber(content, m[0])
		r := makeResult(ctx, url, line, "axios", "axios_calls", method, model.ConfHigh, 17)
		r.QueryParams = extractQueryParams(url)
		results = append(results, r)
	}

	// axios.request({url: '...'}) and axios({url: '...'})
	requestRe := model.GetRegex(`\baxios(?:\s*\.\s*request)?\s*\(\s*\{`)
	requestMatches := requestRe.FindAllStringIndex(content, -1)

	for _, loc := range requestMatches {
		// Find the opening paren
		parenIdx := strings.Index(content[loc[0]:], "(")
		if parenIdx < 0 {
			continue
		}
		argsStr := extractParenContent(content, loc[0]+parenIdx)
		if argsStr == "" {
			continue
		}
		url := extractURLFromObject(argsStr)
		if url == "" || !isValidURL(url) {
			continue
		}
		method := extractMethodFromObject(argsStr)
		line := model.LineNumber(content, loc[0])
		r := makeResult(ctx, url, line, "axios", "axios_calls", strings.ToUpper(method), model.ConfHigh, 17)
		r.QueryParams = extractQueryParams(url)
		results = append(results, r)
	}

	return results
}

// extractURLFromObject finds url: '...' in an object literal string.
func extractURLFromObject(obj string) string {
	re := model.GetRegex(`(?i)(?:^|[,{\s])url\s*:\s*["'\x60]([^"'\x60\n]+)["'\x60]`)
	m := re.FindStringSubmatch(obj)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}

// ---------------------------------------------------------------------------
// Technique 18: jQuery AJAX
// ---------------------------------------------------------------------------

func extractJQueryAjax(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	// $.ajax({url: '...'}) and jQuery.ajax({url: '...'})
	ajaxRe := model.GetRegex(`(?:\$|jQuery)\s*\.\s*ajax\s*\(\s*\{`)
	ajaxMatches := ajaxRe.FindAllStringIndex(content, -1)

	for _, loc := range ajaxMatches {
		parenIdx := strings.Index(content[loc[0]:], "(")
		if parenIdx < 0 {
			continue
		}
		argsStr := extractParenContent(content, loc[0]+parenIdx)
		if argsStr == "" {
			continue
		}
		url := extractURLFromObject(argsStr)
		if url == "" || !isValidURL(url) {
			continue
		}
		method := extractMethodFromObject(argsStr)
		if method == "" {
			typeRe := model.GetRegex(`(?i)(?:^|[,{\s])type\s*:\s*["'\x60](GET|POST|PUT|DELETE|PATCH)["'\x60]`)
			tm := typeRe.FindStringSubmatch(argsStr)
			if len(tm) > 1 {
				method = strings.ToUpper(tm[1])
			}
		}
		line := model.LineNumber(content, loc[0])
		r := makeResult(ctx, url, line, "jquery_ajax", "jquery_ajax", strings.ToUpper(method), model.ConfHigh, 18)
		r.QueryParams = extractQueryParams(url)
		results = append(results, r)
	}

	// $.ajax(url) - string argument form
	ajaxStrRe := model.GetRegex(`(?:\$|jQuery)\s*\.\s*ajax\s*\(\s*["'\x60]([^"'\x60\n]+)["'\x60]`)
	ajaxStrMatches := ajaxStrRe.FindAllStringSubmatchIndex(content, -1)

	for _, m := range ajaxStrMatches {
		if len(m) < 4 {
			continue
		}
		url := content[m[2]:m[3]]
		if !isValidURL(url) {
			continue
		}
		line := model.LineNumber(content, m[0])
		r := makeResult(ctx, url, line, "jquery_ajax", "jquery_ajax", "", model.ConfHigh, 18)
		r.QueryParams = extractQueryParams(url)
		results = append(results, r)
	}

	// $.get(url), $.post(url), $.getJSON(url)
	shortRe := model.GetRegex(`(?:\$|jQuery)\s*\.\s*(get|post|getJSON)\s*\(\s*["'\x60]([^"'\x60\n]+)["'\x60]`)
	shortMatches := shortRe.FindAllStringSubmatchIndex(content, -1)

	for _, m := range shortMatches {
		if len(m) < 6 {
			continue
		}
		methodStr := content[m[2]:m[3]]
		url := content[m[4]:m[5]]
		if !isValidURL(url) {
			continue
		}

		httpMethod := "GET"
		switch strings.ToLower(methodStr) {
		case "post":
			httpMethod = "POST"
		case "get", "getjson":
			httpMethod = "GET"
		}

		line := model.LineNumber(content, m[0])
		r := makeResult(ctx, url, line, "jquery_ajax", "jquery_ajax", httpMethod, model.ConfHigh, 18)
		r.QueryParams = extractQueryParams(url)
		results = append(results, r)
	}

	return results
}

// ---------------------------------------------------------------------------
// Technique 19: navigator.sendBeacon()
// ---------------------------------------------------------------------------

func extractSendBeacon(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	re := model.GetRegex(`\bnavigator\s*\.\s*sendBeacon\s*\(\s*["'\x60]([^"'\x60\n]+)["'\x60]`)
	matches := re.FindAllStringSubmatchIndex(content, -1)

	for _, m := range matches {
		if len(m) < 4 {
			continue
		}
		url := content[m[2]:m[3]]
		if !isValidURL(url) {
			continue
		}
		line := model.LineNumber(content, m[0])
		r := makeResult(ctx, url, line, "sendbeacon", "sendbeacon", "POST", model.ConfHigh, 19)
		r.QueryParams = extractQueryParams(url)
		results = append(results, r)
	}

	return results
}

// ---------------------------------------------------------------------------
// Technique 20: EventSource / WebSocket
// ---------------------------------------------------------------------------

func extractEventSourceWebSocket(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	re := model.GetRegex(`\bnew\s+(EventSource|WebSocket)\s*\(\s*["'\x60]([^"'\x60\n]+)["'\x60]`)
	matches := re.FindAllStringSubmatchIndex(content, -1)

	for _, m := range matches {
		if len(m) < 6 {
			continue
		}
		kind := content[m[2]:m[3]]
		url := content[m[4]:m[5]]
		if url == "" {
			continue
		}
		line := model.LineNumber(content, m[0])
		cat := model.CatWebSocket
		if kind == "EventSource" {
			cat = model.CategorizeURL(url)
			// EventSource uses Server-Sent Events over HTTP
			if cat == model.CatPageRoute {
				cat = model.CatAPIEndpoint
			}
		}
		r := model.Result{
			URL:             url,
			SourceFile:      ctx.FileName,
			SourceLine:      line,
			DetectionMethod: "eventsource_websocket",
			Category:        cat,
			Confidence:      model.ConfHigh,
			TechniqueID:     20,
		}
		r.QueryParams = extractQueryParams(url)
		results = append(results, r)
	}

	return results
}

// ---------------------------------------------------------------------------
// Technique 21: Dynamic import()
// ---------------------------------------------------------------------------

func extractDynamicImport(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	// import('path') or import("path") - use word boundary to avoid matching
	// reimport, etc. Also avoid matching static import statements.
	re := model.GetRegex(`(?:^|[^.\w])import\s*\(\s*["'\x60]([^"'\x60\n]+)["'\x60]\s*\)`)
	matches := re.FindAllStringSubmatchIndex(content, -1)

	for _, m := range matches {
		if len(m) < 4 {
			continue
		}
		path := content[m[2]:m[3]]
		if path == "" {
			continue
		}
		line := model.LineNumber(content, m[0])
		r := makeResult(ctx, path, line, "dynamic_import", "dynamic_import", "", model.ConfHigh, 21)
		results = append(results, r)
	}

	return results
}

// ---------------------------------------------------------------------------
// Technique 22: require() calls
// ---------------------------------------------------------------------------

func extractRequireCalls(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	re := model.GetRegex(`\brequire\s*\(\s*["'\x60]([^"'\x60\n]+)["'\x60]\s*\)`)
	matches := re.FindAllStringSubmatchIndex(content, -1)

	for _, m := range matches {
		if len(m) < 4 {
			continue
		}
		modulePath := content[m[2]:m[3]]
		// Only extract if it contains '/' (local module path)
		if !strings.Contains(modulePath, "/") {
			continue
		}
		line := model.LineNumber(content, m[0])
		r := makeResult(ctx, modulePath, line, "require", "require_calls", "", model.ConfHigh, 22)
		results = append(results, r)
	}

	return results
}

// ---------------------------------------------------------------------------
// Technique 23: document.location / window.location assignments
// ---------------------------------------------------------------------------

func extractLocationAssign(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	// document.location = url, document.location.href = url,
	// window.location = url, window.location.href = url,
	// location.href = url
	assignRe := model.GetRegex(`(?:document|window)\s*\.\s*location\s*(?:\.href)?\s*=\s*["'\x60]([^"'\x60\n]+)["'\x60]`)
	assignMatches := assignRe.FindAllStringSubmatchIndex(content, -1)

	for _, m := range assignMatches {
		if len(m) < 4 {
			continue
		}
		url := content[m[2]:m[3]]
		if !isValidURL(url) {
			continue
		}
		line := model.LineNumber(content, m[0])
		r := makeResult(ctx, url, line, "location_assign", "location_assign", "", model.ConfHigh, 23)
		r.QueryParams = extractQueryParams(url)
		results = append(results, r)
	}

	// location.assign(url) and location.replace(url)
	funcRe := model.GetRegex(`(?:document\s*\.\s*|window\s*\.\s*)?location\s*\.\s*(assign|replace)\s*\(\s*["'\x60]([^"'\x60\n]+)["'\x60]`)
	funcMatches := funcRe.FindAllStringSubmatchIndex(content, -1)

	for _, m := range funcMatches {
		if len(m) < 6 {
			continue
		}
		url := content[m[4]:m[5]]
		if !isValidURL(url) {
			continue
		}
		line := model.LineNumber(content, m[0])
		r := makeResult(ctx, url, line, "location_assign", "location_assign", "", model.ConfHigh, 23)
		r.QueryParams = extractQueryParams(url)
		results = append(results, r)
	}

	return results
}

// ---------------------------------------------------------------------------
// Technique 24: window.open()
// ---------------------------------------------------------------------------

func extractWindowOpen(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	re := model.GetRegex(`\bwindow\s*\.\s*open\s*\(\s*["'\x60]([^"'\x60\n]+)["'\x60]`)
	matches := re.FindAllStringSubmatchIndex(content, -1)

	for _, m := range matches {
		if len(m) < 4 {
			continue
		}
		url := content[m[2]:m[3]]
		if !isValidURL(url) {
			continue
		}
		line := model.LineNumber(content, m[0])
		r := makeResult(ctx, url, line, "window_open", "window_open", "", model.ConfHigh, 24)
		r.QueryParams = extractQueryParams(url)
		results = append(results, r)
	}

	return results
}

// ---------------------------------------------------------------------------
// Technique 25: element.src / .href / .action property assignments
// ---------------------------------------------------------------------------

func extractElementSrcHref(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	re := model.GetRegex(`\.\s*(src|href|action)\s*=\s*["'\x60]([^"'\x60\n]+)["'\x60]`)
	matches := re.FindAllStringSubmatchIndex(content, -1)

	for _, m := range matches {
		if len(m) < 6 {
			continue
		}
		url := content[m[4]:m[5]]
		if !isValidURL(url) {
			continue
		}
		line := model.LineNumber(content, m[0])
		r := makeResult(ctx, url, line, "element_src_href", "element_src_href", "", model.ConfMedium, 25)
		r.QueryParams = extractQueryParams(url)
		results = append(results, r)
	}

	return results
}

// ---------------------------------------------------------------------------
// Technique 26: setAttribute('src'|'href'|'action', url)
// ---------------------------------------------------------------------------

func extractSetAttribute(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	re := model.GetRegex(`\.setAttribute\s*\(\s*["'\x60](src|href|action)["'\x60]\s*,\s*["'\x60]([^"'\x60\n]+)["'\x60]\s*\)`)
	matches := re.FindAllStringSubmatchIndex(content, -1)

	for _, m := range matches {
		if len(m) < 6 {
			continue
		}
		url := content[m[4]:m[5]]
		if !isValidURL(url) {
			continue
		}
		line := model.LineNumber(content, m[0])
		r := makeResult(ctx, url, line, "set_attribute", "set_attribute", "", model.ConfHigh, 26)
		r.QueryParams = extractQueryParams(url)
		results = append(results, r)
	}

	return results
}

// ---------------------------------------------------------------------------
// Technique 27: innerHTML / outerHTML URLs
// ---------------------------------------------------------------------------

func extractInnerHTMLURLs(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	// Match .innerHTML = '...' or .outerHTML = '...'
	// The value could be quite long; capture up to a reasonable length.
	re := model.GetRegex(`\.(?:inner|outer)HTML\s*=\s*["'\x60]([^"'\x60]+)["'\x60]`)
	matches := re.FindAllStringSubmatchIndex(content, -1)

	// Also match += assignments
	reAppend := model.GetRegex(`\.(?:inner|outer)HTML\s*\+=\s*["'\x60]([^"'\x60]+)["'\x60]`)
	appendMatches := reAppend.FindAllStringSubmatchIndex(content, -1)
	matches = append(matches, appendMatches...)

	urlRe := model.GetRegex(`(?:src|href|action)\s*=\s*\\?["'\x60]([^"'\x60\\\s>]+)\\?["'\x60]`)

	for _, m := range matches {
		if len(m) < 4 {
			continue
		}
		htmlStr := content[m[2]:m[3]]
		line := model.LineNumber(content, m[0])

		// Scan HTML string for URLs in src/href/action attributes
		urlMatches := urlRe.FindAllStringSubmatch(htmlStr, -1)
		for _, um := range urlMatches {
			if len(um) < 2 {
				continue
			}
			url := um[1]
			if !isValidURL(url) {
				continue
			}
			r := makeResult(ctx, url, line, "innerhtml_urls", "innerhtml_urls", "", model.ConfMedium, 27)
			r.QueryParams = extractQueryParams(url)
			results = append(results, r)
		}
	}

	return results
}

// ---------------------------------------------------------------------------
// Technique 28: postMessage targetOrigin
// ---------------------------------------------------------------------------

func extractPostMessage(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	// .postMessage(data, 'origin')
	re := model.GetRegex(`\.postMessage\s*\(`)
	locs := re.FindAllStringIndex(content, -1)

	for _, loc := range locs {
		callStart := loc[1] // right after '('
		argsStr := extractParenContent(content, callStart-1)
		if argsStr == "" {
			continue
		}

		// targetOrigin is the second argument
		secondArg := extractNthArg(argsStr, 1)
		origin := extractStringValue(secondArg)
		if origin == "" || origin == "*" {
			continue
		}
		if !isValidURL(origin) {
			continue
		}
		line := model.LineNumber(content, loc[0])
		r := makeResult(ctx, origin, line, "postmessage", "postmessage", "", model.ConfMedium, 28)
		results = append(results, r)
	}

	return results
}

// ---------------------------------------------------------------------------
// Technique 29: form action URLs
// ---------------------------------------------------------------------------

func extractFormAction(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	// Match <form ...> tags that contain an action attribute.
	// We capture the full tag to inspect the method attribute too.
	formTagRe := model.GetRegex(`(?i)<form\b([^>]*)>`)
	actionRe := model.GetRegex(`(?i)\baction\s*=\s*["'\x60]([^"'\x60\n>]+)["'\x60]`)
	methodRe := model.GetRegex(`(?i)\bmethod\s*=\s*["'\x60]?(GET|POST)["'\x60]?`)
	formTagMatches := formTagRe.FindAllStringSubmatchIndex(content, -1)

	for _, m := range formTagMatches {
		if len(m) < 4 {
			continue
		}
		attrs := content[m[2]:m[3]]
		am := actionRe.FindStringSubmatch(attrs)
		if len(am) < 2 {
			continue
		}
		url := am[1]
		if !isValidURL(url) {
			continue
		}
		line := model.LineNumber(content, m[0])
		httpMethod := "GET" // HTML default
		if mm := methodRe.FindStringSubmatch(attrs); len(mm) > 1 {
			httpMethod = strings.ToUpper(mm[1])
		}
		r := makeResult(ctx, url, line, "form_action", "form_action", httpMethod, model.ConfMedium, 29)
		r.QueryParams = extractQueryParams(url)
		results = append(results, r)
	}

	// Also match action= outside of <form> tags (in JS-constructed HTML strings)
	reJS := model.GetRegex(`(?:["'\x60]|\\["'\x60])\s*action\s*=\s*(?:\\?["'\x60])([^"'\x60\\\n>]+)(?:\\?["'\x60])`)
	jsMatches := reJS.FindAllStringSubmatchIndex(content, -1)

	for _, m := range jsMatches {
		if len(m) < 4 {
			continue
		}
		url := content[m[2]:m[3]]
		if !isValidURL(url) {
			continue
		}
		line := model.LineNumber(content, m[0])
		r := makeResult(ctx, url, line, "form_action", "form_action", "POST", model.ConfMedium, 29)
		r.QueryParams = extractQueryParams(url)
		results = append(results, r)
	}

	return results
}

// ---------------------------------------------------------------------------
// Technique 30: Service Worker registration
// ---------------------------------------------------------------------------

func extractServiceWorker(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	re := model.GetRegex(`\bnavigator\s*\.\s*serviceWorker\s*\.\s*register\s*\(\s*["'\x60]([^"'\x60\n]+)["'\x60]`)
	matches := re.FindAllStringSubmatchIndex(content, -1)

	for _, m := range matches {
		if len(m) < 4 {
			continue
		}
		url := content[m[2]:m[3]]
		if url == "" {
			continue
		}
		line := model.LineNumber(content, m[0])
		r := makeResult(ctx, url, line, "serviceworker", "serviceworker", "", model.ConfHigh, 30)
		results = append(results, r)
	}

	return results
}

// ---------------------------------------------------------------------------
// Technique 31: Web Worker / SharedWorker
// ---------------------------------------------------------------------------

func extractWebWorker(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	re := model.GetRegex(`\bnew\s+(?:Worker|SharedWorker)\s*\(\s*["'\x60]([^"'\x60\n]+)["'\x60]`)
	matches := re.FindAllStringSubmatchIndex(content, -1)

	for _, m := range matches {
		if len(m) < 4 {
			continue
		}
		url := content[m[2]:m[3]]
		if url == "" {
			continue
		}
		line := model.LineNumber(content, m[0])
		r := makeResult(ctx, url, line, "webworker", "webworker", "", model.ConfHigh, 31)
		results = append(results, r)
	}

	return results
}

// ---------------------------------------------------------------------------
// Technique 32: Webpack require patterns
// ---------------------------------------------------------------------------

func extractWebpackRequire(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	// require.ensure([], function) - the first string arg in the callback body
	// is typically the module path
	ensureRe := model.GetRegex(`\brequire\s*\.\s*ensure\s*\(\s*\[([^\]]*)\]`)
	ensureMatches := ensureRe.FindAllStringSubmatchIndex(content, -1)

	for _, m := range ensureMatches {
		if len(m) < 4 {
			continue
		}
		deps := content[m[2]:m[3]]
		// Extract string literals from the dependency array
		strRe := model.GetRegex(`["'\x60]([^"'\x60\n]+)["'\x60]`)
		strMatches := strRe.FindAllStringSubmatch(deps, -1)
		line := model.LineNumber(content, m[0])
		for _, sm := range strMatches {
			if len(sm) < 2 {
				continue
			}
			path := sm[1]
			if !strings.Contains(path, "/") {
				continue
			}
			r := makeResult(ctx, path, line, "webpack_require", "webpack_require", "", model.ConfMedium, 32)
			results = append(results, r)
		}
	}

	// require.context('dir', ...)
	contextRe := model.GetRegex(`\brequire\s*\.\s*context\s*\(\s*["'\x60]([^"'\x60\n]+)["'\x60]`)
	contextMatches := contextRe.FindAllStringSubmatchIndex(content, -1)

	for _, m := range contextMatches {
		if len(m) < 4 {
			continue
		}
		dir := content[m[2]:m[3]]
		if dir == "" {
			continue
		}
		line := model.LineNumber(content, m[0])
		r := makeResult(ctx, dir, line, "webpack_require", "webpack_require", "", model.ConfMedium, 32)
		results = append(results, r)
	}

	// __webpack_require__(id) with string path
	wpRe := model.GetRegex(`\b__webpack_require__\s*\(\s*["'\x60]([^"'\x60\n]+)["'\x60]\s*\)`)
	wpMatches := wpRe.FindAllStringSubmatchIndex(content, -1)

	for _, m := range wpMatches {
		if len(m) < 4 {
			continue
		}
		path := content[m[2]:m[3]]
		if !strings.Contains(path, "/") {
			continue
		}
		line := model.LineNumber(content, m[0])
		r := makeResult(ctx, path, line, "webpack_require", "webpack_require", "", model.ConfMedium, 32)
		results = append(results, r)
	}

	return results
}

// ---------------------------------------------------------------------------
// Technique 33: Dynamic script loading
// ---------------------------------------------------------------------------

func extractDynamicScript(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	// Pattern 1: createElement('script') followed by .src = 'url' within nearby lines.
	// We look for createElement('script') and then scan ahead for .src assignments.
	createRe := model.GetRegex(`createElement\s*\(\s*["'\x60]script["'\x60]\s*\)`)
	createMatches := createRe.FindAllStringIndex(content, -1)

	srcRe := model.GetRegex(`\.src\s*=\s*["'\x60]([^"'\x60\n]+)["'\x60]`)

	for _, loc := range createMatches {
		// Scan ahead up to 500 characters for a .src assignment
		end := loc[1] + 500
		if end > len(content) {
			end = len(content)
		}
		ahead := content[loc[1]:end]
		srcMatches := srcRe.FindAllStringSubmatch(ahead, 1)
		if len(srcMatches) == 0 {
			continue
		}
		url := srcMatches[0][1]
		if !isValidURL(url) {
			continue
		}
		line := model.LineNumber(content, loc[0])
		r := makeResult(ctx, url, line, "dynamic_script", "dynamic_script", "", model.ConfHigh, 33)
		results = append(results, r)
	}

	// Pattern 2: Direct patterns for dynamically loaded scripts via other means
	// e.g., document.write('<script src="url"></script>')
	writeRe := model.GetRegex(`(?:document\.write|document\.writeln)\s*\(\s*["'\x60][^"'\x60]*<script[^>]*\bsrc\s*=\s*\\?["'\x60]([^"'\x60\\\n>]+)\\?["'\x60]`)
	writeMatches := writeRe.FindAllStringSubmatchIndex(content, -1)

	for _, m := range writeMatches {
		if len(m) < 4 {
			continue
		}
		url := content[m[2]:m[3]]
		if !isValidURL(url) {
			continue
		}
		line := model.LineNumber(content, m[0])
		r := makeResult(ctx, url, line, "dynamic_script", "dynamic_script", "", model.ConfHigh, 33)
		results = append(results, r)
	}

	// Pattern 3: Loading scripts via insertBefore/appendChild after setting src
	// Already covered by Pattern 1 since we look for createElement + src assignment

	return results
}
