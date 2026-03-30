// ============================================
// Layer 4: Configuration & Metadata Test Data
// ============================================

// Technique 41: Environment variables
const apiUrl = process.env.API_URL;
const appBase = process.env.REACT_APP_API_BASE;
const nextApi = process.env.NEXT_PUBLIC_API_URL;
const viteHost = process.env.VITE_API_HOST;
const vueEndpoint = process.env.VUE_APP_ENDPOINT;
process.env.API_URL = "https://api.production.example.com";
const importMetaUrl = import.meta.env.VITE_BACKEND_URL;

// Technique 42: Config objects
const config = {
    baseURL: "https://api.example.com/v2",
    apiUrl: "/api/internal",
    endpoint: "https://endpoint.example.com/rpc",
    host: "api.example.com",
    origin: "https://origin.example.com",
    server: "https://server.example.com:8443",
    backend: "https://backend.example.com",
    gateway: "https://gateway.example.com/v1",
    proxy: "https://proxy.example.com:3128",
    webhook: "https://hooks.example.com/webhook/abc123",
    serviceUrl: "https://service.example.com/api"
};

// Technique 43: Webpack publicPath
__webpack_public_path__ = "/static/bundles/";
module.exports = {
    output: {
        publicPath: "/assets/"
    }
};

// Technique 44: Source map references
//# sourceMappingURL=app.bundle.js.map
//@ sourceMappingURL=vendor.min.js.map

// Technique 45: Base tags
// <base href="https://www.example.com/app/">

// Technique 46: Manifest data (embedded or referenced)
var manifest = {
    "start_url": "/app/home",
    "scope": "/app/",
    "icons": [
        { "src": "/icons/icon-192.png" },
        { "src": "/icons/icon-512.png" }
    ],
    "related_applications": [
        { "url": "https://play.google.com/store/apps/details?id=com.example" }
    ]
};

// Technique 47: Sitemap/robots references
var sitemapUrl = "/sitemap.xml";
var robotsUrl = "/robots.txt";

// Technique 48: OpenAPI/Swagger paths
var swaggerSpec = {
    "paths": {
        "/api/v1/pets": { "get": {}, "post": {} },
        "/api/v1/pets/{petId}": { "get": {}, "put": {}, "delete": {} },
        "/api/v1/stores": { "get": {} },
        "/api/v1/users/login": { "post": {} }
    }
};
var swaggerUi = "https://petstore.swagger.io/v2/swagger.json";

// Technique 49: Feature flags
var config2 = {
    experiment_url: "https://experiments.example.com/api/flags",
    variant_url: "/api/ab-test/variant",
    flag_endpoint: "/api/feature-flags",
    ab_test_url: "https://ab.example.com/evaluate"
};
