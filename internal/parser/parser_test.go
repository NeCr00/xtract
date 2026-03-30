package parser

import (
	"testing"

	"github.com/Necr00/xtract/internal/model"
)

func makeCtx(content, fileName, fileType string) *model.ExtractionContext {
	return &model.ExtractionContext{
		Content:  content,
		FileName: fileName,
		FileType: fileType,
	}
}

func TestParseHTML(t *testing.T) {
	content := `<html>
<head>
    <link rel="stylesheet" href="/css/main.css">
    <script src="https://cdn.example.com/vendor.js"></script>
</head>
<body>
    <a href="/about">About</a>
    <img src="/images/hero.jpg">
    <form action="/api/submit" method="POST"></form>
    <iframe src="https://embed.example.com/widget"></iframe>
</body>
</html>`

	results := ParseHTML(makeCtx(content, "test.html", "html"))
	urls := resultURLs(results)

	assertContains(t, urls, "/css/main.css")
	assertContains(t, urls, "https://cdn.example.com/vendor.js")
	assertContains(t, urls, "/about")
	assertContains(t, urls, "/images/hero.jpg")
	assertContains(t, urls, "/api/submit")
	assertContains(t, urls, "https://embed.example.com/widget")
}

func TestParseCSS(t *testing.T) {
	content := `@import url('/css/reset.css');
@import "/css/variables.css";
body { background: url('/images/bg.png'); }
@font-face { src: url('/fonts/custom.woff2') format('woff2'); }`

	results := ParseCSS(makeCtx(content, "test.css", "css"))
	urls := resultURLs(results)

	assertContains(t, urls, "/css/reset.css")
	assertContains(t, urls, "/css/variables.css")
	assertContains(t, urls, "/images/bg.png")
	assertContains(t, urls, "/fonts/custom.woff2")
}

func TestParseJSON(t *testing.T) {
	content := `{
    "api_url": "https://api.example.com/v3",
    "callback": "/api/callback",
    "nested": { "endpoint": "/api/nested/path" }
}`

	results := ParseJSON(makeCtx(content, "test.json", "json"))
	urls := resultURLs(results)

	assertContains(t, urls, "https://api.example.com/v3")
	assertContains(t, urls, "/api/callback")
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
