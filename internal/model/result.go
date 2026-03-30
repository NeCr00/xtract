package model

// Result represents a single extracted URL/endpoint with metadata.
type Result struct {
	URL             string            `json:"url"`
	SourceFile      string            `json:"source_file"`
	SourceLine      int               `json:"source_line"`
	DetectionMethod string            `json:"detection_method"`
	HTTPMethod      string            `json:"http_method,omitempty"`
	QueryParams     []string          `json:"query_params,omitempty"`
	BodyParams      []string          `json:"body_params,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
	Category        string            `json:"category"`
	Confidence      string            `json:"confidence"`
	TechniqueID     int               `json:"technique_id"`
}
