# Code-Bridge

[![Release](https://img.shields.io/github/v/release/AI-S-Tools/code-bridge)](https://github.com/AI-S-Tools/code-bridge/releases)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Code indexing and search tool for Go codebases using JSONL format.

**One-line install (Linux/macOS):**
```bash
curl -sSL https://raw.githubusercontent.com/AI-S-Tools/code-bridge/master/install.sh | bash
```

## Overview

Code-Bridge scans your codebase, extracts functions and code elements, and stores them in a searchable JSONL index.

**Current Status (v0.1.0):** Go language support only. Multi-language support coming in future releases.

## Features

- 🔍 **Code Indexing**: Recursively scan and index Go codebases
- 🎯 **Precise References**: File and line number references for every code element
- 📝 **JSONL Format**: Fast, streamable, line-oriented format
- 🔧 **Go Parser**: Full support for functions, methods, structs, interfaces, types
- 🔎 **Search**: Find code by name or content
- 📊 **Statistics**: Overview by type, language, and file
- ♻️ **Deduplication**: Hash-based automatic duplicate detection

### Coming Soon (Phase 2+)
- JavaScript/TypeScript parser
- Python parser
- RAG/semantic search
- Annotations system

## Installation

### Quick Install (Linux/macOS)

```bash
curl -sSL https://raw.githubusercontent.com/AI-S-Tools/code-bridge/master/install.sh | bash
```

### Using Go

```bash
go install github.com/AI-S-Tools/code-bridge/cmd/code-bridge@latest
```

### From Source

```bash
git clone https://github.com/AI-S-Tools/code-bridge.git
cd code-bridge
make install
```

Or manually:

```bash
go build -o code-bridge ./cmd/code-bridge
sudo mv code-bridge /usr/local/bin/
```

## Quick Start

```bash
# Initialize in your project
code-bridge init

# Index your codebase
code-bridge index

# Search for code
code-bridge search "handler"

# Show statistics
code-bridge stats

# Rebuild index (remove duplicates)
code-bridge rebuild
```

### Example Output

```bash
$ code-bridge index
Scanning files...
Found 1 files
Parsing and indexing...
✓ Indexing complete
  Files processed: 1
  Elements indexed: 20

$ code-bridge stats
Code-bridge Statistics

Total Elements: 20
Total Size: 19.37 KB

By Type:
  function: 20

By Language:
  go: 20

$ code-bridge search "generateCommitMessage"
Found 3 results:

  function generateQwenCommitMessage
    main.go:221
    Parameters: gitDiff string
    Returns: (string, error)

  function generateClaudeCommitMessage
    main.go:291
    Parameters: gitDiff string
    Returns: (string, error)
...
```

## Project Status

✅ **v0.1.0 Released** - Go language support

**What Works Now:**
- ✅ Go language parsing (functions, methods, structs, interfaces, types)
- ✅ JSONL indexing with hash-based deduplication
- ✅ CLI commands: init, index, search, stats, rebuild
- ✅ Recursive file scanning with .gitignore support
- ✅ Search by name or content
- ✅ Statistics and analytics

**Roadmap (Phase 2+):**
- 🔜 JavaScript/TypeScript parser
- 🔜 Python parser
- 🔜 Java parser
- 🔜 RAG/semantic search integration
- 🔜 Annotations system

See [TODO](./TODO) for implementation progress and [docs/projektplan.md](./docs/projektplan.md) for detailed project plan.

## Architecture

```
code-bridge/
├── cmd/
│   └── code-bridge/  # CLI application
├── pkg/
│   ├── scanner/      # File scanning
│   ├── parser/       # Code parsing (AST)
│   ├── indexer/      # JSONL indexing
│   └── search/       # Search functionality
├── internal/
│   └── config/       # Configuration
└── .code-bridge/     # Data storage
    ├── codebase.jsonl
    └── annotations.jsonl
```

## License

MIT
