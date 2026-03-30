// ============================================
// Malformed / Broken JavaScript
// ============================================
// This fixture contains intentionally broken JS that should NOT crash
// the URL extraction tool. Valid URLs are scattered among the wreckage.

// --- UTF-8 BOM representation at file start ---
// (In a real file, the BOM bytes EF BB BF would precede this line.
//  We represent it here with a comment since the file is UTF-8 encoded
//  but many editors strip BOMs. The tool should handle BOM gracefully.)

// -----------------------------------------------
// 1. Valid URLs that SHOULD be extracted despite surrounding chaos
// -----------------------------------------------
var validUrl1 = "https://valid.example.com/api/v1/users";
var validUrl2 = "/api/v2/valid/endpoint";
var validUrl3 = "https://valid.example.com/still/works";
var validUrl4 = "/api/valid/among/chaos";
var validUrl5 = "wss://valid.example.com/ws/connection";

// -----------------------------------------------
// 2. Unclosed strings
// -----------------------------------------------
var unclosed1 = "https://unclosed.example.com/api/v1
var nextLine = "this line starts fresh";
var unclosed2 = '/api/v1/unclosed/single
var recovered = '/api/recovered/after/unclosed';
var unclosed3 = "this string never ends...
even across
multiple lines
with /api/v1/inside/unclosed/multiline somewhere
var afterUnclosed = "https://after-unclosed.example.com/api";

// -----------------------------------------------
// 3. Unclosed template literals
// -----------------------------------------------
var unclosedTpl = `/api/v1/unclosed/template/${userId
var nextTpl = `/api/v1/next/template`;
var unclosedNested = `prefix ${`nested unclosed
var recoveredTpl = `/api/recovered/template`;
var unclosedExpr = `start ${ {broken: true
var afterTpl = "https://after-template.example.com/api";

// -----------------------------------------------
// 4. Unmatched brackets and parentheses
// -----------------------------------------------
var obj = {
    url: "/api/unmatched/bracket",
    nested: {
        deep: {
            value: "/api/deeply/unmatched"
// missing closing braces intentionally

var arr = ["/api/unmatched/array", "/api/second/unmatched"
// missing closing bracket

fetch("/api/unmatched/paren"
// missing closing paren

var expr = ((("/api/triple/paren"
// missing three closing parens

function broken( {
    return "/api/broken/function/params";
}

var mixedUnmatched = {[("/api/mixed/unmatched"
// mixed unmatched brackets, parens, braces

// -----------------------------------------------
// 5. Invalid regex literals
// -----------------------------------------------
var badRegex1 = /[/;
var badRegex2 = /(/;
var badRegex3 = /(?<=/;
var urlAfterBadRegex = "/api/after/bad/regex";
var badRegex4 = /https://not-actually-a-regex.com/;
var badRegex5 = ///triple-slash-looks-like-regex;
var badRegex6 = /[^/]*/;
var anotherValid = "https://valid-after-regex.example.com/api";

// -----------------------------------------------
// 6. Binary-looking content mixed with text
// -----------------------------------------------
var binaryMix = "PK\x03\x04\x14\x00\x06\x00\x08\x00/api/inside/binary\x00\x00\x00";
var moreBinary = "\x89PNG\r\n\x1a\nhttps://inside-png-header.example.com/api\x00IHDR";
var elfHeader = "\x7fELF\x02\x01\x01\x00/api/inside/elf";
var pdfContent = "%PDF-1.4 /api/inside/pdf";
var gifHeader = "GIF89a/api/inside/gif";
var validAfterBinary = "https://valid-after-binary.example.com/endpoint";

// -----------------------------------------------
// 7. Very long lines (stress test line-by-line parsers)
// -----------------------------------------------
var longLine = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/api/in/very-long-line/endpointaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa";
var validAfterLong = "/api/valid/after/long/line";

// -----------------------------------------------
// 8. Null bytes representation
// -----------------------------------------------
// NOTE: Real null bytes would make this a binary file.
// We represent them as \x00 in string literals to test handling.
var withNulls = "https://null\x00bytes\x00.example.com/api";
var nullPadded = "\x00\x00\x00/api/null/padded\x00\x00\x00";
var nullTerminated = "/api/null/terminated\x00";
var validAfterNulls = "https://valid-after-nulls.example.com/api";

// -----------------------------------------------
// 9. Deeply nested brackets
// -----------------------------------------------
var deep = {{{{{{{{{{"/api/deep/nested/brackets"}}}}}}}}}};
var deeper = [[[[[[[[[["/api/deeper/nested/arrays"]]]]]]]]]];
var mixed = {[{[{[{[{["/api/mixed/nested/chaos"]}]}]}]}]};
var deepParens = (((((((((("/api/deep/nested/parens"))))))))))
var validAfterNesting = "https://valid-after-nesting.example.com/api";

// -----------------------------------------------
// 10. Extremely long variable names
// -----------------------------------------------
var thisIsAnExtremelyLongVariableNameThatGoesOnAndOnAndOnAndOnAndOnAndOnAndOnAndOnAndOnAndOnAndOnAndOnAndOnAndOnAndOnForever = "/api/long-var-name/endpoint";
var anotherVeryVeryVeryVeryVeryVeryVeryVeryVeryVeryVeryVeryVeryVeryVeryVeryVeryVeryVeryVeryVeryVeryVeryVeryVeryVeryVeryLongName = "https://long-varname.example.com/api/v1";
var SCREAMING_SNAKE_CASE_VERY_LONG_CONSTANT_NAME_THAT_KEEPS_GOING_AND_GOING_AND_GOING_AND_GOING_AND_GOING = "/api/screaming/constant";

// -----------------------------------------------
// 11. Syntax errors and nonsense
// -----------------------------------------------
function { }
class extends { }
var = ;
if () { } else if () { } else { }
switch () { case : break; default: }
for (;;; { }
while { }
do { } while
try { } catch { } finally { }
var url_in_chaos = "https://chaos.example.com/api/v1/endpoint";

return return return;
break break break;
continue continue continue;
throw throw throw;

var 123invalid = "/api/invalid/varname";
var #hashtag = "/api/hashtag/varname";
var @at = "/api/at/varname";

// -----------------------------------------------
// 12. Mixed encoding chaos
// -----------------------------------------------
var mixedEncoding = "https://\x6d\x69\x78\x65\x64.example.com/\u0061\u0070\u0069";
var doubleEscape = "https:\\\\example.com\\\\api\\\\v1\\\\resource";
var unicodeSurrogate = "https://surrogate.example.com/\uD800\uDC00/path";
var overlong = "/api/overlong/\xC0\xAF";

// -----------------------------------------------
// 13. Repeated patterns (stress deduplication)
// -----------------------------------------------
var dup1 = "/api/deduplicate/this";
var dup2 = "/api/deduplicate/this";
var dup3 = "/api/deduplicate/this";
var dup4 = "/api/deduplicate/this";
var dup5 = "/api/deduplicate/this";
fetch("/api/deduplicate/this");
axios.get("/api/deduplicate/this");
$.get("/api/deduplicate/this");

// -----------------------------------------------
// 14. JavaScript keywords as path segments
// -----------------------------------------------
var kwUrl1 = "/api/class/extends/super";
var kwUrl2 = "/api/import/export/default";
var kwUrl3 = "/api/async/await/yield";
var kwUrl4 = "/api/const/let/var";
var kwUrl5 = "/api/try/catch/finally";
var kwUrl6 = "/api/null/undefined/void";
var kwUrl7 = "/api/new/delete/typeof";
var kwUrl8 = "/api/return/throw/break";

// -----------------------------------------------
// 15. HTML entities mixed into JS strings
// -----------------------------------------------
var htmlEntity1 = "https://entity.example.com/api?q=hello&amp;world";
var htmlEntity2 = "/api/entity?filter=a&lt;b&gt;c";
var htmlEntity3 = "https://entity.example.com/path?name=O&apos;Brien";
var htmlEntity4 = "https://entity.example.com/search?q=&quot;test&quot;";

// -----------------------------------------------
// 16. Premature EOF simulation
// -----------------------------------------------
var lastValid = "https://last-valid.example.com/api/v1/final";
var finalPath = "/api/final/path/before/eof";
// The file just ends abruptly, mid-expression:
var abrupt = "https://abrupt-eof.example.com/
