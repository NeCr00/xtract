<p align="center">
  <pre align="center">
  ██╗  ██╗████████╗██████╗  █████╗  ██████╗████████╗
  ╚██╗██╔╝╚══██╔══╝██╔══██╗██╔══██╗██╔════╝╚══██╔══╝
   ╚███╔╝    ██║   ██████╔╝███████║██║        ██║
   ██╔██╗    ██║   ██╔══██╗██╔══██║██║        ██║
  ██╔╝ ██╗   ██║   ██║  ██║██║  ██║╚██████╗   ██║
  ╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝   ╚═╝
  </pre>
  <p align="center"><b>URL & Endpoint Extraction for Security Testing</b></p>
  <p align="center">
    <a href="#installation">Install</a> &bull;
    <a href="#usage">Usage</a> &bull;
    <a href="#what-it-finds">What It Finds</a> &bull;
    <a href="#output">Output</a>
  </p>
</p>

---

**xtract** analyzes JavaScript, HTML, JSON, CSS, and XML files to extract every URL, endpoint, path, and parameter reference it can find. It uses 67 extraction techniques across 7 layers — from simple regex to framework-aware pattern matching and deobfuscation.

Built for penetration testers and bug bounty hunters. Single static binary, zero dependencies, parallel processing.

## Installation

```bash
go install github.com/NeCr00/xtract/cmd/xtract@latest
```

Or build from source:

```bash
git clone https://github.com/NeCr00/xtract.git
cd xtract
go build -o xtract -ldflags="-s -w" ./cmd/xtract/
```

## Usage

```bash
# Analyze local files
xtract -f app.bundle.js
xtract -f index.html -f config.json

# Analyze a directory
xtract -d ./js-files/

# Fetch and analyze a URL
xtract -u https://target.com/static/app.js

# Pipe from other tools
curl -s https://target.com | xtract -raw
cat urls.txt | xargs -I{} xtract -u {}

# Process a list of URLs
xtract -l urls.txt
```

### Filtering

```bash
# Only URLs matching a domain
xtract -f app.js -scope target.com

# Only API endpoints
xtract -f app.js -include '/api/'

# Exclude static assets
xtract -f app.js -exclude '\.(png|jpg|css|woff)$'
```

### Output Formats

```bash
# JSON Lines to stdout
xtract -f app.js -json

# CSV to stdout
xtract -f app.js -csv

# Flat URL list to stdout
xtract -f app.js -urls-only

# Write to specific file
xtract -f app.js -json -o results.json
```

## Output

By default, xtract creates an `xtract_output/` directory with results organized by category:

```
xtract_output/
├── all_urls.txt            # Every URL, one per line
├── results.json            # Full metadata (source, method, confidence)
├── api_endpoints.txt       # /api/v1/users, /graphql, etc.
├── page_routes.txt         # /dashboard, /settings, SPA routes
├── static_assets.txt       # .js, .css, .png, .woff files
├── external_services.txt   # Third-party URLs
├── internal_infra.txt      # localhost, 10.x, *.internal
├── websockets.txt          # ws:// and wss:// endpoints
├── cloud_resources.txt     # S3, CloudFront, Azure, GCS
├── source_maps.txt         # .map file references
├── inferred_endpoints.txt  # CRUD patterns inferred from REST resources
├── emails.txt              # mailto: links
└── custom_protocols.txt    # android-app://, intent://, etc.
```

Use `-oD mydir` to change the output directory name.

## What It Finds

**67 techniques** organized in 7 layers:

| Layer | What | Examples |
|-------|------|---------|
| **Regex** | URLs, paths, encoded strings, template literals | `https://`, `/api/v1/`, `%2Fencoded`, `` `/path/${id}` `` |
| **AST** | Function calls that make requests | `fetch()`, `axios.get()`, `$.ajax()`, `XMLHttpRequest` |
| **Framework** | Router definitions and route configs | React Router, Vue Router, Angular, Next.js, Express |
| **Config** | Endpoints in configuration objects | `baseURL`, `process.env.API_URL`, OpenAPI paths |
| **Infra** | Internal and cloud infrastructure | S3 buckets, CloudFront, `*.staging`, private IPs |
| **Comments** | URLs in developer comments and docs | `// API: https://...`, `@see`, `TODO` references |
| **Decode** | Obfuscated and encoded URLs | Base64, `atob()`, `String.fromCharCode()`, hex/unicode escapes |

Run `xtract --list-techniques` to see all 67 techniques.

## Supported File Types

`.js` `.mjs` `.ts` `.tsx` `.jsx` `.html` `.htm` `.json` `.xml` `.svg` `.css` `.map` `.vue` `.svelte`

## Flags

```
Input:
  -u URL       Fetch and analyze a URL (repeatable)
  -l FILE      URL list file (repeatable)
  -f FILE      Local file (repeatable)
  -d DIR       Directory, recursive (repeatable)
  -raw         Treat stdin as raw content

Output:
  -oD DIR      Output directory (default: xtract_output/)
  -o FILE      Single-file output
  -json        JSON Lines format
  -csv         CSV format
  -urls-only   Flat list to stdout

Filters:
  -scope       Only URLs containing this domain
  -include     Only URLs matching this regex
  -exclude     Exclude URLs matching this regex

Performance:
  -t N         Worker threads (default: 10)
  -timeout N   HTTP fetch timeout in seconds (default: 10)
  -max-size N  Max file size in MB (default: 100)

Display:
  -v           Verbose (per-file progress)
  -q           Quiet (suppress all stderr)
  -debug       Show detection method per URL
```


