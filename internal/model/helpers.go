package model

import (
	"regexp"
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

// AddAll adds multiple results.
func (rs *ResultSet) AddAll(results []Result) {
	for _, r := range results {
		rs.Add(r)
	}
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
