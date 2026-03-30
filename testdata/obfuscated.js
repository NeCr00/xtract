// ============================================
// Obfuscated JavaScript - Layer 7 Deobfuscation Tests
// ============================================
// This fixture tests extraction of URLs hidden behind multiple
// obfuscation techniques: fromCharCode, atob, hex, unicode,
// array-based lookups, URL constructor, string reversal, and
// nested encoding.

// -----------------------------------------------
// 1. String.fromCharCode with different URLs
// -----------------------------------------------
// "https://charcode.example.com/api/v1/users"
var _0xa1 = String.fromCharCode(104,116,116,112,115,58,47,47,99,104,97,114,99,111,100,101,46,101,120,97,109,112,108,101,46,99,111,109,47,97,112,105,47,118,49,47,117,115,101,114,115);
// "/api/v2/secret/admin"
var _0xa2 = String.fromCharCode(47,97,112,105,47,118,50,47,115,101,99,114,101,116,47,97,100,109,105,110);
// "https://charcode.example.com/internal/debug"
var _0xa3 = String.fromCharCode(104,116,116,112,115,58,47,47,99,104,97,114,99,111,100,101,46,101,120,97,109,112,108,101,46,99,111,109,47,105,110,116,101,114,110,97,108,47,100,101,98,117,103);
// "wss://charcode.example.com/ws/realtime"
var _0xa4 = String.fromCharCode(119,115,115,58,47,47,99,104,97,114,99,111,100,101,46,101,120,97,109,112,108,101,46,99,111,109,47,119,115,47,114,101,97,108,116,105,109,101);
// "/api/charcode/health"
var _0xa5 = String.fromCharCode(47,97,112,105,47,99,104,97,114,99,111,100,101,47,104,101,97,108,116,104);

// Spread with apply pattern
var _0xa6 = String.fromCharCode.apply(null, [104,116,116,112,115,58,47,47,97,112,112,108,121,46,101,120,97,109,112,108,101,46,99,111,109,47,97,112,105]);
// "https://apply.example.com/api"

// Dynamic charcode construction
var _codes = [47,97,112,105,47,100,121,110,97,109,105,99,47,99,104,97,114,99,111,100,101];
var _0xa7 = _codes.map(function(c) { return String.fromCharCode(c); }).join('');
// "/api/dynamic/charcode"

// -----------------------------------------------
// 2. atob() with valid base64-encoded URLs
// -----------------------------------------------
// "https://b64.example.com/api/v1/secret"
var _0xb1 = atob("aHR0cHM6Ly9iNjQuZXhhbXBsZS5jb20vYXBpL3YxL3NlY3JldA==");
// "/api/v2/hidden/endpoint"
var _0xb2 = atob("L2FwaS92Mi9oaWRkZW4vZW5kcG9pbnQ=");
// "https://b64.example.com/admin/panel"
var _0xb3 = atob("aHR0cHM6Ly9iNjQuZXhhbXBsZS5jb20vYWRtaW4vcGFuZWw=");
// "wss://b64.example.com/ws/events"
var _0xb4 = atob("d3NzOi8vYjY0LmV4YW1wbGUuY29tL3dzL2V2ZW50cw==");
// "/api/b64/internal/config"
var _0xb5 = atob("L2FwaS9iNjQvaW50ZXJuYWwvY29uZmln");
// "https://b64.example.com/graphql"
var _0xb6 = atob("aHR0cHM6Ly9iNjQuZXhhbXBsZS5jb20vZ3JhcGhxbA==");
// "/api/b64/users/export"
var _0xb7 = atob("L2FwaS9iNjQvdXNlcnMvZXhwb3J0");

// atob with window reference
var _0xb8 = window.atob("aHR0cHM6Ly93aW5kb3ctYXRvYi5leGFtcGxlLmNvbS9hcGk=");
// "https://window-atob.example.com/api"

// atob with variable indirection
var _encoded = "aHR0cHM6Ly9pbmRpcmVjdC5leGFtcGxlLmNvbS9zZWNyZXQ=";
var _0xb9 = atob(_encoded);
// "https://indirect.example.com/secret"

// -----------------------------------------------
// 3. Hex-encoded URLs (\xNN sequences)
// -----------------------------------------------
// "https://hex.example.com/api/v1/users"
var _0xc1 = "\x68\x74\x74\x70\x73\x3a\x2f\x2f\x68\x65\x78\x2e\x65\x78\x61\x6d\x70\x6c\x65\x2e\x63\x6f\x6d\x2f\x61\x70\x69\x2f\x76\x31\x2f\x75\x73\x65\x72\x73";
// "/api/hex/secret/admin"
var _0xc2 = "\x2f\x61\x70\x69\x2f\x68\x65\x78\x2f\x73\x65\x63\x72\x65\x74\x2f\x61\x64\x6d\x69\x6e";
// "https://hex.example.com/internal/debug/trace"
var _0xc3 = "\x68\x74\x74\x70\x73\x3a\x2f\x2f\x68\x65\x78\x2e\x65\x78\x61\x6d\x70\x6c\x65\x2e\x63\x6f\x6d\x2f\x69\x6e\x74\x65\x72\x6e\x61\x6c\x2f\x64\x65\x62\x75\x67\x2f\x74\x72\x61\x63\x65";
// "/api/hex/v2/config"
var _0xc4 = "\x2f\x61\x70\x69\x2f\x68\x65\x78\x2f\x76\x32\x2f\x63\x6f\x6e\x66\x69\x67";
// "wss://hex.example.com/ws/stream"
var _0xc5 = "\x77\x73\x73\x3a\x2f\x2f\x68\x65\x78\x2e\x65\x78\x61\x6d\x70\x6c\x65\x2e\x63\x6f\x6d\x2f\x77\x73\x2f\x73\x74\x72\x65\x61\x6d";

// Mixed hex with regular chars
var _0xc6 = "\x2f\x61pi/he\x78/mi\x78ed/pa\x74h";
// "/api/hex/mixed/path"

// -----------------------------------------------
// 4. Unicode-encoded URLs (\uNNNN sequences)
// -----------------------------------------------
// "https://uni.example.com/api/v1/data"
var _0xd1 = "\u0068\u0074\u0074\u0070\u0073\u003a\u002f\u002f\u0075\u006e\u0069\u002e\u0065\u0078\u0061\u006d\u0070\u006c\u0065\u002e\u0063\u006f\u006d\u002f\u0061\u0070\u0069\u002f\u0076\u0031\u002f\u0064\u0061\u0074\u0061";
// "/api/unicode/hidden/path"
var _0xd2 = "\u002f\u0061\u0070\u0069\u002f\u0075\u006e\u0069\u0063\u006f\u0064\u0065\u002f\u0068\u0069\u0064\u0064\u0065\u006e\u002f\u0070\u0061\u0074\u0068";
// "https://uni.example.com/admin/secret/panel"
var _0xd3 = "\u0068\u0074\u0074\u0070\u0073\u003a\u002f\u002f\u0075\u006e\u0069\u002e\u0065\u0078\u0061\u006d\u0070\u006c\u0065\u002e\u0063\u006f\u006d\u002f\u0061\u0064\u006d\u0069\u006e\u002f\u0073\u0065\u0063\u0072\u0065\u0074\u002f\u0070\u0061\u006e\u0065\u006c";
// "/api/uni/v3/export"
var _0xd4 = "\u002f\u0061\u0070\u0069\u002f\u0075\u006e\u0069\u002f\u0076\u0033\u002f\u0065\u0078\u0070\u006f\u0072\u0074";

// Mixed unicode with regular chars
var _0xd5 = "/api/\u0075\u006e\u0069/mixed/\u0070\u0061\u0074\u0068";
// "/api/uni/mixed/path"

// -----------------------------------------------
// 5. Array-based obfuscation
// -----------------------------------------------
var _0x4e2a = [
    "https://arr.example.com/api/v1/users",       // 0
    "/api/arr/admin/settings",                     // 1
    "https://arr.example.com/internal/debug",      // 2
    "/api/arr/v2/config",                          // 3
    "wss://arr.example.com/ws/live",               // 4
    "/api/arr/health",                             // 5
    "https://arr.example.com/graphql",             // 6
    "/api/arr/export/csv",                         // 7
    "https://arr.example.com/webhooks/receive",    // 8
    "/api/arr/upload/presigned",                   // 9
    "https://arr.example.com/oauth/authorize",     // 10
    "/api/arr/v3/search",                          // 11
    "https://arr.example.com/api/v1/billing",      // 12
    "/api/arr/notifications/ws",                   // 13
    "https://arr.example.com/sso/saml/callback"    // 14
];

// Array access patterns
fetch(_0x4e2a[0]);
axios.get(_0x4e2a[1]);
var adminUrl = _0x4e2a[2];
var configEndpoint = _0x4e2a[3];
new WebSocket(_0x4e2a[4]);
setInterval(function() { fetch(_0x4e2a[5]); }, 30000);

// Shuffled array with index mapping
var _0x5f3b = ["settings", "https://shuffled.example.com/api", "/api/shuffled/secret", "users", "/api/shuffled/admin"];
var _0x2c1d = function(i) { return _0x5f3b[i]; };
fetch(_0x2c1d(1));
fetch(_0x2c1d(2));
fetch(_0x2c1d(4));

// Rotated array pattern (common in obfuscators)
(function(_arr, _shift) {
    var _rotate = function() { _arr.push(_arr.shift()); };
    while (_shift--) { _rotate(); }
}(_0x4e2a, 3));

// -----------------------------------------------
// 6. new URL() constructor with two arguments
// -----------------------------------------------
var url1 = new URL("/api/url-constructor/users", "https://base.example.com");
var url2 = new URL("/api/url-constructor/admin/panel", "https://base.example.com");
var url3 = new URL("/api/url-constructor/v2/config", window.location.origin);
var url4 = new URL("https://absolute.example.com/api/v1/override");
var url5 = new URL("/api/url-constructor/search?q=test", document.baseURI);
var url6 = new URL("../relative/path/api", "https://base.example.com/current/page");
var url7 = new URL("/api/url-constructor/ws/connect", "wss://ws.example.com");

// URL with searchParams
var url8 = new URL("/api/url-constructor/filtered", "https://base.example.com");
url8.searchParams.set("page", "1");
url8.searchParams.set("limit", "50");

// URL.createObjectURL (should note the pattern, not extract blob URLs)
var blobUrl = URL.createObjectURL(new Blob(["test"]));

// -----------------------------------------------
// 7. String reverse patterns
// -----------------------------------------------
// "https://reversed.example.com/api/v1/secret" reversed
var _0xe1 = "terces/1v/ipa/moc.elpmaxe.desrever//:sptth".split("").reverse().join("");
// "/api/reverse/admin/hidden" reversed
var _0xe2 = "neddih/nimda/esrever/ipa/".split("").reverse().join("");
// "https://reversed.example.com/internal/config" reversed
var _0xe3 = "gifnoc/lanretni/moc.elpmaxe.desrever//:sptth".split("").reverse().join("");
// "wss://reversed.example.com/ws/events" reversed
var _0xe4 = "stneve/sw/moc.elpmaxe.desrever//:ssw".split("").reverse().join("");
// "/api/reverse/export/data" reversed
var _0xe5 = "atad/tropxe/esrever/ipa/".split("").reverse().join("");

// Array reverse pattern
var _0xe6 = ["moc",".",  "elpmaxe", ".", "yarra", "//:", "sptth"].reverse().join("");
// Results in: "https://array.example.com"

// Split, reverse, rejoin with different separator
var _0xe7 = "api|reverse|pipe|path".split("|").map(function(s){return "/"+s}).join("");
// "/api/reverse/pipe/path"

// -----------------------------------------------
// 8. Nested encoding (multi-layer obfuscation)
// -----------------------------------------------
// Layer 1: base64 of "https://nested.example.com/api/v1/deep"
// Layer 2: that base64 stored in a variable
// Layer 3: atob inside another function
var _layer1 = "aHR0cHM6Ly9uZXN0ZWQuZXhhbXBsZS5jb20vYXBpL3YxL2RlZXA=";
var _layer2 = (function(s) { return atob(s); })(_layer1);
// "https://nested.example.com/api/v1/deep"

// Double base64: base64(base64("https://double-b64.example.com/secret"))
// Inner base64: "aHR0cHM6Ly9kb3VibGUtYjY0LmV4YW1wbGUuY29tL3NlY3JldA=="
// Outer base64 of that:
var _double = atob(atob("YUhSMGNITTZMeTlrYjNWaWJHVXRZalkwTG1WNFlXMXdiR1V1WTI5dEwzTmxZM0psZEE9PQ=="));
// "https://double-b64.example.com/secret"

// Hex inside atob
var _hexThenB64 = atob("\x61\x48\x52\x30\x63\x48\x4d\x36\x4c\x79\x39\x6f\x5a\x58\x68\x69\x4e\x6a\x51\x75\x5a\x58\x68\x68\x62\x58\x42\x73\x5a\x53\x35\x6a\x62\x32\x30\x76\x59\x58\x42\x70");
// atob("aHR0cHM6Ly9oZXhiNjQuZXhhbXBsZS5jb20vYXBp")
// "https://hexb64.example.com/api"

// fromCharCode building a base64 string, then decoding
var _charB64 = String.fromCharCode(97,72,82,48,99,72,77,54,76,121,57,106,97,71,70,121,89,106,89,48,76,109,86,52,89,87,49,119,98,71,85,117,89,50,57,116,76,50,70,119,97,81,61,61);
var _charB64Decoded = atob(_charB64);
// atob("aHR0cHM6Ly9jaGFyYjY0LmV4YW1wbGUuY29tL2FwaQ==")
// "https://charb64.example.com/api"

// Variable indirection chain
var _step1 = "L2FwaS9jaGFpbi9zZWNyZXQ=";
var _step2 = _step1;
var _step3 = _step2;
var _step4 = atob(_step3);
// "/api/chain/secret"

// Array join then decode
var _parts = ["aHR0cHM6Ly", "9qb2luLmV4", "YW1wbGUuY2", "9tL2FwaQ=="];
var _joined = atob(_parts.join(""));
// atob("aHR0cHM6Ly9qb2luLmV4YW1wbGUuY29tL2FwaQ==")
// "https://join.example.com/api"

// XOR-based obfuscation pattern (conceptual)
var _xorKey = 42;
var _xorEncoded = [90,126,126,100,107,16,5,5,96,103,104,17,99,96,101,109,100,108,99,17,97,103,109,5,101,100,97,104,99,126];
var _xorDecoded = _xorEncoded.map(function(c) { return String.fromCharCode(c ^ _xorKey); }).join("");
// XOR each byte with 42 to reveal the URL: "https://xor.example.com/secret"

// Replace-based deobfuscation
var _obfReplace = "hZZtZZtZZpZZs://rZZeZZpZZlZZaZZcZZe.eZZxZZaZZmZZpZZlZZe.cZZoZZm/aZZpZZi".replace(/ZZ/g, "");
// "https://replace.example.com/api"

// Char substitution cipher
var _cipher = "iuuqt://djqifs.fybnqmf.dpn/bqj/w2/tfdsfu";
var _decipher = _cipher.split("").map(function(c) {
    var code = c.charCodeAt(0);
    if (code >= 97 && code <= 122) return String.fromCharCode(((code - 97 - 1 + 26) % 26) + 97);
    if (code >= 65 && code <= 90) return String.fromCharCode(((code - 65 - 1 + 26) % 26) + 65);
    return c;
}).join("");
// ROT-1 (shifted by 1): "https://cipher.example.com/api/v2/secret"

// eval-based pattern (common in malware, but also in obfuscated bundles)
eval(atob("ZmV0Y2goImh0dHBzOi8vZXZhbC5leGFtcGxlLmNvbS9hcGkvdjEvc2VjcmV0Iik="));
// eval('fetch("https://eval.example.com/api/v1/secret")')

// Function constructor pattern
var _fn = new Function("return " + atob("Imh0dHBzOi8vZm4tY29uc3RydWN0b3IuZXhhbXBsZS5jb20vYXBpIg=="))();
// new Function("return \"https://fn-constructor.example.com/api\"")()

// setTimeout/setInterval with string (eval-like)
setTimeout("fetch('/api/timeout/eval/pattern')", 1000);
setInterval("fetch('/api/interval/eval/pattern')", 5000);

// -----------------------------------------------
// 9. Bonus: Combined multi-technique obfuscation
// -----------------------------------------------
// This block combines several techniques to hide a single URL
var _multi = (function() {
    var _a = _0x4e2a[0]; // array lookup
    var _b = atob("L2FwaS9tdWx0aS9jb21iaW5lZA=="); // base64: "/api/multi/combined"
    var _c = String.fromCharCode(47,97,112,105,47,109,117,108,116,105,47,102,105,110,97,108); // "/api/multi/final"
    var _d = "\x2f\x61\x70\x69\x2f\x6d\x75\x6c\x74\x69\x2f\x68\x65\x78"; // "/api/multi/hex"
    var _e = "\u002f\u0061\u0070\u0069\u002f\u006d\u0075\u006c\u0074\u0069\u002f\u0075\u006e\u0069"; // "/api/multi/uni"
    var _f = "lanif/itlum/ipa/".split("").reverse().join(""); // "/api/multi/final" reversed
    return [_a, _b, _c, _d, _e, _f];
})();
