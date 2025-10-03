package scanner

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ScannedFile represents a file found during scanning
type ScannedFile struct {
	Path         string
	RelativePath string
	Extension    string
	Size         int64
	ModifiedAt   time.Time
}

// Scanner handles recursive directory traversal and file filtering
type Scanner struct {
	rootPath       string
	includePatterns []string
	excludePatterns []string
	followSymlinks bool
}

// New creates a new Scanner instance
func New(rootPath string) *Scanner {
	return &Scanner{
		rootPath: rootPath,
		includePatterns: []string{
			"*.js", "*.ts", "*.jsx", "*.tsx",
			"*.go", "*.py", "*.java",
		},
		excludePatterns: []string{
			"node_modules", ".git", "dist", "build",
			".code-bridge", "coverage", ".next", ".nuxt",
			"target", "__pycache__", "vendor",
		},
		followSymlinks: false,
	}
}

// SetIncludePatterns sets the file patterns to include
func (s *Scanner) SetIncludePatterns(patterns []string) {
	s.includePatterns = patterns
}

// SetExcludePatterns sets the directory/file patterns to exclude
func (s *Scanner) SetExcludePatterns(patterns []string) {
	s.excludePatterns = patterns
}

// Scan recursively scans the directory and returns matching files
func (s *Scanner) Scan() ([]ScannedFile, error) {
	var files []ScannedFile

	err := filepath.WalkDir(s.rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip files/dirs we can't access
		}

		relPath, _ := filepath.Rel(s.rootPath, path)

		// Check if should be excluded
		if s.shouldExclude(relPath) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		// Handle symlinks
		if d.Type()&fs.ModeSymlink != 0 {
			if !s.followSymlinks {
				return nil
			}
			// Resolve symlink
			target, err := os.Readlink(path)
			if err != nil {
				return nil
			}
			info, err := os.Stat(target)
			if err != nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}
		}

		// Only process files
		if !d.IsDir() && s.shouldInclude(path) {
			info, err := d.Info()
			if err != nil {
				return nil
			}

			files = append(files, ScannedFile{
				Path:         path,
				RelativePath: relPath,
				Extension:    filepath.Ext(path),
				Size:         info.Size(),
				ModifiedAt:   info.ModTime(),
			})
		}

		return nil
	})

	return files, err
}

// shouldExclude checks if the path matches exclude patterns
func (s *Scanner) shouldExclude(relPath string) bool {
	for _, pattern := range s.excludePatterns {
		if strings.Contains(relPath, pattern) {
			return true
		}
	}
	return false
}

// shouldInclude checks if the file matches include patterns
func (s *Scanner) shouldInclude(path string) bool {
	if len(s.includePatterns) == 0 {
		return true
	}

	for _, pattern := range s.includePatterns {
		matched, _ := filepath.Match(pattern, filepath.Base(path))
		if matched {
			return true
		}
	}
	return false
}

// Stats returns statistics about the scan
type Stats struct {
	TotalFiles   int
	ByExtension  map[string]int
	TotalSize    int64
}

// GetStats performs a scan and returns statistics
func (s *Scanner) GetStats() (*Stats, error) {
	files, err := s.Scan()
	if err != nil {
		return nil, err
	}

	stats := &Stats{
		TotalFiles:  len(files),
		ByExtension: make(map[string]int),
		TotalSize:   0,
	}

	for _, file := range files {
		stats.ByExtension[file.Extension]++
		stats.TotalSize += file.Size
	}

	return stats, nil
}

// LoadGitignore loads patterns from .gitignore file
func (s *Scanner) LoadGitignore() error {
	gitignorePath := filepath.Join(s.rootPath, ".gitignore")

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		return nil // .gitignore doesn't exist, that's OK
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		s.excludePatterns = append(s.excludePatterns, line)
	}

	return nil
}
