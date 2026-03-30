package output

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/NeCr00/xtract/internal/model"
)

// categoryFile maps each category constant to its output filename.
var categoryFile = map[string]string{
	model.CatAPIEndpoint:    "api_endpoints.txt",
	model.CatPageRoute:      "page_routes.txt",
	model.CatStaticAsset:    "static_assets.txt",
	model.CatExternalSvc:    "external_services.txt",
	model.CatInternalInfra:  "internal_infra.txt",
	model.CatWebSocket:      "websockets.txt",
	model.CatSourceMap:      "source_maps.txt",
	model.CatCloudResource:  "cloud_resources.txt",
	model.CatInferred:       "inferred_endpoints.txt",
	model.CatDataURI:        "data_uris.txt",
	model.CatMailto:         "emails.txt",
	model.CatCustomProtocol: "custom_protocols.txt",
}

// FormatOutput writes results to the configured output destination.
// Default: creates a directory with categorized files.
// -urls-only: flat URL list to stdout.
// -o FILE: single file output (plain/json/csv).
// -json/-csv: forces single-stream output to stdout.
func FormatOutput(results []model.Result, cfg *model.Config) {
	// -json, -csv, -o, or -urls-only → single stream output.
	if cfg.JSONOutput || cfg.CSVOutput || cfg.OutputFile != "" || cfg.URLsOnly {
		writeSingleStream(results, cfg)
		return
	}

	// Default: write categorized output directory.
	writeCategorizedDir(results, cfg)
}

// writeCategorizedDir creates an output directory with one file per category
// plus an all_urls.txt containing everything.
func writeCategorizedDir(results []model.Result, cfg *model.Config) {
	dir := cfg.OutputDir
	if dir == "" {
		dir = "xtract_output"
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "[error] cannot create output directory %q: %v\n", dir, err)
		return
	}

	// Group results by category.
	groups := make(map[string][]model.Result)
	for i := range results {
		groups[results[i].Category] = append(groups[results[i].Category], results[i])
	}

	// Write each category to its own file.
	totalWritten := 0
	for cat, items := range groups {
		fname := categoryFile[cat]
		if fname == "" {
			fname = cat + ".txt"
		}
		path := filepath.Join(dir, fname)
		n, err := writeCategoryFile(path, items, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[error] writing %s: %v\n", path, err)
			continue
		}
		totalWritten += n
	}

	// Write all_urls.txt with every URL.
	allPath := filepath.Join(dir, "all_urls.txt")
	writeAllURLs(allPath, results, cfg)

	// Write a full results.json with complete metadata.
	jsonPath := filepath.Join(dir, "results.json")
	writeFullJSON(jsonPath, results)

	if !cfg.Quiet {
		fmt.Fprintf(os.Stderr, "  Output written to: %s/\n", dir)
		fmt.Fprintf(os.Stderr, "  ├── all_urls.txt (%s URLs)\n", formatNumber(len(results)))
		fmt.Fprintf(os.Stderr, "  ├── results.json (full metadata)\n")

		// List category files with counts, sorted.
		catOrder := []string{
			model.CatAPIEndpoint, model.CatPageRoute, model.CatStaticAsset,
			model.CatExternalSvc, model.CatInternalInfra, model.CatWebSocket,
			model.CatCloudResource, model.CatSourceMap, model.CatInferred,
			model.CatDataURI, model.CatMailto, model.CatCustomProtocol,
		}
		lastIdx := -1
		for i, cat := range catOrder {
			if len(groups[cat]) > 0 {
				lastIdx = i
			}
		}
		for i, cat := range catOrder {
			items := groups[cat]
			if len(items) == 0 {
				continue
			}
			fname := categoryFile[cat]
			if fname == "" {
				fname = cat + ".txt"
			}
			prefix := "├──"
			if i == lastIdx {
				prefix = "└──"
			}
			fmt.Fprintf(os.Stderr, "  %s %s (%s)\n", prefix, fname, formatNumber(len(items)))
		}
		fmt.Fprintln(os.Stderr)
	}
}

// writeCategoryFile writes URLs from a single category to a file.
func writeCategoryFile(path string, items []model.Result, cfg *model.Config) (int, error) {
	f, err := os.Create(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	for _, r := range items {
		url := sanitizeURL(r.URL)
		if cfg.WithMethods && r.HTTPMethod != "" {
			fmt.Fprintf(w, "%s %s\n", r.HTTPMethod, url)
		} else {
			fmt.Fprintln(w, url)
		}
	}

	return len(items), nil
}

// writeAllURLs writes all URLs to a single file, one per line.
func writeAllURLs(path string, results []model.Result, _ *model.Config) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[error] writing %s: %v\n", path, err)
		return
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	for _, r := range results {
		fmt.Fprintln(w, sanitizeURL(r.URL))
	}
}

// writeFullJSON writes the complete results array as a JSON file.
func writeFullJSON(path string, results []model.Result) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[error] writing %s: %v\n", path, err)
		return
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(results); err != nil {
		fmt.Fprintf(os.Stderr, "[error] encoding JSON: %v\n", err)
	}
}

// writeSingleStream writes output to a single file or stdout (legacy mode).
func writeSingleStream(results []model.Result, cfg *model.Config) {
	var w io.Writer
	var f *os.File

	if cfg.OutputFile != "" {
		var err error
		f, err = os.Create(cfg.OutputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[error] cannot open output file %q: %v\n", cfg.OutputFile, err)
			return
		}
		defer f.Close()
		w = bufio.NewWriter(f)
		defer w.(*bufio.Writer).Flush()
	} else {
		w = bufio.NewWriter(os.Stdout)
		defer w.(*bufio.Writer).Flush()
	}

	switch {
	case cfg.JSONOutput:
		writeJSONLines(w, results)
	case cfg.CSVOutput:
		writeCSV(w, results)
	default:
		writeText(w, results, cfg)
	}
}

// writeJSONLines writes one JSON object per line.
func writeJSONLines(w io.Writer, results []model.Result) {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	for _, r := range results {
		if err := enc.Encode(r); err != nil {
			fmt.Fprintf(os.Stderr, "[error] encoding JSON: %v\n", err)
			return
		}
	}
}

// writeCSV writes results as CSV with a header row.
func writeCSV(w io.Writer, results []model.Result) {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	header := []string{"url", "source_file", "source_line", "detection_method", "http_method", "category", "confidence"}
	if err := cw.Write(header); err != nil {
		fmt.Fprintf(os.Stderr, "[error] writing CSV header: %v\n", err)
		return
	}

	for _, r := range results {
		row := []string{
			r.URL,
			r.SourceFile,
			fmt.Sprintf("%d", r.SourceLine),
			r.DetectionMethod,
			r.HTTPMethod,
			r.Category,
			r.Confidence,
		}
		if err := cw.Write(row); err != nil {
			fmt.Fprintf(os.Stderr, "[error] writing CSV row: %v\n", err)
			return
		}
	}
}

// writeText writes results as plain text lines with optional decorations.
func writeText(w io.Writer, results []model.Result, cfg *model.Config) {
	for _, r := range results {
		var parts []string

		if cfg.WithMethods && r.HTTPMethod != "" {
			parts = append(parts, r.HTTPMethod)
		}

		parts = append(parts, sanitizeURL(r.URL))

		if cfg.WithParams {
			params := collectParams(r)
			if len(params) > 0 {
				parts = append(parts, "["+strings.Join(params, ",")+"]")
			}
		}

		if cfg.WithSource {
			parts = append(parts, fmt.Sprintf("(%s:%d)", r.SourceFile, r.SourceLine))
		}

		if cfg.Debug {
			parts = append(parts, fmt.Sprintf("{%s #%d}", r.DetectionMethod, r.TechniqueID))
		}

		fmt.Fprintln(w, strings.Join(parts, " "))
	}
}

// sanitizeURL strips newlines and carriage returns from a URL.
func sanitizeURL(url string) string {
	return strings.NewReplacer("\n", "", "\r", "").Replace(url)
}

// collectParams gathers all parameter names from a result.
func collectParams(r model.Result) []string {
	var params []string
	params = append(params, r.QueryParams...)
	params = append(params, r.BodyParams...)
	return params
}

// formatNumber formats an integer with comma separators.
func formatNumber(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	remainder := len(s) % 3
	if remainder > 0 {
		b.WriteString(s[:remainder])
	}
	for i := remainder; i < len(s); i += 3 {
		if b.Len() > 0 {
			b.WriteByte(',')
		}
		b.WriteString(s[i : i+3])
	}
	return b.String()
}
