# Code-Bridge

RAG-enabled code indexing and search tool using JSONL format.

## Overview

Code-Bridge scans your codebase, extracts functions and classes, and stores them in a searchable JSONL index. It provides semantic search capabilities using RAG (Retrieval-Augmented Generation) and allows annotation of code elements without modifying source files.

## Features

- 🔍 **Code Indexing**: Recursively scan and index entire codebases
- 🎯 **Precise References**: File and line number references for every code element
- 🤖 **RAG Search**: Semantic search using natural language queries
- 🏷️ **Annotations**: Add metadata to code without modifying files
- 📝 **JSONL Format**: Fast, streamable, line-oriented format
- 🔧 **Multi-language**: Support for JavaScript, TypeScript, Python, Go, Java, and more

## Installation

```bash
go install github.com/AI-S-Tools/code-bridge/cmd/code-bridge@latest
```

Or build from source:

```bash
git clone https://github.com/AI-S-Tools/code-bridge.git
cd code-bridge
go build -o code-bridge ./cmd/code-bridge
```

## Quick Start

```bash
# Initialize in your project
code-bridge init

# Index your codebase
code-bridge index

# Search for code
code-bridge search "authentication function"

# RAG query
code-bridge rag "find functions that validate user input"

# Add annotation
code-bridge annotate add --target myFunction --tags "reviewed,critical"
```

## Project Status

✅ **Phase 1 Complete** - Core Infrastructure implemented in Golang

**Currently Supported:**
- Go language parsing (functions, methods, structs, interfaces, types)
- JSONL indexing with deduplication
- CLI commands: init, index, search, stats, rebuild
- File scanning with .gitignore support

**Coming Soon (Phase 2):**
- JavaScript/TypeScript parser
- Python parser
- RAG integration
- Advanced search features

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
