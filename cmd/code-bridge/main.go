package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AI-S-Tools/code-bridge/pkg/indexer"
	"github.com/AI-S-Tools/code-bridge/pkg/parser"
	"github.com/AI-S-Tools/code-bridge/pkg/scanner"
)

const version = "0.1.1"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		cmdInit()
	case "index":
		cmdIndex()
	case "search":
		if len(os.Args) < 3 {
			fmt.Println("Usage: code-bridge search <query>")
			os.Exit(1)
		}
		cmdSearch(os.Args[2])
	case "stats":
		cmdStats()
	case "rebuild":
		cmdRebuild()
	case "rag":
		cmdRAG()
	case "version":
		fmt.Printf("code-bridge version %s\n", version)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Code-Bridge - Code indexing and search tool")
	fmt.Println("\nUsage:")
	fmt.Println("  code-bridge init         Initialize code-bridge in current directory")
	fmt.Println("  code-bridge index        Index the codebase")
	fmt.Println("  code-bridge search <q>   Search for code elements")
	fmt.Println("  code-bridge rag          List all indexed code elements (RAG format)")
	fmt.Println("  code-bridge stats        Show index statistics")
	fmt.Println("  code-bridge rebuild      Rebuild the index")
	fmt.Println("  code-bridge version      Show version")
}

func cmdInit() {
	cwd, _ := os.Getwd()
	configDir := filepath.Join(cwd, ".code-bridge")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	config := map[string]interface{}{
		"root":     cwd,
		"include":  []string{"*.go", "*.js", "*.ts", "*.py"},
		"exclude":  []string{"node_modules", ".git", "dist", "vendor"},
		"languages": []string{"go", "javascript", "typescript"},
	}

	configPath := filepath.Join(configDir, "config.json")
	data, _ := json.MarshalIndent(config, "", "  ")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Code-bridge initialized")
	fmt.Printf("  Config: %s\n", configPath)
	fmt.Printf("  Index: %s/codebase.jsonl\n", configDir)
}

func cmdIndex() {
	cwd, _ := os.Getwd()
	configDir := filepath.Join(cwd, ".code-bridge")
	indexPath := filepath.Join(configDir, "codebase.jsonl")

	fmt.Println("Scanning files...")
	s := scanner.New(cwd)
	s.LoadGitignore()
	files, err := s.Scan()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d files\n", len(files))

	goParser := parser.NewGoParser()
	idx := indexer.New(indexPath, true)

	if err := idx.Init(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	totalElements := 0
	totalFiles := 0

	fmt.Println("Parsing and indexing...")
	for _, file := range files {
		if !goParser.SupportsFile(file.Path) {
			continue
		}

		content, err := os.ReadFile(file.Path)
		if err != nil {
			fmt.Printf("  Warning: cannot read %s\n", file.RelativePath)
			continue
		}

		result, err := goParser.Parse(file.RelativePath, content)
		if err != nil {
			fmt.Printf("  Warning: cannot parse %s\n", file.RelativePath)
			continue
		}

		if len(result.Errors) > 0 {
			fmt.Printf("  Warning: %s has parse errors\n", file.RelativePath)
		}

		indexed, err := idx.Index(result.Elements)
		if err != nil {
			fmt.Printf("  Error indexing %s: %v\n", file.RelativePath, err)
			continue
		}

		totalElements += indexed
		totalFiles++

		if totalFiles%10 == 0 {
			fmt.Printf("\r  Processed: %d files, %d elements", totalFiles, totalElements)
		}
	}

	fmt.Printf("\n✓ Indexing complete\n")
	fmt.Printf("  Files processed: %d\n", totalFiles)
	fmt.Printf("  Elements indexed: %d\n", totalElements)
}

func cmdSearch(query string) {
	cwd, _ := os.Getwd()
	indexPath := filepath.Join(cwd, ".code-bridge", "codebase.jsonl")

	idx := indexer.New(indexPath, true)

	results, err := idx.Search(func(el parser.CodeElement) bool {
		lowerQuery := strings.ToLower(query)
		return strings.Contains(strings.ToLower(el.Name), lowerQuery) ||
			strings.Contains(strings.ToLower(el.Body), lowerQuery)
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Println("No results found")
		return
	}

	// Limit to 10 results
	if len(results) > 10 {
		results = results[:10]
	}

	fmt.Printf("Found %d results:\n\n", len(results))
	for _, result := range results {
		fmt.Printf("  %s %s\n", result.Type, result.Name)
		fmt.Printf("    %s:%d\n", result.File, result.Line)
		if len(result.Params) > 0 {
			params := make([]string, len(result.Params))
			for i, p := range result.Params {
				if p.Type != "" {
					params[i] = fmt.Sprintf("%s %s", p.Name, p.Type)
				} else {
					params[i] = p.Name
				}
			}
			fmt.Printf("    Parameters: %s\n", strings.Join(params, ", "))
		}
		if result.Returns != "" {
			fmt.Printf("    Returns: %s\n", result.Returns)
		}
		fmt.Println()
	}
}

func cmdStats() {
	cwd, _ := os.Getwd()
	indexPath := filepath.Join(cwd, ".code-bridge", "codebase.jsonl")

	idx := indexer.New(indexPath, true)
	stats, err := idx.GetStats()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Code-bridge Statistics\n")
	fmt.Printf("Total Elements: %d\n", stats.TotalElements)
	fmt.Printf("Total Size: %.2f KB\n\n", float64(stats.TotalSize)/1024)

	fmt.Println("By Type:")
	for typ, count := range stats.ByType {
		fmt.Printf("  %s: %d\n", typ, count)
	}

	fmt.Println("\nBy Language:")
	for lang, count := range stats.ByLanguage {
		fmt.Printf("  %s: %d\n", lang, count)
	}

	fmt.Println("\nTop Files:")
	// Sort and show top 10 files
	type fileStat struct {
		file  string
		count int
	}
	fileStats := make([]fileStat, 0, len(stats.ByFile))
	for file, count := range stats.ByFile {
		fileStats = append(fileStats, fileStat{file, count})
	}

	// Simple bubble sort (good enough for small lists)
	for i := 0; i < len(fileStats); i++ {
		for j := i + 1; j < len(fileStats); j++ {
			if fileStats[j].count > fileStats[i].count {
				fileStats[i], fileStats[j] = fileStats[j], fileStats[i]
			}
		}
	}

	limit := 10
	if len(fileStats) < limit {
		limit = len(fileStats)
	}

	for i := 0; i < limit; i++ {
		fmt.Printf("  %s: %d elements\n", fileStats[i].file, fileStats[i].count)
	}
}

func cmdRebuild() {
	cwd, _ := os.Getwd()
	indexPath := filepath.Join(cwd, ".code-bridge", "codebase.jsonl")

	idx := indexer.New(indexPath, true)
	if err := idx.Init(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Rebuilding index...")
	if err := idx.Rebuild(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Index rebuilt")

	stats, _ := idx.GetStats()
	fmt.Printf("  Total elements: %d\n", stats.TotalElements)
}

func cmdRAG() {
	cwd, _ := os.Getwd()
	indexPath := filepath.Join(cwd, ".code-bridge", "codebase.jsonl")

	idx := indexer.New(indexPath, true)

	// Get RAG index
	ragOutput, err := idx.GetRAGIndex("type")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Format output based on flag (default: compact)
	format := "compact"
	if len(os.Args) > 2 {
		format = os.Args[2]
	}

	var output string
	switch format {
	case "file":
		output = indexer.FormatRAGByFile(ragOutput)
	case "type":
		output = indexer.FormatRAGByType(ragOutput)
	case "compact":
		fallthrough
	default:
		output = indexer.FormatRAGCompact(ragOutput)
	}

	fmt.Println(output)
}
