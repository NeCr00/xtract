package model

// Category constants.
const (
	CatAPIEndpoint    = "api_endpoint"
	CatPageRoute      = "page_route"
	CatStaticAsset    = "static_asset"
	CatExternalSvc    = "external_service"
	CatInternalInfra  = "internal_infra"
	CatWebSocket      = "websocket"
	CatSourceMap      = "source_map"
	CatCloudResource  = "cloud_resource"
	CatInferred       = "inferred"
	CatDataURI        = "data_uri"
	CatMailto         = "mailto"
	CatCustomProtocol = "custom_protocol"
)

// Confidence constants.
const (
	ConfHigh   = "high"
	ConfMedium = "medium"
	ConfLow    = "low"
)

// staticAssetExts lists file extensions considered static assets.
var staticAssetExts = map[string]bool{
	".js": true, ".mjs": true, ".css": true, ".png": true, ".jpg": true,
	".jpeg": true, ".gif": true, ".svg": true, ".ico": true, ".woff": true,
	".woff2": true, ".ttf": true, ".eot": true, ".otf": true, ".mp4": true,
	".webm": true, ".mp3": true, ".ogg": true, ".pdf": true, ".map": true,
	".webp": true, ".avif": true,
}

// CategorizeURL assigns a category to a URL based on its pattern.
func CategorizeURL(url string) string {
	if len(url) >= 4 && (url[:3] == "ws:" || url[:4] == "wss:") {
		return CatWebSocket
	}
	if len(url) > 4 && url[len(url)-4:] == ".map" {
		return CatSourceMap
	}
	if len(url) >= 5 && url[:5] == "data:" {
		return CatDataURI
	}
	if len(url) >= 7 && url[:7] == "mailto:" {
		return CatMailto
	}
	for _, proto := range []string{"android-app://", "ios-app://", "intent://", "deeplink://"} {
		if len(url) >= len(proto) && url[:len(proto)] == proto {
			return CatCustomProtocol
		}
	}
	cloudPatterns := []string{
		"s3.amazonaws.com", "s3-", "cloudfront.net",
		"blob.core.windows.net", "storage.googleapis.com",
		"firebasestorage.googleapis.com",
	}
	for _, p := range cloudPatterns {
		if ContainsStr(url, p) {
			return CatCloudResource
		}
	}
	internalPatterns := []string{
		".internal", ".local", ".staging", ".dev", ".test",
		"localhost", "127.0.0.1", "10.", "172.16.", "172.17.",
		"172.18.", "172.19.", "172.20.", "172.21.", "172.22.",
		"172.23.", "172.24.", "172.25.", "172.26.", "172.27.",
		"172.28.", "172.29.", "172.30.", "172.31.",
		"192.168.",
	}
	for _, p := range internalPatterns {
		if ContainsStr(url, p) {
			return CatInternalInfra
		}
	}
	apiPatterns := GetRegex(`(?i)(/api/|/graphql|/rest/|/rpc/|/v\d+/)`)
	if apiPatterns.MatchString(url) {
		return CatAPIEndpoint
	}
	for ext := range staticAssetExts {
		if len(url) > len(ext) && url[len(url)-len(ext):] == ext {
			return CatStaticAsset
		}
	}
	if len(url) >= 8 && (url[:7] == "http://" || url[:8] == "https://") {
		return CatExternalSvc
	}
	return CatPageRoute
}

// ContainsStr checks if s contains substr.
func ContainsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
