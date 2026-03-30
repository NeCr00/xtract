package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Necr00/xtract/internal/engine"
	"github.com/Necr00/xtract/internal/model"
	"github.com/Necr00/xtract/internal/output"
)

const version = "1.0.0"

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
	verbose := flag.Bool("v", false, "Verbose output (progress to stderr)")
	debug := flag.Bool("debug", false, "Debug output showing technique details")
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

	// Print banner to stderr.
	fmt.Fprintf(os.Stderr, "xtract v%s — URL & Endpoint Extractor\n", version)

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

	// Run the extraction engine.
	results := engine.RunEngine(&cfg)

	// Format and write output.
	output.FormatOutput(results, &cfg)

	// Print summary to stderr if verbose.
	if cfg.Verbose {
		fmt.Fprintf(os.Stderr, "\n[summary] %d unique URLs/endpoints extracted\n", len(results))
	}
}

// printUsage prints the CLI usage message with examples.
func printUsage() {
	fmt.Fprintf(os.Stderr, `
Usage:
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
  -v              Verbose output (progress to stderr)
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
