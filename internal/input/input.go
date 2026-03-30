package input

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pentester/xtract/internal/model"
)

// supportedExtensions lists file extensions that the tool can process.
var supportedExtensions = map[string]bool{
	".js": true, ".mjs": true, ".ts": true, ".tsx": true, ".jsx": true,
	".html": true, ".htm": true, ".json": true, ".xml": true, ".svg": true,
	".css": true, ".map": true, ".vue": true, ".svelte": true,
}

// CollectInputs gathers all input items from CLI arguments and stdin.
func CollectInputs(cfg *model.Config) []model.InputItem {
	var items []model.InputItem

	// Process -u (individual URLs).
	for _, u := range cfg.URLs {
		u = strings.TrimSpace(u)
		if u != "" {
			items = append(items, model.InputItem{
				Type: model.InputURL,
				Path: u,
			})
		}
	}

	// Process -l (URL list files).
	for _, listFile := range cfg.URLListFiles {
		urls := readLines(listFile)
		for _, u := range urls {
			items = append(items, model.InputItem{
				Type: model.InputURL,
				Path: u,
			})
		}
	}

	// Process -f (individual files).
	for _, f := range cfg.Files {
		f = strings.TrimSpace(f)
		if f != "" {
			items = append(items, model.InputItem{
				Type:     model.InputFile,
				Path:     f,
				FileName: filepath.Base(f),
			})
		}
	}

	// Process -d (directories).
	for _, dir := range cfg.Dirs {
		dirItems := walkDirectory(dir)
		items = append(items, dirItems...)
	}

	// If no other inputs provided or -raw is set, read from stdin.
	if len(items) == 0 || cfg.RawMode {
		stdinItems := readStdin()
		items = append(items, stdinItems...)
	}

	return items
}

// readLines reads a file line by line and returns non-empty, non-comment lines.
func readLines(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: cannot open list file %s: %v\n", path, err)
		return nil
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	// Increase scanner buffer for large lines.
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: error reading list file %s: %v\n", path, err)
	}

	return lines
}

// walkDirectory recursively walks a directory and returns InputFile items
// for all files with supported extensions.
func walkDirectory(dir string) []model.InputItem {
	var items []model.InputItem

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip entries that cannot be accessed.
			return nil
		}

		// Skip hidden directories and common non-useful directories.
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" ||
				name == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if supportedExtensions[ext] {
			items = append(items, model.InputItem{
				Type:     model.InputFile,
				Path:     path,
				FileName: filepath.Base(path),
			})
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: error walking directory %s: %v\n", dir, err)
	}

	return items
}

// readStdin reads all content from stdin and returns it as an InputRaw item.
func readStdin() []model.InputItem {
	// Check if stdin has data (is not a terminal).
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil
	}

	// Only read if stdin is piped or redirected.
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: error reading stdin: %v\n", err)
		return nil
	}

	if len(data) == 0 {
		return nil
	}

	return []model.InputItem{
		{
			Type:     model.InputRaw,
			Content:  data,
			FileName: "stdin",
		},
	}
}

// FetchResult holds the response body and detected content type from a URL fetch.
type FetchResult struct {
	Body        []byte
	ContentType string // e.g. "text/html", "application/javascript"
}

// FetchURL fetches a URL with the given timeout and returns the response body
// along with the Content-Type header for file type detection.
func FetchURL(url string, timeout int) (*FetchResult, error) {
	if timeout <= 0 {
		timeout = 30
	}

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects (max 10)")
			}
			return nil
		},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid URL %q: %w", url, err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; xtract/1.0)")
	req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch failed for %q: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d for %q", resp.StatusCode, url)
	}

	// Limit reading to a reasonable size (50MB).
	const maxFetchSize = 50 * 1024 * 1024
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxFetchSize))
	if err != nil {
		return nil, fmt.Errorf("error reading response from %q: %w", url, err)
	}

	// Extract Content-Type (strip parameters like charset).
	ct := resp.Header.Get("Content-Type")
	if idx := strings.Index(ct, ";"); idx >= 0 {
		ct = strings.TrimSpace(ct[:idx])
	}

	return &FetchResult{Body: body, ContentType: strings.ToLower(ct)}, nil
}

// ReadFile reads a file from disk with a size limit and binary detection.
func ReadFile(path string, maxSizeMB int) ([]byte, error) {
	if maxSizeMB <= 0 {
		maxSizeMB = 10
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot stat file %q: %w", path, err)
	}

	maxBytes := int64(maxSizeMB) * 1024 * 1024
	if info.Size() > maxBytes {
		return nil, fmt.Errorf("file %q exceeds size limit (%d MB > %d MB)", path, info.Size()/(1024*1024), maxSizeMB)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %q: %w", path, err)
	}

	if IsBinary(data) {
		return nil, fmt.Errorf("file %q appears to be binary, skipping", path)
	}

	return data, nil
}

// DetectFileType returns the file type based on the file extension.
func DetectFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".js", ".mjs":
		return "js"
	case ".ts", ".tsx":
		return "ts"
	case ".jsx":
		return "jsx"
	case ".html", ".htm":
		return "html"
	case ".json":
		return "json"
	case ".xml":
		return "xml"
	case ".svg":
		return "svg"
	case ".css":
		return "css"
	case ".map":
		return "sourcemap"
	case ".vue":
		return "vue"
	case ".svelte":
		return "svelte"
	default:
		return "unknown"
	}
}

// DetectFileTypeFromContentType maps an HTTP Content-Type to a file type.
func DetectFileTypeFromContentType(ct string) string {
	switch {
	case strings.Contains(ct, "html"):
		return "html"
	case strings.Contains(ct, "javascript") || strings.Contains(ct, "ecmascript"):
		return "js"
	case strings.Contains(ct, "json"):
		return "json"
	case strings.Contains(ct, "svg"):
		return "svg"
	case strings.Contains(ct, "xml"):
		return "xml"
	case strings.Contains(ct, "css"):
		return "css"
	default:
		return "unknown"
	}
}

// SniffFileType attempts to detect the file type from content when other
// methods fail. Checks for common signatures at the start of the content.
func SniffFileType(content []byte) string {
	s := strings.TrimSpace(string(content[:min(len(content), 512)]))
	lower := strings.ToLower(s)
	switch {
	case strings.HasPrefix(lower, "<!doctype html") || strings.HasPrefix(lower, "<html"):
		return "html"
	case strings.HasPrefix(s, "{") || strings.HasPrefix(s, "["):
		return "json"
	case strings.HasPrefix(lower, "<?xml") || strings.HasPrefix(lower, "<svg"):
		return "xml"
	default:
		return "unknown"
	}
}


// IsBinary checks if content appears to be binary by looking for null bytes
// in the first 8000 bytes.
func IsBinary(data []byte) bool {
	checkLen := min(len(data), 8000)
	for i := 0; i < checkLen; i++ {
		if data[i] == 0 {
			return true
		}
	}
	return false
}
