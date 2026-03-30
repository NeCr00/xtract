// ============================================
// Adversarial Edge Cases for URL Extraction
// ============================================
// This fixture is designed to stress-test every edge of a URL extractor.
// It covers minified code, template literals, concatenation, Unicode,
// false positives, empty strings, obfuscation, and many more.

// -----------------------------------------------
// 1. Minified code with URLs embedded (no whitespace)
// -----------------------------------------------
var _min=function(){var a="/api/minified/endpoint";var b="https://cdn.example.com/min.js";var c=fetch("/api/v2/minified/users");return{a:a,b:b,c:c}};var _min2=function(){return fetch("https://api.example.com/v1/minified/data").then(function(r){return r.json()}).then(function(d){return d})};

// -----------------------------------------------
// 2. Deeply nested template literals
// -----------------------------------------------
var a = "users";
var b = "profile";
var c = "settings";
var nested1 = `/api/${a}/${b}/${c}`;
var nested2 = `/api/v1/${a}/${b}/${c}/preferences`;
var nested3 = `https://api.example.com/v2/${a}/${b}/${c}/${a}/${b}`;
var doubleNested = `/api/${a}/${`nested/${b}`}/${c}`;
var tripleNested = `/api/${a}/${`deep/${b}/${`deeper/${c}`}`}/end`;
var exprInTemplate = `/api/users/${1 + 2}/profile`;
var ternaryInTemplate = `/api/${true ? "admin" : "user"}/dashboard`;
var callInTemplate = `/api/${getVersion()}/resource`;

// -----------------------------------------------
// 3. URLs split across string concatenation
// -----------------------------------------------
var concatUrl1 = "http" + "s://" + "api" + ".example" + ".com";
var concatUrl2 = "http" + "s://" + "api" + ".example" + ".com" + "/v1" + "/users" + "/list";
var concatUrl3 = "/api" + "/" + "v1" + "/" + "concat" + "/" + "endpoint";
var concatUrl4 = "https://" + host + "/api/" + version + "/resource";
var concatUrl5 = "ws" + "s://" + "realtime" + ".example.com" + "/events";
var concatMultiLine = "/api/"
    + "v2/"
    + "multiline/"
    + "concatenated/"
    + "path";

// -----------------------------------------------
// 4. Unicode in URLs
// -----------------------------------------------
var unicodePath1 = "/api/用户/profile";
var unicodePath2 = "/api/données/résumé";
var unicodePath3 = "/api/ユーザー/設定";
var unicodePath4 = "https://example.com/путь/к/ресурсу";
var unicodePath5 = "/api/benutzér/Ñoño/path";
var emojiPath = "/api/users/😀/avatar";

// -----------------------------------------------
// 5. Very long paths
// -----------------------------------------------
var longPath1 = "/api/v1/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p";
var longPath2 = "/api/v1/organizations/12345/departments/67890/teams/11111/members/22222/roles/33333/permissions/44444/audit/55555";
var longPath3 = "https://deep.example.com/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z";

// -----------------------------------------------
// 6. False positive traps (NOT URLs)
// -----------------------------------------------
// Math/numeric fractions
var fraction1 = "1/2/3";
var fraction2 = "100/200/300";
var calculation = "price * 1/3";

// Boolean-like
var toggle = "true/false";
var switch1 = "on/off";
var yesno = "yes/no";
var enableDisable = "enable/disable";

// MIME types
var mime1 = "application/json";
var mime2 = "text/html";
var mime3 = "image/png";
var mime4 = "application/x-www-form-urlencoded";
var mime5 = "multipart/form-data";
var mime6 = "text/plain";
var mime7 = "application/octet-stream";
var mime8 = "video/mp4";
var mime9 = "audio/mpeg";
var mime10 = "font/woff2";

// Version strings
var ver1 = "1.2.3/4.5.6";
var ver2 = "v1.0.0/v2.0.0";
var semver = ">=1.0.0/<=2.0.0";

// Date-like
var dateLike = "2024/01/15";
var dateRange = "2024/01/01/2024/12/31";

// OS/platform strings
var platform = "win32/x64";
var arch = "linux/amd64";

// -----------------------------------------------
// 7. Empty strings assigned to URL variables
// -----------------------------------------------
fetch("");
fetch('');
axios.get("");
axios.get('');
axios.post("", {data: 1});
var apiUrl = "";
var endpoint = '';
const baseUrl = ``;
const emptyTemplate = `${""}`;
$.ajax({url: "", method: "GET"});
XMLHttpRequest.open("GET", "");

// -----------------------------------------------
// 8. URLs with fragments and query params
// -----------------------------------------------
var fqUrl1 = "/api/users?page=1&sort=name#section";
var fqUrl2 = "https://example.com/search?q=hello%20world&lang=en&limit=50#results";
var fqUrl3 = "/api/v2/items?filter[status]=active&filter[type]=premium&include=metadata";
var fqUrl4 = "/dashboard?tab=settings&subtab=security#two-factor";
var fqUrl5 = "https://api.example.com/oauth/authorize?client_id=abc&redirect_uri=https%3A%2F%2Fexample.com%2Fcallback&scope=read+write&state=xyz";
var hashOnly = "#/dashboard/admin/users";

// -----------------------------------------------
// 9. Escaped quotes in strings
// -----------------------------------------------
var escaped1 = "https://example.com/path?q=\"test\"";
var escaped2 = 'https://example.com/path?q=\'test\'';
var escaped3 = "https://example.com/search?q=\"hello world\"&page=1";
var escaped4 = "https://example.com/api?filter={\"name\":\"john\"}";
var escaped5 = 'it\'s at /api/v1/user\'s/profile';
var escaped6 = "path/to/file \"with\" quotes";
var escapedBackslash = "https://example.com\\path\\to\\resource";

// -----------------------------------------------
// 10. Multiple URLs on one line
// -----------------------------------------------
{a:"/api/one",b:"/api/two",c:"/api/three"}
var routes={home:"/",login:"/auth/login",signup:"/auth/signup",dashboard:"/dashboard",profile:"/users/profile",settings:"/settings/general",admin:"/admin/panel",api:"/api/v1",health:"/health"};
["/endpoint/a","/endpoint/b","/endpoint/c","/endpoint/d","/endpoint/e"].forEach(function(u){fetch(u)});

// -----------------------------------------------
// 11. Null/undefined/variable passed to fetch
// -----------------------------------------------
fetch(null);
fetch(undefined);
fetch(someVariable);
fetch(config.apiUrl);
fetch(getEndpoint());
fetch(window.API_BASE + "/users");
axios.get(null);
axios.post(undefined, payload);
fetch(process.env.API_URL);
fetch(import.meta.env.VITE_API_URL);

// -----------------------------------------------
// 12. Obfuscated with mixed techniques
// -----------------------------------------------
// Hex-encoded URL
var hexUrl = "\x68\x74\x74\x70\x73\x3a\x2f\x2f\x68\x65\x78\x2e\x65\x78\x61\x6d\x70\x6c\x65\x2e\x63\x6f\x6d\x2f\x73\x65\x63\x72\x65\x74";
// Unicode-encoded URL
var uniUrl = "\u0068\u0074\u0074\u0070\u0073\u003a\u002f\u002f\u0075\u006e\u0069\u002e\u0065\u0078\u0061\u006d\u0070\u006c\u0065\u002e\u0063\u006f\u006d\u002f\u0068\u0069\u0064\u0064\u0065\u006e";
// Base64-encoded URL
var b64Url = atob("aHR0cHM6Ly9iNjQuZXhhbXBsZS5jb20vc2VjcmV0L2FwaQ==");
// String.fromCharCode
var charUrl = String.fromCharCode(104,116,116,112,115,58,47,47,99,104,97,114,46,101,120,97,109,112,108,101,46,99,111,109);
// Mixed: hex + concatenation
var mixed1 = "\x2f\x61\x70\x69" + "/mixed/" + "\x65\x6e\x64\x70\x6f\x69\x6e\x74";

// -----------------------------------------------
// 13. Double-encoded URL
// -----------------------------------------------
var doubleEncoded1 = "https%253A%252F%252Fexample.com%252Fapi%252Fv1";
var doubleEncoded2 = "https%253A%252F%252Fhidden.example.com%252Fsecret";
var singleEncoded = "https%3A%2F%2Fencoded.example.com%2Fpath";
var tripleEncoded = "https%25253A%25252F%25252Ftriple.example.com";

// -----------------------------------------------
// 14. Tab and carriage return in URLs
// -----------------------------------------------
var tabUrl = "/api/v1/\tusers/\tprofile";
var crUrl = "/api/v1/\rusers/\rprofile";
var crlfUrl = "/api/v1/\r\nusers/\r\nprofile";
var mixedWhitespace = "https://example.com/path\t/with\r\n/whitespace";

// -----------------------------------------------
// 15. URLs ending with trailing punctuation
// -----------------------------------------------
// Trailing comma
var arr = ["/api/users/trailing-comma",];
var obj = {url: "/api/endpoint/trailing-comma",};
// Trailing semicolon
var semi = "/api/endpoint/trailing-semi";
// Trailing parenthesis
console.log("/api/endpoint/in-parens");
alert("Visit https://example.com/alert-url");
// Trailing bracket
var inBracket = ["/api/in/bracket"];
// Trailing angle bracket (in HTML-like contexts)
var inTag = "<a href=\"/api/link/in-tag\">";
// Trailing period (sentence context)
var sentence = "Go to https://example.com/path.";
var sentence2 = "The endpoint is /api/v1/resource.";

// -----------------------------------------------
// 16. Self-referencing URLs
// -----------------------------------------------
window.location.href = window.location.href;
window.location.assign(window.location.href);
document.location = document.location;
var selfRef = location.pathname + location.search;
var redirect = window.location.origin + "/redirect/target";

// -----------------------------------------------
// 17. Protocol-relative URLs
// -----------------------------------------------
var protoRel1 = "//cdn.example.com/script.js";
var protoRel2 = "//api.example.com/v1/data";
var protoRel3 = "//fonts.googleapis.com/css?family=Roboto";
var protoRel4 = "//maps.google.com/maps/api/js?key=ABC123";
// Without quotes (in HTML-like context or assignments)
document.write('<script src="//cdn.example.com/unquoted.js"><\/script>');

// -----------------------------------------------
// 18. Data URIs with long payloads
// -----------------------------------------------
var dataUri = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJgggAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==";
var dataUri2 = "data:application/javascript;base64,dmFyIHNlY3JldFVybCA9ICJodHRwczovL2hpZGRlbi5leGFtcGxlLmNvbS9hcGkvdjEvc2VjcmV0IjsKZmV0Y2goc2VjcmV0VXJsKTs=";
var dataUri3 = "data:text/html;base64,PGEgaHJlZj0iaHR0cHM6Ly9saW5rLmV4YW1wbGUuY29tL2hpZGRlbiI+Y2xpY2s8L2E+";

// -----------------------------------------------
// 19. JSON.parse of strings containing URLs
// -----------------------------------------------
var config1 = JSON.parse('{"apiUrl":"https://json-parsed.example.com/api/v1","wsUrl":"wss://json-parsed.example.com/ws"}');
var config2 = JSON.parse('{"endpoints":["/api/parsed/one","/api/parsed/two","/api/parsed/three"]}');
var config3 = JSON.parse("{\"redirect\":\"https://json-parsed.example.com/redirect\"}");
var nested = JSON.parse('{"a":{"b":{"c":"/api/deeply/nested/json/parsed"}}}');

// -----------------------------------------------
// 20. Arrow functions returning URLs
// -----------------------------------------------
const getUrl = () => "/api/endpoint";
const getFullUrl = () => "https://arrow.example.com/api/v1/resource";
const getConditionalUrl = (v) => v > 1 ? "/api/v2/resource" : "/api/v1/resource";
const getUrlFromEnv = () => `${process.env.API_HOST}/api/arrow/dynamic`;
const multiLineArrow = () => {
    return "/api/multiline/arrow/endpoint";
};
const arrowChain = () => fetch("/api/arrow/chained").then(r => r.json());

// -----------------------------------------------
// 21. Ternary expressions with URLs
// -----------------------------------------------
var ternUrl1 = condition ? "/api/true" : "/api/false";
var ternUrl2 = isAdmin ? "/admin/dashboard" : "/user/dashboard";
var ternUrl3 = process.env.NODE_ENV === "production"
    ? "https://prod.example.com/api"
    : "https://staging.example.com/api";
var ternUrl4 = debug ? "http://localhost:3000/api" : "https://live.example.com/api";
var nestedTern = a ? (b ? "/api/a-b" : "/api/a-notb") : (c ? "/api/nota-c" : "/api/nota-notc");

// -----------------------------------------------
// 22. Arrays of URLs
// -----------------------------------------------
var urlArray1 = ["/api/a", "/api/b", "/api/c"];
var urlArray2 = [
    "https://arr1.example.com/endpoint",
    "https://arr2.example.com/endpoint",
    "https://arr3.example.com/endpoint",
];
var mixedArray = [
    "/api/array/first",
    42,
    "/api/array/second",
    null,
    "/api/array/third",
    undefined,
    "https://array.example.com/fourth"
];
const endpoints = new Set(["/api/set/alpha", "/api/set/beta", "/api/set/gamma"]);
const urlMap = new Map([
    ["users", "/api/map/users"],
    ["posts", "/api/map/posts"],
    ["comments", "/api/map/comments"]
]);

// -----------------------------------------------
// 23. Object shorthand (no string literal - should NOT extract)
// -----------------------------------------------
var endpoint = getSomeUrl();
var config = { url: endpoint };
var opts = { method: "GET", url: endpoint, headers: headers };
var req = { endpoint, method, headers };
function makeRequest({ url, method }) { return fetch(url); }
// BUT these SHOULD extract:
var configWithLiteral = { url: "/api/literal/in/object" };
var optsWithLiteral = { method: "POST", url: "https://literal.example.com/api" };

// -----------------------------------------------
// 24. Template tags (tagged template literals)
// -----------------------------------------------
var query1 = gql`query { user(id: 1) { name email } }`;
var query2 = gql`
    mutation {
        updateUser(id: 1, input: { name: "test" }) {
            id
            name
        }
    }
`;
var styledComp = css`
    background-image: url("/assets/images/bg.png");
    font-face: url("https://fonts.example.com/roboto.woff2");
`;
var html = html`<a href="/tagged/template/link">Click</a>`;
var sql = sql`SELECT * FROM users WHERE endpoint = '/api/sql/tagged'`;

// -----------------------------------------------
// 25. Comments containing URL-like code
// -----------------------------------------------
// TODO: migrate from /api/v1/old/endpoint to /api/v2/new/endpoint
// FIXME: https://bugtracker.example.com/issues/12345
// HACK: workaround for https://github.com/org/repo/issues/999
// @see https://docs.example.com/api/reference#section
// @deprecated Use /api/v3/replacement instead
/*
 * Previous implementation used /api/legacy/removed/path
 * See: https://wiki.example.com/migration-guide
 * curl -X GET https://example.com/api/test -H "Authorization: Bearer token"
 */
// fetch("/api/commented-out/should-still-extract");
// const oldUrl = "https://commented.example.com/old";
/// Triple slash: /api/triple-slash/comment

// -----------------------------------------------
// 26. Bonus edge cases
// -----------------------------------------------

// URL in regex
var urlRegex = /https?:\/\/[^\s]+/g;
var pathRegex = /\/api\/v[0-9]+\/\w+/;

// URL in error message
throw new Error("Failed to fetch https://error.example.com/api/v1/resource");

// URL in console output
console.log("Connecting to https://log.example.com/api/ws");
console.error("502 Bad Gateway: https://log.example.com/api/broken");
console.warn("Deprecated endpoint: /api/v1/deprecated/warn");

// URL in string interpolation inside object
var logEntry = {
    message: `Request to /api/v1/interpolated/object failed`,
    url: `https://interpolated.example.com/api/${version}/resource`,
    fallback: `https://fallback.example.com/api/v1/default`
};

// Dynamic import with URL
import("/modules/dynamic/import/module.js");
const lazyModule = () => import("https://cdn.example.com/modules/lazy.js");

// Webpack magic comments in import
const AdminPanel = import(/* webpackChunkName: "admin" */ "/chunks/admin-panel.js");
const Dashboard = import(/* webpackPrefetch: true */ "/chunks/dashboard.js");

// URL in default parameter
function fetchData(url = "/api/default/parameter") { return fetch(url); }
function loadConfig(endpoint = "https://default.example.com/config") { return fetch(endpoint); }

// URL in destructuring default
const { apiUrl: myApi = "/api/destructured/default" } = config;
const [firstUrl = "/api/array/destructured/default"] = urls;

// URL in class property
class ApiClient {
    baseUrl = "https://class.example.com/api/v1";
    wsUrl = "wss://class.example.com/ws";
    healthCheck = "/health";
    static ADMIN_URL = "/admin/static/property";
}

// URL in Symbol description
const sym = Symbol("https://symbol.example.com/described");

// URL in WeakRef or FinalizationRegistry context
const cache = new Map();
cache.set("primary", "https://cache.example.com/api/primary");
cache.set("secondary", "https://cache.example.com/api/secondary");

// URL in Proxy handler
const handler = {
    get(target, prop) {
        return fetch("/api/proxy/handler/" + prop);
    }
};

// URL in generator function
function* urlGenerator() {
    yield "/api/generator/first";
    yield "/api/generator/second";
    yield "/api/generator/third";
}

// URL in async iteration
async function* streamUrls() {
    yield "https://stream.example.com/api/chunk/1";
    yield "https://stream.example.com/api/chunk/2";
    yield "https://stream.example.com/api/chunk/3";
}
