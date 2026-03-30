package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/NeCr00/xtract/internal/engine"
	"github.com/NeCr00/xtract/internal/model"
	"github.com/NeCr00/xtract/internal/output"
)

const version = "1.0.0"

// banner is the ASCII art header printed on startup.
const banner = `
  ██╗  ██╗████████╗██████╗  █████╗  ██████╗████████╗
  ╚██╗██╔╝╚══██╔══╝██╔══██╗██╔══██╗██╔════╝╚══██╔══╝
   ╚███╔╝    ██║   ██████╔╝███████║██║        ██║
   ██╔██╗    ██║   ██╔══██╗██╔══██║██║        ██║
  ██╔╝ ██╗   ██║   ██║  ██║██║  ██║╚██████╗   ██║
  ╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝   ╚═╝
`

// spinnerFrames are Unicode braille characters for the animated spinner.
var spinnerFrames = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}

// stringSlice implements flag.Value for repeatable string flags.
type stringSlice []string

func (s *stringSlice) String() string {
	return strings.Join(*s, ", ")
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	var urls stringSlice
	var urlListFiles stringSlice
	var files stringSlice
	var dirs stringSlice

	flag.Var(&urls, "u", "URL to fetch and analyze (repeatable)")
	flag.Var(&urlListFiles, "l", "File containing list of URLs (repeatable)")
	flag.Var(&files, "f", "Local file to analyze (repeatable)")
	flag.Var(&dirs, "d", "Directory to analyze recursively (repeatable)")

	rawMode := flag.Bool("raw", false, "Treat stdin as raw content")
	threads := flag.Int("t", 10, "Thread count for parallel processing")
	timeout := flag.Int("timeout", 10, "HTTP timeout in seconds")
	maxSize := flag.Int("max-size", 100, "Max file size in MB")
	verbose := flag.Bool("v", false, "Verbose output (per-file details to stderr)")
	debug := flag.Bool("debug", false, "Debug output showing technique details")
	quiet := flag.Bool("q", false, "Quiet mode: suppress all stderr output")
	outputFile := flag.String("o", "", "Output file path")
	jsonOutput := flag.Bool("json", false, "JSON Lines output")
	csvOutput := flag.Bool("csv", false, "CSV output")
	urlsOnly := flag.Bool("urls-only", false, "Only raw URLs (default behavior)")
	withParams := flag.Bool("with-params", false, "Include parameter names")
	withMethods := flag.Bool("with-methods", false, "Include HTTP methods")
	withSource := flag.Bool("with-source", false, "Include source file info")
	scope := flag.String("scope", "", "Only output URLs matching this domain")
	exclude := flag.String("exclude", "", "Exclude URLs matching regex pattern")
	include := flag.String("include", "", "Only include URLs matching regex pattern")
	listTechniques := flag.Bool("list-techniques", false, "Print all techniques and exit")

	flag.Usage = printUsage
	flag.Parse()

	// Determine if stderr is a terminal (for progress bar rendering).
	stderrIsTerminal := isTerminal(os.Stderr)

	// Print banner to stderr (unless quiet).
	if !*quiet {
		fmt.Fprintf(os.Stderr, "%s  v%s — URL & Endpoint Extractor\n\n", banner, version)
	}

	// Handle --list-techniques.
	if *listTechniques {
		model.PrintTechniques()
		os.Exit(0)
	}

	// Detect if stdin is a pipe.
	stdinIsPipe := false
	stat, err := os.Stdin.Stat()
	if err == nil {
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			stdinIsPipe = true
		}
	}

	// Check if any inputs were provided.
	hasExplicitInputs := len(urls) > 0 || len(urlListFiles) > 0 || len(files) > 0 || len(dirs) > 0
	hasInputs := hasExplicitInputs || *rawMode || stdinIsPipe

	if !hasInputs {
		printUsage()
		os.Exit(1)
	}

	// Build configuration.
	cfg := model.Config{
		URLs:           urls,
		URLListFiles:   urlListFiles,
		Files:          files,
		Dirs:           dirs,
		RawMode:        *rawMode,
		Threads:        *threads,
		Timeout:        *timeout,
		MaxSizeMB:      *maxSize,
		Verbose:        *verbose,
		Debug:          *debug,
		Quiet:          *quiet,
		ListTechniques: *listTechniques,
		OutputFile:     *outputFile,
		JSONOutput:     *jsonOutput,
		CSVOutput:      *csvOutput,
		URLsOnly:       *urlsOnly,
		WithParams:     *withParams,
		WithMethods:    *withMethods,
		WithSource:     *withSource,
		Scope:          *scope,
		Exclude:        *exclude,
		Include:        *include,
	}

	// Determine whether to show the progress bar.
	// Show progress when: not quiet, not verbose (verbose prints per-file),
	// and stderr is a terminal.
	showProgress := !cfg.Quiet && !cfg.Verbose && stderrIsTerminal

	// Set up progress display goroutine.
	stopProgress := make(chan struct{})
	progressDone := make(chan struct{})

	// latestStats holds the most recent stats pointer from the callback.
	// Written by worker goroutines (via callback), read by the ticker goroutine.
	var latestStats *engine.Stats

	progressCallback := func(stats *engine.Stats) {
		// Store the pointer; the progress goroutine reads it under the ticker.
		// This is safe because Stats uses an internal mutex for reads.
		latestStats = stats
	}

	startTime := time.Now()

	if showProgress {
		go func() {
			defer close(progressDone)
			tick := time.NewTicker(100 * time.Millisecond)
			defer tick.Stop()
			frame := 0
			for {
				select {
				case <-stopProgress:
					// Clear the progress line.
					fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", 100))
					return
				case <-tick.C:
					s := latestStats
					if s == nil {
						continue
					}
					total, processed, totalURLs, errors := s.Snapshot()

					spinner := spinnerFrames[frame%len(spinnerFrames)]
					frame++

					bar := renderProgressBar(processed, total, 20)
					pct := 0
					if total > 0 {
						pct = processed * 100 / total
					}

					line := fmt.Sprintf("  %s Processing [%s/%s] %s %d%% | %s URLs found | %s errors",
						spinner,
						formatNumber(processed),
						formatNumber(total),
						bar,
						pct,
						formatNumber(totalURLs),
						formatNumber(errors),
					)

					// Pad to clear any leftover characters from previous longer lines.
					if len(line) < 100 {
						line += strings.Repeat(" ", 100-len(line))
					}
					fmt.Fprintf(os.Stderr, "\r%s", line)
				}
			}
		}()
	} else {
		close(progressDone)
	}

	// Run the extraction engine.
	results := engine.RunEngine(&cfg, progressCallback)

	// Stop progress goroutine.
	if showProgress {
		close(stopProgress)
		<-progressDone
	}

	elapsed := time.Since(startTime)

	// Print summary table to stderr (unless quiet).
	if !cfg.Quiet {
		var filesProcessed, totalErrors int
		if latestStats != nil {
			_, filesProcessed, _, totalErrors = latestStats.Snapshot()
		}
		printSummary(filesProcessed, len(results), totalErrors, elapsed, results)
	}

	// Format and write output.
	output.FormatOutput(results, &cfg)
}

// isTerminal returns true if the given file is connected to a terminal.
func isTerminal(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// renderProgressBar builds a progress bar string of the given width using
// block characters.
func renderProgressBar(current, total, width int) string {
	if total <= 0 {
		return strings.Repeat("░", width)
	}
	filled := current * width / total
	if filled > width {
		filled = width
	}
	empty := width - filled
	return strings.Repeat("█", filled) + strings.Repeat("░", empty)
}

// formatNumber formats an integer with comma separators (e.g., 1247 -> "1,247").
func formatNumber(n int) string {
	if n < 0 {
		return "-" + formatNumber(-n)
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	// Insert commas from right to left.
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

// printSummary renders the bordered summary table to stderr.
func printSummary(filesProcessed, urlsExtracted, errors int, elapsed time.Duration, results []model.Result) {
	const boxWidth = 47 // inner width between the border chars

	// Count categories.
	catCounts := make(map[string]int)
	for _, r := range results {
		catCounts[r.Category]++
	}

	// Ordered list of categories to display.
	categoryOrder := []string{
		"api_endpoint",
		"page_route",
		"static_asset",
		"external_service",
		"internal_infra",
		"websocket",
		"cloud_resource",
		"source_map",
	}

	// Collect any categories not in the predefined list.
	otherCount := 0
	knownCats := make(map[string]bool)
	for _, c := range categoryOrder {
		knownCats[c] = true
	}
	for cat, count := range catCounts {
		if !knownCats[cat] {
			otherCount += count
		}
	}

	topBorder := "  ┌" + strings.Repeat("─", boxWidth) + "┐"
	midBorder := "  ├" + strings.Repeat("─", boxWidth) + "┤"
	botBorder := "  └" + strings.Repeat("─", boxWidth) + "┘"

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, topBorder)
	fmt.Fprintln(os.Stderr, boxLine("Extraction Complete", boxWidth, true))
	fmt.Fprintln(os.Stderr, midBorder)

	fmt.Fprintln(os.Stderr, boxKV("Files Processed", formatNumber(filesProcessed), boxWidth))
	fmt.Fprintln(os.Stderr, boxKV("URLs Extracted", formatNumber(urlsExtracted), boxWidth))
	fmt.Fprintln(os.Stderr, boxKV("Errors", formatNumber(errors), boxWidth))
	fmt.Fprintln(os.Stderr, boxKV("Time Elapsed", fmt.Sprintf("%.2fs", elapsed.Seconds()), boxWidth))

	// Only show category breakdown if there are results.
	if len(results) > 0 {
		fmt.Fprintln(os.Stderr, midBorder)
		for _, cat := range categoryOrder {
			count := catCounts[cat]
			if count > 0 {
				fmt.Fprintln(os.Stderr, boxKV(cat, formatNumber(count), boxWidth))
			}
		}
		if otherCount > 0 {
			fmt.Fprintln(os.Stderr, boxKV("other", formatNumber(otherCount), boxWidth))
		}
	}

	fmt.Fprintln(os.Stderr, botBorder)
	fmt.Fprintln(os.Stderr)
}

// boxLine renders a centered text line inside the box.
func boxLine(text string, width int, center bool) string {
	if center {
		pad := width - len(text)
		if pad < 0 {
			pad = 0
		}
		left := pad / 2
		right := pad - left
		return "  │" + strings.Repeat(" ", left) + text + strings.Repeat(" ", right) + "│"
	}
	pad := width - len(text)
	if pad < 0 {
		pad = 0
	}
	return "  │" + text + strings.Repeat(" ", pad) + "│"
}

// boxKV renders a key-value pair line inside the box with consistent alignment.
func boxKV(key, value string, width int) string {
	// Format: "  key            : value                     "
	const keyWidth = 17
	padKey := keyWidth - len(key)
	if padKey < 0 {
		padKey = 0
	}
	content := "  " + key + strings.Repeat(" ", padKey) + ": " + value
	pad := width - len(content)
	if pad < 0 {
		pad = 0
	}
	return "  │" + content + strings.Repeat(" ", pad) + "│"
}

// printUsage prints the CLI usage message with examples.
func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage:
  xtract [options] [-u URL...] [-f FILE...] [-d DIR...] [-l LIST...]

Input Options:
  -u URL          URL to fetch and analyze (repeatable)
  -l FILE         File containing list of URLs, one per line (repeatable)
  -f FILE         Local file to analyze (repeatable)
  -d DIR          Directory to analyze recursively (repeatable)
  -raw            Treat stdin as raw content

Processing Options:
  -t N            Thread count (default: 10)
  -timeout N      HTTP timeout in seconds (default: 10)
  -max-size N     Max file size in MB (default: 100)

Output Format:
  -o FILE         Write output to file
  -json           JSON Lines output (one object per line)
  -csv            CSV output with headers
  -urls-only      Only raw URLs, one per line (default)
  -with-params    Include parameter names after URL
  -with-methods   Prepend HTTP method before URL
  -with-source    Append source file and line number

Filtering:
  -scope DOMAIN   Only output URLs containing this domain
  -exclude REGEX  Exclude URLs matching this regex pattern
  -include REGEX  Only include URLs matching this regex pattern

Other:
  -v              Verbose output (per-file details to stderr)
  -q              Quiet mode (suppress all stderr output)
  -debug          Debug output showing technique details
  --list-techniques  Print all 67 extraction techniques and exit

Examples:
  # Analyze a single URL
  xtract -u https://example.com/app.js

  # Analyze multiple files
  xtract -f bundle.js -f index.html

  # Analyze a directory of downloaded assets
  xtract -d ./site-mirror/

  # Pipe content from another tool
  curl -s https://example.com | xtract -raw

  # Feed URLs from a list, output as JSON
  xtract -l urls.txt -json -o results.json

  # Filter results to a specific domain
  xtract -f app.js -scope example.com

  # Show full details with source info and methods
  xtract -u https://example.com/app.js -with-source -with-methods -with-params

  # Exclude static assets, include only API endpoints
  xtract -f bundle.js -include '/api/' -exclude '\.(css|png|jpg)$'

`)
}
