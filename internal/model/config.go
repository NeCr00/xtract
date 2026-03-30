package model

// Config holds all CLI configuration.
type Config struct {
	URLs           []string
	URLListFiles   []string
	Files          []string
	Dirs           []string
	RawMode        bool
	Threads        int
	Timeout        int
	MaxSizeMB      int
	Verbose        bool
	Debug          bool
	ListTechniques bool
	OutputFile     string
	JSONOutput     bool
	CSVOutput      bool
	URLsOnly       bool
	WithParams     bool
	WithMethods    bool
	WithSource     bool
	Scope          string
	Exclude        string
	Include        string
}
