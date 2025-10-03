package indexer

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/AI-S-Tools/code-bridge/pkg/parser"
)

// Indexer handles JSONL index operations
type Indexer struct {
	indexPath     string
	deduplication bool
	hashSet       map[string]bool
	mu            sync.RWMutex
}

// New creates a new Indexer instance
func New(indexPath string, dedup bool) *Indexer {
	return &Indexer{
		indexPath:     indexPath,
		deduplication: dedup,
		hashSet:       make(map[string]bool),
	}
}

// Init initializes the indexer (creates directory, loads existing hashes)
func (idx *Indexer) Init() error {
	dir := filepath.Dir(idx.indexPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if idx.deduplication {
		return idx.loadExistingHashes()
	}

	return nil
}

// Index adds elements to the index
func (idx *Indexer) Index(elements []parser.CodeElement) (int, error) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	toWrite := make([]parser.CodeElement, 0)

	for _, element := range elements {
		if idx.deduplication && idx.hashSet[element.Hash] {
			continue // Skip duplicates
		}

		toWrite = append(toWrite, element)
		idx.hashSet[element.Hash] = true
	}

	if len(toWrite) > 0 {
		if err := idx.appendToIndex(toWrite); err != nil {
			return 0, err
		}
	}

	return len(toWrite), nil
}

// appendToIndex appends elements to JSONL file
func (idx *Indexer) appendToIndex(elements []parser.CodeElement) error {
	file, err := os.OpenFile(idx.indexPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, element := range elements {
		if err := encoder.Encode(element); err != nil {
			return err
		}
	}

	return nil
}

// ReadAll reads all elements from the index
func (idx *Indexer) ReadAll() ([]parser.CodeElement, error) {
	file, err := os.Open(idx.indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []parser.CodeElement{}, nil
		}
		return nil, err
	}
	defer file.Close()

	elements := make([]parser.CodeElement, 0)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var element parser.CodeElement
		if err := json.Unmarshal(scanner.Bytes(), &element); err != nil {
			continue // Skip malformed lines
		}
		elements = append(elements, element)
	}

	return elements, scanner.Err()
}

// Search searches elements by predicate
func (idx *Indexer) Search(predicate func(parser.CodeElement) bool) ([]parser.CodeElement, error) {
	file, err := os.Open(idx.indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []parser.CodeElement{}, nil
		}
		return nil, err
	}
	defer file.Close()

	results := make([]parser.CodeElement, 0)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var element parser.CodeElement
		if err := json.Unmarshal(scanner.Bytes(), &element); err != nil {
			continue
		}
		if predicate(element) {
			results = append(results, element)
		}
	}

	return results, scanner.Err()
}

// FindByName finds elements by name
func (idx *Indexer) FindByName(name string) ([]parser.CodeElement, error) {
	return idx.Search(func(el parser.CodeElement) bool {
		return el.Name == name
	})
}

// FindByType finds elements by type
func (idx *Indexer) FindByType(elemType parser.ElementType) ([]parser.CodeElement, error) {
	return idx.Search(func(el parser.CodeElement) bool {
		return el.Type == elemType
	})
}

// FindByFile finds elements by file path
func (idx *Indexer) FindByFile(filePath string) ([]parser.CodeElement, error) {
	return idx.Search(func(el parser.CodeElement) bool {
		return el.File == filePath
	})
}

// Exists checks if element exists by hash
func (idx *Indexer) Exists(hash string) bool {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.hashSet[hash]
}

// Stats represents index statistics
type Stats struct {
	TotalElements int
	ByType        map[parser.ElementType]int
	ByLanguage    map[string]int
	ByFile        map[string]int
	TotalSize     int64
}

// GetStats returns index statistics
func (idx *Indexer) GetStats() (*Stats, error) {
	elements, err := idx.ReadAll()
	if err != nil {
		return nil, err
	}

	stats := &Stats{
		TotalElements: len(elements),
		ByType:        make(map[parser.ElementType]int),
		ByLanguage:    make(map[string]int),
		ByFile:        make(map[string]int),
		TotalSize:     0,
	}

	for _, el := range elements {
		stats.ByType[el.Type]++
		stats.ByLanguage[el.Language]++
		stats.ByFile[el.File]++
		stats.TotalSize += int64(len(el.Body))
	}

	return stats, nil
}

// Clear clears the index
func (idx *Indexer) Clear() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.hashSet = make(map[string]bool)

	if _, err := os.Stat(idx.indexPath); err == nil {
		return os.Remove(idx.indexPath)
	}

	return nil
}

// Rebuild rebuilds the index (removes duplicates)
func (idx *Indexer) Rebuild() error {
	elements, err := idx.ReadAll()
	if err != nil {
		return err
	}

	// Create unique map
	unique := make(map[string]parser.CodeElement)
	for _, el := range elements {
		unique[el.Hash] = el
	}

	// Clear and rewrite
	if err := idx.Clear(); err != nil {
		return err
	}

	if err := idx.Init(); err != nil {
		return err
	}

	uniqueElements := make([]parser.CodeElement, 0, len(unique))
	for _, el := range unique {
		uniqueElements = append(uniqueElements, el)
	}

	_, err = idx.Index(uniqueElements)
	return err
}

// loadExistingHashes loads existing hashes for deduplication
func (idx *Indexer) loadExistingHashes() error {
	file, err := os.Open(idx.indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var element parser.CodeElement
		if err := json.Unmarshal(scanner.Bytes(), &element); err != nil {
			continue
		}
		idx.hashSet[element.Hash] = true
	}

	return scanner.Err()
}
