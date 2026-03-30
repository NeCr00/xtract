package model

// InputItem represents a single unit of work for the extraction engine.
type InputItem struct {
	Type     InputType
	Path     string // file path or URL
	Content  []byte // raw content (for stdin/raw mode)
	FileName string // virtual filename for raw/stdin content
}

// InputType distinguishes the source of input.
type InputType int

const (
	InputFile InputType = iota
	InputURL
	InputRaw
)
