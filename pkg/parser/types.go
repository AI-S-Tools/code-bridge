package parser

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// ElementType represents the type of code element
type ElementType string

const (
	TypeFunction  ElementType = "function"
	TypeClass     ElementType = "class"
	TypeInterface ElementType = "interface"
	TypeType      ElementType = "type"
	TypeStruct    ElementType = "struct"
	TypeVariable  ElementType = "variable"
)

// CodeElement represents a parsed code element
type CodeElement struct {
	Type       ElementType `json:"type"`
	Name       string      `json:"name"`
	File       string      `json:"file"`
	Line       int         `json:"line"`
	EndLine    int         `json:"endLine"`
	Hash       string      `json:"hash"`

	// Function/Method specific
	Params    []Parameter `json:"params,omitempty"`
	Returns   string      `json:"returns,omitempty"`
	Async     bool        `json:"async,omitempty"`
	Generator bool        `json:"generator,omitempty"`

	// Class/Struct specific
	Methods    []string `json:"methods,omitempty"`
	Extends    string   `json:"extends,omitempty"`
	Implements []string `json:"implements,omitempty"`
	Fields     []string `json:"fields,omitempty"`

	// Common
	Body       string   `json:"body"`
	Docstring  string   `json:"docstring,omitempty"`
	Imports    []string `json:"imports,omitempty"`
	Exports    bool     `json:"exports,omitempty"`

	// Metadata
	Language  string    `json:"language"`
	IndexedAt time.Time `json:"indexedAt"`
}

// Parameter represents a function/method parameter
type Parameter struct {
	Name     string `json:"name"`
	Type     string `json:"type,omitempty"`
	Default  string `json:"default,omitempty"`
	Optional bool   `json:"optional,omitempty"`
}

// ParseResult contains parsing results and errors
type ParseResult struct {
	Elements []CodeElement
	Errors   []ParseError
}

// ParseError represents a parsing error
type ParseError struct {
	Message string
	Line    int
	Column  int
}

// Parser interface for language-specific parsers
type Parser interface {
	Parse(filePath string, content []byte) (*ParseResult, error)
	SupportsFile(filePath string) bool
}

// HashCode generates a hash from code body
func HashCode(body string) string {
	hash := sha256.Sum256([]byte(body))
	return hex.EncodeToString(hash[:])[:16]
}
