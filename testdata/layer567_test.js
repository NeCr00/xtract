// ============================================
// Layer 5: Subdomain & Infrastructure Test Data
// ============================================

// Technique 50: Subdomains
var cdn = "https://cdn.assets.example.com/bundle.js";
var staging = "https://staging.api.example.com/v1/test";
var internal = "https://internal.tools.example.com/debug";

// Technique 51: CDN/Asset domains
var s3 = "https://my-bucket.s3.amazonaws.com/uploads/file.pdf";
var s3region = "https://s3-us-west-2.amazonaws.com/bucket/key";
var cloudfront = "https://d1234567890.cloudfront.net/assets/app.js";
var azure = "https://myaccount.blob.core.windows.net/container/blob.dat";
var gcs = "https://storage.googleapis.com/my-bucket/object.json";
var firebase = "https://firebasestorage.googleapis.com/v0/b/project/o/file";

// Technique 52: Internal hostnames
var internalHost = "https://api.internal.example.com/health";
var localHost = "http://service.local:8080/status";
var stagingHost = "https://app.staging.example.com/debug";
var devHost = "http://api.dev.example.com/test";
var testHost = "https://mock.test.example.com/fixture";
var privateIp = "http://10.0.1.50:9090/metrics";
var localhost = "http://localhost:3000/dev/hot-reload";

// Technique 53: WebSocket endpoints
var wsEndpoint = "ws://realtime.example.com:8080/events";
var wssEndpoint = "wss://secure-ws.example.com/notifications";
var wsInternal = "ws://10.0.0.5:9090/internal/stream";

// ============================================
// Layer 6: Comments & Developer Artifacts
// ============================================

// Technique 54: Single-line comments with URLs
// See https://docs.example.com/api/v3/reference for API docs
// Old endpoint: /api/v1/deprecated/users
// Dev server: http://localhost:8080/dev

// Technique 55: Multi-line comments with URLs
/*
 * API Documentation: https://wiki.example.com/api-guide
 * Staging endpoint: https://staging.example.com/api/v2
 * TODO: Migrate from /api/v1/legacy to /api/v2/modern
 */

/* Debug endpoint: http://debug.example.com/trace */

// Technique 56: HTML comments (in HTML context)
<!-- This links to https://admin.example.com/panel -->
<!-- Old API: /api/v0/users -->

// Technique 57: JSDoc tags
/**
 * @link https://jsdoc.example.com/reference
 * @see https://github.com/example/repo/issues/123
 * {@link https://inline.example.com/docs}
 */

// Technique 58: Developer annotations
// API: https://api.example.com/v3/annotated
// endpoint: /api/internal/annotated
// url: https://service.example.com/annotated
// route: /admin/dashboard/annotated
// TODO: Fix the auth flow at https://auth.example.com/fix-this

// Technique 59: Commented-out code
// fetch('/api/v1/commented-out/users');
// const url = "https://old.example.com/removed-feature";
// axios.get('/api/v1/commented-out/data');

// ============================================
// Layer 7: Encoded & Obfuscated Recovery
// ============================================

// Technique 60: Base64 encoded strings
var b64 = "aHR0cHM6Ly9zZWNyZXQuZXhhbXBsZS5jb20vYXBpL2hpZGRlbg==";

// Technique 61: Hex-encoded strings
var hexStr = "\x68\x74\x74\x70\x3a\x2f\x2f\x68\x65\x78\x2e\x65\x78\x61\x6d\x70\x6c\x65\x2e\x63\x6f\x6d\x2f\x73\x65\x63\x72\x65\x74";

// Technique 62: Unicode-escaped strings
var uniStr = "\u0068\u0074\u0074\u0070\u003a\u002f\u002f\u0075\u006e\u0069\u0063\u006f\u0064\u0065\u002e\u0065\u0078\u0061\u006d\u0070\u006c\u0065\u002e\u0063\u006f\u006d\u002f\u0070\u0061\u0074\u0068";

// Technique 63: Array-based obfuscation
var _0xabc = ['/api/obfuscated/endpoint1', 'https://hidden.example.com/api', '/secret/admin/panel'];

// Technique 64: String.fromCharCode
var charUrl = String.fromCharCode(104,116,116,112,58,47,47,99,104,97,114,99,111,100,101,46,101,120,97,109,112,108,101,46,99,111,109);

// Technique 65: atob() calls
var decoded = atob("aHR0cHM6Ly9hdG9iLmV4YW1wbGUuY29tL3NlY3JldA==");

// Technique 66: Reverse/split/join
var reversed = "moc.elpmaxe.desrever//:sptth".split('').reverse().join('');

// Technique 67: URL constructor
var fullUrl = new URL('/api/v3/constructed', 'https://base.example.com');
var simpleUrl = new URL('https://simple.example.com/direct');
