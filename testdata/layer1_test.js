// ============================================
// Layer 1: Regex-Based Extraction Test Data
// ============================================

// Technique 1: Absolute URLs
var url1 = "https://api.example.com/v1/users";
var url2 = "http://legacy.example.com/old-api";
var url3 = "ftp://files.example.com/downloads/report.pdf";
var url4 = "wss://ws.example.com/realtime";
var url5 = "ws://ws.example.com/socket";
var url6 = "//cdn.example.com/assets/bundle.js";

// Technique 2: Relative paths
var path1 = "/api/v1/users";
var path2 = "./components/Header";
var path3 = "../utils/helpers";
var path4 = "/dashboard/settings/profile";

// Technique 3: API patterns
var api1 = "/api/v1/users/list";
var api2 = "/api/v2/admin/settings";
var api3 = "/graphql";
var api4 = "/rest/services/data";
var api5 = "/rpc/execute";
var api6 = "/ws/notifications";

// Technique 4: Query strings
var qs1 = "https://api.example.com/search?q=test&page=1&limit=20";
var qs2 = "/api/users?sort=name&order=asc&filter=active";

// Technique 5: Hash fragments (SPA routing)
var hash1 = "#/dashboard";
var hash2 = "#/users/profile/edit";
var hash3 = "#/settings/notifications";

// Technique 6: Template literals
var userId = 123;
var tpl1 = `/api/v1/users/${userId}/profile`;
var tpl2 = `https://api.example.com/v2/${service}/endpoint`;
var tpl3 = `/dashboard/${orgId}/settings/${tab}`;

// Technique 7: String concatenation
var concat1 = "/api/" + "users/" + userId;
var concat2 = "https://api.example.com" + "/v1/" + "accounts";
var concat3 = baseUrl + "/search?q=" + query;

// Technique 8: Data URIs
var dataUri = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUg==";
var dataUri2 = "data:application/json;charset=utf-8,{\"key\":\"value\"}";

// Technique 9: Blob/JavaScript URIs
var blob1 = "blob:https://example.com/abc-123-def";
var jsUri = "javascript:void(0)";

// Technique 10: Mailto links
var email1 = "mailto:admin@example.com";
var email2 = "mailto:support@example.com?subject=Bug%20Report";

// Technique 11: IP-based URLs
var ip1 = "http://192.168.1.100:8080/api";
var ip2 = "https://10.0.0.1/internal/debug";
var ip3 = "http://127.0.0.1:3000/dev";
var ip4 = "http://172.16.0.50:9090/metrics";

// Technique 12: Encoded URLs
var encoded1 = "https%3A%2F%2Fapi.example.com%2Fv1%2Fsecret";
var encoded2 = "\u0068\u0074\u0074\u0070\u003a\u002f\u002fhidden.example.com/path";
var encoded3 = "\x68\x74\x74\x70\x3a\x2f\x2fhex.example.com/path";
var encoded4 = "https:&#x2F;&#x2F;entity.example.com&#x2F;path";

// Technique 13: Custom protocols
var custom1 = "android-app://com.example.app/deep/link";
var custom2 = "ios-app://123456789/path/to/content";
var custom3 = "intent://scan/#Intent;scheme=zxing;end";
var custom4 = "deeplink://app/screen/detail";

// Technique 14: String literals with paths
var str1 = "/static/images/logo.png";
var str2 = "/assets/fonts/roboto.woff2";
var str3 = "/public/docs/terms-of-service";
