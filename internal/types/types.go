package types

import (
	"encoding/json"
	"sync"
)

type CliFlags struct {
	CleanupEmptyDirs  bool
	ConfigurationPath string
	Directory         string
	NumOfWorkers      int
	Recursive         bool
	ShowProgress      bool
}

type OrganizeOptions struct {
	ConfigPath   string
	SourcePath   string
	NumWorkers   int
	Recursive    bool
	ShowProgress bool
}

type FileJob struct {
	SourcePath string
	TargetDir  string
	Filename   string
}

// Stats tracks the progress of the file organization
type Stats struct {
	TotalFiles     int
	ProcessedFiles int
	OrganizedFiles int
	SkippedFiles   int
	mu             sync.Mutex
}

func (s *Stats) IncrementProcessed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ProcessedFiles++
}

func (s *Stats) IncrementOrganized() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.OrganizedFiles++
}

func (s *Stats) IncrementSkipped() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SkippedFiles++
}

// Config represents the structure of the JSON configuration file
type Config struct {
	// Map of folder names to lists of extensions or nested categories
	Categories map[string]json.RawMessage `json:"categories"`
}

// Category represents either a list of extensions or nested subcategories
type Category struct {
	// Extensions holds a list of file extensions if this is a leaf category
	Extensions []string
	// Subcategories holds nested categories if this is not a leaf category
	Subcategories map[string]*Category
}

// ExtensionMapping maps file extensions to their target directories
type ExtensionMapping struct {
	// Map of extension to directory path (relative to source)
	ExtToPath map[string]string
}
