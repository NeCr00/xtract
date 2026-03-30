package model

import "fmt"

// Technique describes a single extraction technique.
type Technique struct {
	ID          int
	Name        string
	Layer       int
	Description string
	Enabled     bool
}

// AllTechniques returns the full list of 67 extraction techniques.
func AllTechniques() []Technique {
	return []Technique{
		{1, "absolute_urls", 1, "Absolute URLs (http://, https://, ftp://, wss://, //)", true},
		{2, "relative_paths", 1, "Relative paths (/path, ./relative, ../parent)", true},
		{3, "api_patterns", 1, "API patterns (/api/vN/, /graphql, /rest/, /rpc/, /ws/)", true},
		{4, "query_strings", 1, "Query strings and parameter extraction", true},
		{5, "hash_fragments", 1, "Hash fragments (#/route/path) for SPA routing", true},
		{6, "template_literals", 1, "Template literals with ${variable} interpolation", true},
		{7, "string_concatenation", 1, "String concatenation path assembly", true},
		{8, "data_uris", 1, "Data URIs (data:...)", true},
		{9, "blob_javascript_uris", 1, "Blob and javascript: URIs", true},
		{10, "mailto_links", 1, "Email mailto: links", true},
		{11, "ip_based_urls", 1, "IP-based URLs (IPv4 and IPv6)", true},
		{12, "encoded_urls", 1, "Encoded URLs (URL-encoded, Unicode, hex, HTML entities)", true},
		{13, "custom_protocols", 1, "Custom protocol handlers (android-app://, ios-app://, intent://)", true},
		{14, "string_literals", 1, "Path patterns inside string literals", true},
		{15, "fetch_calls", 2, "fetch() calls with URL, method, headers, body", true},
		{16, "xmlhttprequest", 2, "XMLHttpRequest .open(method, url)", true},
		{17, "axios_calls", 2, "axios.get/post/put/delete/patch/request calls", true},
		{18, "jquery_ajax", 2, "jQuery $.ajax(), $.get(), $.post(), $.getJSON()", true},
		{19, "sendbeacon", 2, "navigator.sendBeacon() calls", true},
		{20, "eventsource_websocket", 2, "new EventSource(url), new WebSocket(url)", true},
		{21, "dynamic_import", 2, "import() dynamic imports", true},
		{22, "require_calls", 2, "require() calls", true},
		{23, "location_assign", 2, "document.location / window.location assignments", true},
		{24, "window_open", 2, "window.open() calls", true},
		{25, "element_src_href", 2, "Element .src / .href property assignments", true},
		{26, "set_attribute", 2, "setAttribute('src'|'href'|'action', url)", true},
		{27, "innerhtml_urls", 2, "URLs in innerHTML/outerHTML assignments", true},
		{28, "postmessage", 2, "postMessage targetOrigin extraction", true},
		{29, "form_action", 2, "Form action URLs in JS-constructed HTML", true},
		{30, "serviceworker", 2, "Service Worker registration URLs", true},
		{31, "webworker", 2, "Web Worker / SharedWorker construction URLs", true},
		{32, "webpack_require", 2, "Webpack require.ensure / require.context", true},
		{33, "dynamic_script", 2, "Dynamic script loading (createElement + src)", true},
		{34, "react_router", 3, "React Router paths (<Route>, <Link>, useNavigate)", true},
		{35, "vue_router", 3, "Vue Router paths (routes config, router-link, $router)", true},
		{36, "angular_router", 3, "Angular Router (RouterModule, routerLink, router.navigate)", true},
		{37, "nextjs", 3, "Next.js (<Link>, useRouter, API routes)", true},
		{38, "express_routes", 3, "Express routes in client bundles (app.get/post/put/delete)", true},
		{39, "graphql_ops", 3, "GraphQL operations (queries, mutations, endpoints)", true},
		{40, "rest_inference", 3, "REST resource pattern inference (CRUD endpoints)", true},
		{41, "env_variables", 4, "Environment variable references (process.env, REACT_APP_, etc)", true},
		{42, "config_objects", 4, "Config objects (baseURL, apiUrl, endpoint, host, etc)", true},
		{43, "webpack_public_path", 4, "Webpack publicPath configuration", true},
		{44, "sourcemap_refs", 4, "Source map references (//# sourceMappingURL)", true},
		{45, "base_tags", 4, "HTML <base> tags", true},
		{46, "manifest_files", 4, "Manifest file URLs (start_url, scope, icons)", true},
		{47, "sitemap_robots", 4, "Sitemap and robots.txt references", true},
		{48, "openapi_swagger", 4, "OpenAPI/Swagger path definitions", true},
		{49, "feature_flags", 4, "Feature flags and A/B test URL patterns", true},
		{50, "subdomain_extract", 5, "Subdomain extraction from discovered URLs", true},
		{51, "cdn_asset_domains", 5, "CDN/asset domain references (S3, CloudFront, Azure, GCS)", true},
		{52, "internal_hostnames", 5, "Internal hostname patterns (*.internal, *.local, etc)", true},
		{53, "websocket_endpoints", 5, "WebSocket endpoint discovery (ws://, wss://)", true},
		{54, "single_line_comments", 6, "URLs in single-line comments (//)", true},
		{55, "multi_line_comments", 6, "URLs in multi-line comments (/* */)", true},
		{56, "html_comments", 6, "URLs in HTML comments (<!-- -->)", true},
		{57, "jsdoc_tags", 6, "JSDoc @link and @see URL references", true},
		{58, "dev_annotations", 6, "Developer annotations (// API:, // endpoint:, etc)", true},
		{59, "commented_code", 6, "URLs from commented-out code", true},
		{60, "base64_decode", 7, "Base64-encoded string detection and decoding", true},
		{61, "hex_decode", 7, "Hex-encoded string detection and decoding", true},
		{62, "unicode_decode", 7, "Unicode-escaped string decoding", true},
		{63, "array_obfuscation", 7, "Array-based string obfuscation recovery", true},
		{64, "fromcharcode", 7, "String.fromCharCode() resolution", true},
		{65, "atob_calls", 7, "atob() Base64 decoding", true},
		{66, "reverse_split_join", 7, "Reverse/split/join obfuscation patterns", true},
		{67, "url_constructor", 7, "new URL() constructor resolution", true},
	}
}

// PrintTechniques prints all techniques in a formatted table.
func PrintTechniques() {
	techniques := AllTechniques()
	currentLayer := 0
	layerNames := map[int]string{
		1: "Regex-Based Extraction",
		2: "AST-Based Extraction",
		3: "Framework-Aware Extraction",
		4: "Configuration & Metadata",
		5: "Subdomain & Infrastructure",
		6: "Comments & Developer Artifacts",
		7: "Encoded & Obfuscated Recovery",
	}
	for _, t := range techniques {
		if t.Layer != currentLayer {
			currentLayer = t.Layer
			fmt.Printf("\n  Layer %d: %s\n", currentLayer, layerNames[currentLayer])
			fmt.Printf("  %s\n", "────────────────────────────────────────────────────────────────")
		}
		status := "✓"
		if !t.Enabled {
			status = "✗"
		}
		fmt.Printf("  [%s] %2d. %-28s %s\n", status, t.ID, t.Name, t.Description)
	}
	fmt.Printf("\n  Total: %d techniques\n\n", len(techniques))
}
