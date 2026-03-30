package model

import (
	"regexp"
	"strings"
	"sync"
)

// regexCache caches compiled regexes for performance.
var regexCache sync.Map

// GetRegex returns a compiled regex, caching it for reuse.
func GetRegex(pattern string) *regexp.Regexp {
	if v, ok := regexCache.Load(pattern); ok {
		return v.(*regexp.Regexp)
	}
	re := regexp.MustCompile(pattern)
	regexCache.Store(pattern, re)
	return re
}

// LineNumber calculates the 1-based line number of a byte offset in content.
func LineNumber(content string, offset int) int {
	line := 1
	for i := 0; i < offset && i < len(content); i++ {
		if content[i] == '\n' {
			line++
		}
	}
	return line
}

// LineIndex caches newline positions for O(log n) line number lookups.
type LineIndex struct {
	offsets []int // byte offset of each '\n'
}

// NewLineIndex builds a line index from content.
func NewLineIndex(content string) *LineIndex {
	offsets := make([]int, 0, strings.Count(content, "\n")+1)
	for i, b := range []byte(content) {
		if b == '\n' {
			offsets = append(offsets, i)
		}
	}
	return &LineIndex{offsets: offsets}
}

// Line returns the 1-based line number for the given byte offset.
func (li *LineIndex) Line(offset int) int {
	// Binary search: find how many newlines are before offset
	lo, hi := 0, len(li.offsets)
	for lo < hi {
		mid := (lo + hi) / 2
		if li.offsets[mid] < offset {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo + 1
}

// ResultSet is a concurrent-safe deduplicated collection of results.
type ResultSet struct {
	mu      sync.Mutex
	seen    map[string]bool
	results []Result
}

// NewResultSet creates a new ResultSet.
func NewResultSet() *ResultSet {
	return &ResultSet{
		seen:    make(map[string]bool),
		results: make([]Result, 0, 1024),
	}
}

// Add adds a result if its URL hasn't been seen before.
func (rs *ResultSet) Add(r Result) bool {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	key := r.URL
	if r.HTTPMethod != "" {
		key = r.HTTPMethod + " " + r.URL
	}
	if rs.seen[key] {
		return false
	}
	rs.seen[key] = true
	rs.results = append(rs.results, r)
	return true
}

// AddAll adds multiple results, returning the number of new results added.
func (rs *ResultSet) AddAll(results []Result) int {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	added := 0
	for _, r := range results {
		key := r.URL
		if r.HTTPMethod != "" {
			key = r.HTTPMethod + " " + r.URL
		}
		if !rs.seen[key] {
			rs.seen[key] = true
			rs.results = append(rs.results, r)
			added++
		}
	}
	return added
}

// Results returns all collected results.
func (rs *ResultSet) Results() []Result {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	out := make([]Result, len(rs.results))
	copy(out, rs.results)
	return out
}

// Count returns the number of unique results.
func (rs *ResultSet) Count() int {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return len(rs.results)
}
