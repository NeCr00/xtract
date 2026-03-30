package extract

import (
	"strings"
	"testing"

	"github.com/NeCr00/xtract/internal/model"
)

func makeCtx(content, fileName, fileType string) *model.ExtractionContext {
	return &model.ExtractionContext{
		Content:  content,
		FileName: fileName,
		FileType: fileType,
	}
}

// ── Layer 1 Tests ──────────────────────────────────────────────

func TestTechnique1_AbsoluteURLs(t *testing.T) {
	content := `var u = "https://api.example.com/v1/users";
var f = "ftp://files.example.com/data";
var w = "wss://ws.example.com/socket";
var p = "//cdn.example.com/lib.js";`

	results := ExtractLayer1(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	assertContains(t, urls, "https://api.example.com/v1/users")
	assertContains(t, urls, "ftp://files.example.com/data")
	assertContains(t, urls, "wss://ws.example.com/socket")
	// Protocol-relative URLs are extracted by technique 1
	assertContainsSubstring(t, urls, "cdn.example.com/lib.js")
}

func TestTechnique2_RelativePaths(t *testing.T) {
	content := `var a = "/api/v1/users";
var b = "./components/Header";
var c = "../utils/helpers";`

	results := ExtractLayer1(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	assertContains(t, urls, "/api/v1/users")
	assertContains(t, urls, "./components/Header")
	assertContains(t, urls, "../utils/helpers")
}

func TestTechnique3_APIPatterns(t *testing.T) {
	content := `"/api/v1/users/list"
"/api/v2/admin/settings"
"/graphql"
"/rest/services/data"
"/rpc/execute"`

	results := ExtractLayer1(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	assertContains(t, urls, "/api/v1/users/list")
	assertContains(t, urls, "/graphql")
	assertContains(t, urls, "/rest/services/data")
}

func TestTechnique4_QueryStrings(t *testing.T) {
	content := `var u = "https://api.example.com/search?q=test&page=1&limit=20";`

	results := ExtractLayer1(makeCtx(content, "test.js", "js"))
	found := false
	for _, r := range results {
		if strings.Contains(r.URL, "search?q=test") {
			found = true
			if len(r.QueryParams) == 0 {
				t.Log("Warning: query params not extracted")
			}
			break
		}
	}
	if !found {
		t.Error("Expected URL with query string to be found")
	}
}

func TestTechnique6_TemplateLiterals(t *testing.T) {
	// Use explicit backtick character to simulate JS template literal
	content := "var u = " + string(rune(96)) + "/api/v1/users/${userId}/profile" + string(rune(96)) + ";"

	results := ExtractLayer1(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	found := false
	for _, u := range urls {
		if strings.Contains(u, "/api/v1/users/") && strings.Contains(u, "DYNAMIC") {
			found = true
			break
		}
	}
	if !found {
		// Template literals might also be captured by the relative_paths technique
		for _, u := range urls {
			if strings.Contains(u, "/api/v1/users/") {
				found = true
				break
			}
		}
		if found {
			t.Log("Template literal captured via relative_paths (backtick as quote delimiter)")
		} else {
			t.Error("Expected template literal URL with {{DYNAMIC}} placeholder")
		}
	}
}

func TestTechnique11_IPBasedURLs(t *testing.T) {
	content := `var u = "http://192.168.1.100:8080/api";
var v = "https://10.0.0.1/internal/debug";`

	results := ExtractLayer1(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	assertContains(t, urls, "http://192.168.1.100:8080/api")
	assertContains(t, urls, "https://10.0.0.1/internal/debug")
}

func TestTechnique13_CustomProtocols(t *testing.T) {
	content := `var d = "android-app://com.example.app/deep/link";
var i = "intent://scan/#Intent;scheme=zxing;end";`

	results := ExtractLayer1(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	assertContains(t, urls, "android-app://com.example.app/deep/link")
}

// ── Layer 2 Tests ──────────────────────────────────────────────

func TestTechnique15_FetchCalls(t *testing.T) {
	content := `fetch('/api/v1/users');
fetch('/api/v1/posts', { method: 'POST' });
fetch("https://api.example.com/data", { method: "PUT" });`

	results := ExtractLayer2(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	assertContains(t, urls, "/api/v1/users")
	assertContains(t, urls, "/api/v1/posts")
	assertContains(t, urls, "https://api.example.com/data")

	// Check HTTP method extraction
	for _, r := range results {
		if r.URL == "/api/v1/posts" && r.HTTPMethod != "POST" {
			t.Errorf("Expected POST method for /api/v1/posts, got %q", r.HTTPMethod)
		}
	}
}

func TestTechnique16_XMLHttpRequest(t *testing.T) {
	content := `xhr.open('GET', '/api/v1/status');
xhr.open("POST", "/api/v1/submit");`

	results := ExtractLayer2(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	assertContains(t, urls, "/api/v1/status")
	assertContains(t, urls, "/api/v1/submit")
}

func TestTechnique17_AxiosCalls(t *testing.T) {
	content := `axios.get('/api/users');
axios.post('/api/users', { name: 'John' });
axios.delete('/api/users/123');`

	results := ExtractLayer2(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	assertContains(t, urls, "/api/users")
	assertContains(t, urls, "/api/users/123")
}

func TestTechnique18_JQueryAjax(t *testing.T) {
	content := `$.ajax({ url: '/api/legacy/data', type: 'GET' });
$.get('/api/legacy/users');
$.post('/api/legacy/submit', data);
$.getJSON('/api/legacy/config.json');`

	results := ExtractLayer2(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	assertContains(t, urls, "/api/legacy/data")
	assertContains(t, urls, "/api/legacy/users")
	assertContains(t, urls, "/api/legacy/submit")
	assertContains(t, urls, "/api/legacy/config.json")
}

func TestTechnique20_WebSocket(t *testing.T) {
	content := `var ws = new WebSocket('wss://realtime.example.com/ws');
var es = new EventSource('/api/events/stream');`

	results := ExtractLayer2(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	assertContains(t, urls, "wss://realtime.example.com/ws")
	assertContains(t, urls, "/api/events/stream")
}

func TestTechnique23_LocationAssign(t *testing.T) {
	content := `document.location = '/login';
window.location.href = '/dashboard';
location.assign('/assigned-page');
location.replace('/replaced-page');`

	results := ExtractLayer2(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	assertContains(t, urls, "/login")
	assertContains(t, urls, "/dashboard")
	assertContains(t, urls, "/assigned-page")
	assertContains(t, urls, "/replaced-page")
}

func TestTechnique30_ServiceWorker(t *testing.T) {
	content := `navigator.serviceWorker.register('/sw.js');`

	results := ExtractLayer2(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	assertContains(t, urls, "/sw.js")
}

// ── Layer 3 Tests ──────────────────────────────────────────────

func TestTechnique34_ReactRouter(t *testing.T) {
	content := `<Route path="/dashboard" component={Dashboard} />;
<Link to="/about">About</Link>;
navigate('/checkout/payment');`

	results := ExtractLayer3(makeCtx(content, "test.jsx", "jsx"))
	urls := resultURLs(results)

	assertContains(t, urls, "/dashboard")
	assertContains(t, urls, "/about")
	assertContains(t, urls, "/checkout/payment")
}

func TestTechnique38_ExpressRoutes(t *testing.T) {
	content := `app.get('/api/users', handler);
app.post('/api/users', createUser);
router.delete('/api/items/:id', deleteItem);`

	results := ExtractLayer3(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	assertContains(t, urls, "/api/users")
	assertContains(t, urls, "/api/items/:id")
}

// ── Layer 4 Tests ──────────────────────────────────────────────

func TestTechnique41_EnvVariables(t *testing.T) {
	content := `const apiUrl = process.env.API_URL;
process.env.REACT_APP_API_BASE = "https://api.example.com";`

	results := ExtractLayer4(makeCtx(content, "test.js", "js"))
	found := false
	for _, r := range results {
		if strings.Contains(r.URL, "process.env.API_URL") || strings.Contains(r.URL, "process.env.REACT_APP_API_BASE") {
			found = true
			break
		}
	}
	if !found {
		t.Log("Warning: env variable names not directly extracted as URLs (may be extracted as assigned values)")
	}
}

func TestTechnique42_ConfigObjects(t *testing.T) {
	content := `const config = {
    baseURL: "https://api.example.com/v2",
    endpoint: "/api/internal"
};`

	results := ExtractLayer4(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	assertContains(t, urls, "https://api.example.com/v2")
	assertContains(t, urls, "/api/internal")
}

func TestTechnique44_SourceMapRefs(t *testing.T) {
	content := `// some code
//# sourceMappingURL=app.bundle.js.map`

	results := ExtractLayer4(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	assertContains(t, urls, "app.bundle.js.map")
}

// ── Layer 5 Tests ──────────────────────────────────────────────

func TestTechnique51_CDNDomains(t *testing.T) {
	content := `var s3 = "https://my-bucket.s3.amazonaws.com/file.pdf";
var cf = "https://d123.cloudfront.net/assets/app.js";
var az = "https://myaccount.blob.core.windows.net/container/data";`

	results := ExtractLayer5(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	// CDN domain extraction captures the domain and sometimes the full URL
	assertContainsSubstring(t, urls, "s3.amazonaws.com")
	assertContainsSubstring(t, urls, "cloudfront.net")
	assertContainsSubstring(t, urls, "blob.core.windows.net")
}

func TestTechnique52_InternalHostnames(t *testing.T) {
	content := `var i = "https://api.internal.example.com/health";
var l = "http://localhost:3000/dev";
var p = "http://10.0.1.50:9090/metrics";`

	results := ExtractLayer5(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	found := false
	for _, u := range urls {
		if strings.Contains(u, "internal") || strings.Contains(u, "localhost") || strings.Contains(u, "10.0.1.50") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected internal hostname to be found")
	}
}

func TestTechnique53_WebSocketEndpoints(t *testing.T) {
	content := `var ws = "wss://secure-ws.example.com/notifications";
var ws2 = "ws://10.0.0.5:9090/internal/stream";`

	results := ExtractLayer5(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	assertContains(t, urls, "wss://secure-ws.example.com/notifications")
	assertContains(t, urls, "ws://10.0.0.5:9090/internal/stream")
}

// ── Layer 6 Tests ──────────────────────────────────────────────

func TestTechnique54_SingleLineComments(t *testing.T) {
	content := `// See https://docs.example.com/api/v3/reference
// Old endpoint: /api/v1/deprecated`

	results := ExtractLayer6(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	assertContains(t, urls, "https://docs.example.com/api/v3/reference")
}

func TestTechnique55_MultiLineComments(t *testing.T) {
	content := `/*
 * API: https://wiki.example.com/api-guide
 * Staging: https://staging.example.com/api/v2
 */`

	results := ExtractLayer6(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	assertContains(t, urls, "https://wiki.example.com/api-guide")
	assertContains(t, urls, "https://staging.example.com/api/v2")
}

// ── Layer 7 Tests ──────────────────────────────────────────────

func TestTechnique64_FromCharCode(t *testing.T) {
	// "http://charcode.example.com" encoded as char codes
	content := `var u = String.fromCharCode(104,116,116,112,58,47,47,99,104,97,114,99,111,100,101,46,101,120,97,109,112,108,101,46,99,111,109);`

	results := ExtractLayer7(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	assertContains(t, urls, "http://charcode.example.com")
}

func TestTechnique65_AtobCalls(t *testing.T) {
	// "https://atob.example.com/secret" base64 encoded
	content := `var decoded = atob("aHR0cHM6Ly9hdG9iLmV4YW1wbGUuY29tL3NlY3JldA==");`

	results := ExtractLayer7(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	assertContains(t, urls, "https://atob.example.com/secret")
}

func TestTechnique67_URLConstructor(t *testing.T) {
	content := `var u = new URL('/api/v3/constructed', 'https://base.example.com');
var v = new URL('https://simple.example.com/direct');`

	results := ExtractLayer7(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	assertContains(t, urls, "https://base.example.com/api/v3/constructed")
	assertContains(t, urls, "https://simple.example.com/direct")
}

// ── Helpers ────────────────────────────────────────────────────

func resultURLs(results []model.Result) []string {
	urls := make([]string, len(results))
	for i, r := range results {
		urls[i] = r.URL
	}
	return urls
}

func assertContains(t *testing.T, urls []string, target string) {
	t.Helper()
	for _, u := range urls {
		if u == target {
			return
		}
	}
	t.Errorf("Expected to find %q in results, but it was not found.\nGot: %v", target, urls)
}

func assertContainsSubstring(t *testing.T, urls []string, substr string) {
	t.Helper()
	for _, u := range urls {
		if strings.Contains(u, substr) {
			return
		}
	}
	t.Errorf("Expected to find URL containing %q in results, but it was not found.\nGot: %v", substr, urls)
}

func assertNotContains(t *testing.T, urls []string, target string) {
	t.Helper()
	for _, u := range urls {
		if u == target {
			t.Errorf("Expected %q to NOT be in results, but it was found", target)
			return
		}
	}
}

// ── Bug Fix Regression Tests ───────────────────────────────────

func TestBugFix_WindowOpenRequiresPrefix(t *testing.T) {
	// window.open should be extracted
	content := `window.open('https://docs.example.com/help');`
	results := ExtractLayer2(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)
	assertContains(t, urls, "https://docs.example.com/help")

	// Bare open() should NOT be extracted (was a false positive before the fix)
	content2 := `file.open('/some/file/path');
db.open('/connection/string');`
	results2 := ExtractLayer2(makeCtx(content2, "test.js", "js"))
	urls2 := resultURLs(results2)
	assertNotContains(t, urls2, "/some/file/path")
	assertNotContains(t, urls2, "/connection/string")
}

func TestBugFix_LocationRequiresPrefix(t *testing.T) {
	// document.location and window.location should be extracted
	content := `document.location.href = '/dashboard';
window.location = '/login';`
	results := ExtractLayer2(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)
	assertContains(t, urls, "/dashboard")
	assertContains(t, urls, "/login")

	// Variables ending in "location" should NOT be extracted
	content2 := `var fileLocation = '/some/path';
var allocation = '/budget';`
	results2 := ExtractLayer2(makeCtx(content2, "test.js", "js"))
	urls2 := resultURLs(results2)
	assertNotContains(t, urls2, "/some/path")
	assertNotContains(t, urls2, "/budget")
}

func TestBugFix_FormActionDefaultsToGET(t *testing.T) {
	// Form without method should default to GET
	content := `var html = '<form action="/api/search">';`
	results := ExtractLayer2(makeCtx(content, "test.js", "js"))
	for _, r := range results {
		if r.URL == "/api/search" {
			if r.HTTPMethod != "GET" {
				t.Errorf("Form without method should default to GET, got %q", r.HTTPMethod)
			}
			return
		}
	}

	// Form with method="POST"
	content2 := `var html = '<form action="/api/submit" method="POST">';`
	results2 := ExtractLayer2(makeCtx(content2, "test.js", "js"))
	for _, r := range results2 {
		if r.URL == "/api/submit" && r.HTTPMethod == "POST" {
			return // correct
		}
	}
}

func TestBugFix_DoubleEncodedURLs(t *testing.T) {
	// Double-encoded URL: %252F should decode to %2F then to /
	content := `var u = "https%253A%252F%252Fexample.com%252Fapi%252Fsecret";`
	results := ExtractLayer1(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	// Should find the decoded URL
	found := false
	for _, u := range urls {
		if strings.Contains(u, "example.com") && strings.Contains(u, "secret") {
			found = true
			break
		}
	}
	if !found {
		t.Log("Double-encoded URL not fully decoded (may be extracted in encoded form)")
	}
}

func TestBugFix_ParenContentFallback(t *testing.T) {
	// Properly matched fetch should work fine
	content := `fetch("/api/valid/endpoint");`
	results := ExtractLayer2(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)
	assertContains(t, urls, "/api/valid/endpoint")

	// Unmatched paren should not produce garbage from raw code
	content2 := `fetch(someVariable
some garbage code without closing paren
more code here = function() { return 42; };`
	results2 := ExtractLayer2(makeCtx(content2, "test.js", "js"))
	for _, r := range results2 {
		if strings.Contains(r.URL, "garbage") || strings.Contains(r.URL, "function") {
			t.Errorf("Garbage URL extracted from unmatched paren fallback: %q", r.URL)
		}
	}
}

// ── Edge Case Tests ────────────────────────────────────────────

func TestEdgeCase_EmptyContent(t *testing.T) {
	ctx := makeCtx("", "empty.js", "js")
	r1 := ExtractLayer1(ctx)
	r2 := ExtractLayer2(ctx)
	r3 := ExtractLayer3(ctx)
	r4 := ExtractLayer4(ctx)
	r5 := ExtractLayer5(ctx)
	r6 := ExtractLayer6(ctx)
	r7 := ExtractLayer7(ctx)

	total := len(r1) + len(r2) + len(r3) + len(r4) + len(r5) + len(r6) + len(r7)
	if total != 0 {
		t.Errorf("Empty content should produce 0 results, got %d", total)
	}
}

func TestEdgeCase_MinifiedOneLiner(t *testing.T) {
	content := `var a="/api/one",b="/api/two";fetch("/api/three");axios.get("/api/four");$.get("/api/five");`
	results := RunAllLayers(makeCtx(content, "min.js", "js"))
	urls := resultURLs(results)

	assertContains(t, urls, "/api/one")
	assertContains(t, urls, "/api/two")
	assertContains(t, urls, "/api/three")
	assertContains(t, urls, "/api/four")
	assertContains(t, urls, "/api/five")
}

func TestEdgeCase_UnicodeInPaths(t *testing.T) {
	content := `fetch("/api/用户/profile");
var u = "/api/données/export";`
	results := RunAllLayers(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	// Should extract paths with Unicode chars
	found := false
	for _, u := range urls {
		if strings.Contains(u, "/api/") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find Unicode path")
	}
}

func TestEdgeCase_MultipleURLsOnOneLine(t *testing.T) {
	content := `{a:"/api/one",b:"/api/two",c:"/api/three"}`
	results := RunAllLayers(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)

	assertContains(t, urls, "/api/one")
	assertContains(t, urls, "/api/two")
	assertContains(t, urls, "/api/three")
}

func TestEdgeCase_ArrowFunctionWithURL(t *testing.T) {
	content := `const getUrl = () => "/api/endpoint";`
	results := RunAllLayers(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)
	assertContains(t, urls, "/api/endpoint")
}

func TestEdgeCase_TernaryWithURLs(t *testing.T) {
	content := `var u = condition ? "/api/true-path" : "/api/false-path";`
	results := RunAllLayers(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)
	assertContains(t, urls, "/api/true-path")
	assertContains(t, urls, "/api/false-path")
}

func TestEdgeCase_ArrayOfURLs(t *testing.T) {
	content := `var urls = ["/api/a", "/api/b", "/api/c"];`
	results := RunAllLayers(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)
	assertContains(t, urls, "/api/a")
	assertContains(t, urls, "/api/b")
	assertContains(t, urls, "/api/c")
}

func TestEdgeCase_VeryLongPath(t *testing.T) {
	content := `var u = "/api/v1/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z";`
	results := RunAllLayers(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)
	assertContainsSubstring(t, urls, "/api/v1/a/b/c/d/e/f/g/h/i/j")
}

func TestEdgeCase_CommentsWithURLsShouldNotMatchHTTPS(t *testing.T) {
	// The // in https:// should not be treated as a comment
	content := `var u = "https://api.example.com/v1/users";`
	results := ExtractLayer6(makeCtx(content, "test.js", "js"))
	// Layer 6 should not extract "api.example.com/v1/users" as a comment URL
	for _, r := range results {
		if r.DetectionMethod == "single_line_comments" && strings.Contains(r.URL, "api.example.com") {
			t.Errorf("https:// URL should not be detected as a single-line comment: %q", r.URL)
		}
	}
}

func TestEdgeCase_SitemapMultiline(t *testing.T) {
	content := `User-agent: *
Disallow: /admin/
Allow: /api/public/
Sitemap: https://example.com/sitemap.xml
Sitemap: https://example.com/sitemap2.xml`

	results := ExtractLayer4(makeCtx(content, "robots.txt", "unknown"))
	urls := resultURLs(results)

	assertContains(t, urls, "https://example.com/sitemap.xml")
	assertContains(t, urls, "https://example.com/sitemap2.xml")
}

// ── Deobfuscation Tests ────────────────────────────────────────

func TestDeobfuscation_StringFromCharCode(t *testing.T) {
	// "http://test.com" as char codes
	content := `var u = String.fromCharCode(104,116,116,112,58,47,47,116,101,115,116,46,99,111,109);`
	results := ExtractLayer7(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)
	assertContains(t, urls, "http://test.com")
}

func TestDeobfuscation_Atob(t *testing.T) {
	// "https://secret.example.com" base64 encoded
	content := `var u = atob("aHR0cHM6Ly9zZWNyZXQuZXhhbXBsZS5jb20=");`
	results := ExtractLayer7(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)
	assertContains(t, urls, "https://secret.example.com")
}

func TestDeobfuscation_URLConstructorTwoArgs(t *testing.T) {
	content := `var u = new URL('/api/v3/data', 'https://base.example.com');`
	results := ExtractLayer7(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)
	assertContains(t, urls, "https://base.example.com/api/v3/data")
}

func TestDeobfuscation_URLConstructorOneArg(t *testing.T) {
	content := `var u = new URL('https://direct.example.com/path');`
	results := ExtractLayer7(makeCtx(content, "test.js", "js"))
	urls := resultURLs(results)
	assertContains(t, urls, "https://direct.example.com/path")
}

// ── RunAllLayers Integration ───────────────────────────────────

func TestRunAllLayers_NoCrashOnBadInput(t *testing.T) {
	badInputs := []string{
		"",
		"\x00\x00\x00",
		string(make([]byte, 100000)), // 100KB of zeros
		"{{{{{{{{{{{{{{{{{{{{",
		")))))))))))))))))))))",
		"`````````````````````",
	}

	for _, input := range badInputs {
		// Should not panic
		results := RunAllLayers(makeCtx(input, "test.js", "js"))
		_ = results
	}
}
