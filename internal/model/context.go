package model

// ExtractionContext provides context for extraction functions.
type ExtractionContext struct {
	Content    string
	FileName   string
	FileType   string
	BaseURL    string // from <base href="..."> if found
	SourceLine int    // offset for line number calculation
}
