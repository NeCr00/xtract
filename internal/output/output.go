package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/NeCr00/xtract/internal/model"
)

// FormatOutput writes results to the configured output destination in the
// requested format (JSON Lines, CSV, or plain text with optional decorations).
func FormatOutput(results []model.Result, cfg *model.Config) {
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
		w = f
	} else {
		w = os.Stdout
	}

	switch {
	case cfg.JSONOutput:
		writeJSONLines(w, results, cfg)
	case cfg.CSVOutput:
		writeCSV(w, results)
	default:
		writeText(w, results, cfg)
	}
}

// writeJSONLines writes one JSON object per line.
func writeJSONLines(w io.Writer, results []model.Result, cfg *model.Config) {
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
// Columns: url,source_file,source_line,detection_method,http_method,category,confidence
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

		// Prepend HTTP method if requested.
		if cfg.WithMethods && r.HTTPMethod != "" {
			parts = append(parts, r.HTTPMethod)
		}

		// The URL itself (sanitize newlines to prevent log injection).
		url := strings.ReplaceAll(strings.ReplaceAll(r.URL, "\n", ""), "\r", "")
		parts = append(parts, url)

		// Append parameter names if requested.
		if cfg.WithParams {
			params := collectParams(r)
			if len(params) > 0 {
				parts = append(parts, "["+strings.Join(params, ",")+"]")
			}
		}

		// Append source file and line if requested.
		if cfg.WithSource {
			parts = append(parts, fmt.Sprintf("(%s:%d)", r.SourceFile, r.SourceLine))
		}

		// Append detection method and technique ID if debug mode.
		if cfg.Debug {
			parts = append(parts, fmt.Sprintf("{%s #%d}", r.DetectionMethod, r.TechniqueID))
		}

		line := strings.Join(parts, " ")
		fmt.Fprintln(w, line)
	}
}

// collectParams gathers all parameter names from a result (query + body params).
func collectParams(r model.Result) []string {
	var params []string
	params = append(params, r.QueryParams...)
	params = append(params, r.BodyParams...)
	return params
}
