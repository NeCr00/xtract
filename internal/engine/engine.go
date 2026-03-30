package engine

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"sync"

	"github.com/pentester/xtract/internal/extract"
	"github.com/pentester/xtract/internal/input"
	"github.com/pentester/xtract/internal/model"
	"github.com/pentester/xtract/internal/parser"
)

// RunEngine orchestrates parallel extraction across all input items.
// It collects inputs, spins up a worker pool, runs all extraction layers
// and file-type parsers, deduplicates results, applies filters, and
// returns sorted results.
func RunEngine(cfg *model.Config) []model.Result {
	items := input.CollectInputs(cfg)
	if len(items) == 0 {
		return nil
	}

	rs := model.NewResultSet()
	threads := cfg.Threads
	if threads < 1 {
		threads = 1
	}

	// Create a buffered channel for work distribution.
	work := make(chan model.InputItem, len(items))
	for _, item := range items {
		work <- item
	}
	close(work)

	var wg sync.WaitGroup
	wg.Add(threads)

	for i := 0; i < threads; i++ {
		go func() {
			defer wg.Done()
			for item := range work {
				processItem(item, cfg, rs)
			}
		}()
	}

	wg.Wait()

	results := rs.Results()
	results = filterResults(results, cfg)
	results = sortResults(results)
	return results
}

// processItem handles a single input item: fetches/reads content,
// detects file type, runs parsers and extraction layers, and adds
// results to the shared ResultSet.
func processItem(item model.InputItem, cfg *model.Config, rs *model.ResultSet) {
	var content []byte
	var err error
	var fileName string

	var contentType string // HTTP Content-Type for URL inputs

	switch item.Type {
	case model.InputURL:
		result, fetchErr := input.FetchURL(item.Path, cfg.Timeout)
		if fetchErr != nil {
			fmt.Fprintf(os.Stderr, "[error] fetching %s: %v\n", item.Path, fetchErr)
			return
		}
		content = result.Body
		contentType = result.ContentType
		if input.IsBinary(content) {
			fmt.Fprintf(os.Stderr, "[skip] %s appears to be binary\n", item.Path)
			return
		}
		fileName = item.Path

	case model.InputFile:
		content, err = input.ReadFile(item.Path, cfg.MaxSizeMB)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[error] reading %s: %v\n", item.Path, err)
			return
		}
		fileName = item.Path

	case model.InputRaw:
		content = item.Content
		fileName = item.FileName
		if fileName == "" {
			fileName = "<stdin>"
		}
	}

	if len(content) == 0 {
		return
	}

	fileType := input.DetectFileType(fileName)
	// For URL inputs where extension-based detection fails, use Content-Type
	// header or content sniffing to determine file type.
	if fileType == "unknown" && contentType != "" {
		fileType = input.DetectFileTypeFromContentType(contentType)
	}
	if fileType == "unknown" && len(content) > 0 {
		fileType = input.SniffFileType(content)
	}
	ctx := &model.ExtractionContext{
		Content:  string(content),
		FileName: fileName,
		FileType: fileType,
	}

	var localResults []model.Result

	// Run file-type specific parsers.
	switch fileType {
	case "html":
		localResults = append(localResults, parser.ParseHTML(ctx)...)
		// For HTML, extract inline script contents and process them as JS.
		scriptContexts := parser.ExtractScriptContents(ctx)
		for _, sc := range scriptContexts {
			localResults = append(localResults, extract.RunAllLayers(&sc)...)
		}
	case "css":
		localResults = append(localResults, parser.ParseCSS(ctx)...)
	case "json":
		localResults = append(localResults, parser.ParseJSON(ctx)...)
	case "xml", "svg":
		localResults = append(localResults, parser.ParseXML(ctx)...)
	case "sourcemap":
		localResults = append(localResults, parser.ParseSourceMap(ctx)...)
	case "vue", "svelte":
		localResults = append(localResults, parser.ParseVueSvelte(ctx)...)
	}

	// Run all 7 extraction layers on the main content.
	localResults = append(localResults, extract.RunAllLayers(ctx)...)

	// Add all results to the shared set.
	rs.AddAll(localResults)

	if cfg.Verbose {
		fmt.Fprintf(os.Stderr, "[%s] %d URLs found\n", fileName, len(localResults))
	}
}

// filterResults applies scope, exclude, and include filters to a result slice.
func filterResults(results []model.Result, cfg *model.Config) []model.Result {
	if cfg.Scope == "" && cfg.Exclude == "" && cfg.Include == "" {
		return results
	}

	var filtered []model.Result

	var excludeRe *regexp.Regexp
	var includeRe *regexp.Regexp

	if cfg.Exclude != "" {
		var err error
		excludeRe, err = regexp.Compile(cfg.Exclude)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[error] invalid -exclude pattern %q: %v\n", cfg.Exclude, err)
		}
	}

	if cfg.Include != "" {
		var err error
		includeRe, err = regexp.Compile(cfg.Include)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[error] invalid -include pattern %q: %v\n", cfg.Include, err)
		}
	}

	for _, r := range results {
		// Scope filter: URL must contain the domain.
		if cfg.Scope != "" {
			if !model.ContainsStr(r.URL, cfg.Scope) {
				continue
			}
		}

		// Exclude filter: remove results matching the regex.
		if excludeRe != nil {
			if excludeRe.MatchString(r.URL) {
				continue
			}
		}

		// Include filter: only keep results matching the regex.
		if includeRe != nil {
			if !includeRe.MatchString(r.URL) {
				continue
			}
		}

		filtered = append(filtered, r)
	}

	return filtered
}

// sortResults sorts results by Category first, then by URL alphabetically.
func sortResults(results []model.Result) []model.Result {
	sort.Slice(results, func(i, j int) bool {
		if results[i].Category != results[j].Category {
			return results[i].Category < results[j].Category
		}
		return results[i].URL < results[j].URL
	})
	return results
}
