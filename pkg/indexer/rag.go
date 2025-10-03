package indexer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AI-S-Tools/code-bridge/pkg/parser"
)

// RAGOutput represents organized code elements for RAG consumption
type RAGOutput struct {
	Summary       string
	TotalElements int
	ByFile        map[string][]RAGElement
	ByType        map[parser.ElementType][]RAGElement
}

// RAGElement is a simplified element for RAG output
type RAGElement struct {
	Type       parser.ElementType
	Name       string
	File       string
	Line       int
	Signature  string
	Docstring  string
}

// GetRAGIndex returns organized code elements for RAG/LLM consumption
func (idx *Indexer) GetRAGIndex(groupBy string) (*RAGOutput, error) {
	elements, err := idx.ReadAll()
	if err != nil {
		return nil, err
	}

	output := &RAGOutput{
		TotalElements: len(elements),
		ByFile:        make(map[string][]RAGElement),
		ByType:        make(map[parser.ElementType][]RAGElement),
	}

	// Convert to RAG elements and organize
	for _, el := range elements {
		ragEl := RAGElement{
			Type:      el.Type,
			Name:      el.Name,
			File:      el.File,
			Line:      el.Line,
			Signature: buildSignature(el),
			Docstring: el.Docstring,
		}

		output.ByFile[el.File] = append(output.ByFile[el.File], ragEl)
		output.ByType[el.Type] = append(output.ByType[el.Type], ragEl)
	}

	// Sort elements
	for _, fileElements := range output.ByFile {
		sort.Slice(fileElements, func(i, j int) bool {
			return fileElements[i].Line < fileElements[j].Line
		})
	}

	for _, typeElements := range output.ByType {
		sort.Slice(typeElements, func(i, j int) bool {
			return typeElements[i].Name < typeElements[j].Name
		})
	}

	output.Summary = generateSummary(output)

	return output, nil
}

// buildSignature creates a readable signature for an element
func buildSignature(el parser.CodeElement) string {
	switch el.Type {
	case parser.TypeFunction:
		params := make([]string, len(el.Params))
		for i, p := range el.Params {
			if p.Type != "" {
				params[i] = fmt.Sprintf("%s %s", p.Name, p.Type)
			} else {
				params[i] = p.Name
			}
		}
		sig := fmt.Sprintf("%s(%s)", el.Name, strings.Join(params, ", "))
		if el.Returns != "" {
			sig += " " + el.Returns
		}
		return sig

	case parser.TypeStruct:
		if len(el.Fields) > 0 {
			return fmt.Sprintf("%s {%d fields}", el.Name, len(el.Fields))
		}
		return el.Name

	case parser.TypeInterface:
		if len(el.Methods) > 0 {
			return fmt.Sprintf("%s {%d methods}", el.Name, len(el.Methods))
		}
		return el.Name

	default:
		return el.Name
	}
}

// generateSummary creates a text summary for RAG
func generateSummary(output *RAGOutput) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Codebase Index - %d elements\n\n", output.TotalElements))

	// Summary by type
	sb.WriteString("## By Type\n")
	types := make([]parser.ElementType, 0, len(output.ByType))
	for t := range output.ByType {
		types = append(types, t)
	}
	sort.Slice(types, func(i, j int) bool {
		return string(types[i]) < string(types[j])
	})

	for _, t := range types {
		sb.WriteString(fmt.Sprintf("- %s: %d\n", t, len(output.ByType[t])))
	}
	sb.WriteString("\n")

	// Summary by file
	sb.WriteString("## By File\n")
	files := make([]string, 0, len(output.ByFile))
	for f := range output.ByFile {
		files = append(files, f)
	}
	sort.Strings(files)

	for _, f := range files {
		sb.WriteString(fmt.Sprintf("- %s: %d elements\n", f, len(output.ByFile[f])))
	}

	return sb.String()
}

// FormatRAGByFile formats RAG output grouped by file
func FormatRAGByFile(output *RAGOutput) string {
	var sb strings.Builder

	sb.WriteString(output.Summary)
	sb.WriteString("\n---\n\n")

	// Sort files
	files := make([]string, 0, len(output.ByFile))
	for f := range output.ByFile {
		files = append(files, f)
	}
	sort.Strings(files)

	for _, file := range files {
		elements := output.ByFile[file]
		sb.WriteString(fmt.Sprintf("## File: %s (%d elements)\n\n", file, len(elements)))

		for _, el := range elements {
			sb.WriteString(fmt.Sprintf("### %s %s\n", el.Type, el.Name))
			sb.WriteString(fmt.Sprintf("**Location:** %s:%d\n", el.File, el.Line))
			sb.WriteString(fmt.Sprintf("**Signature:** `%s`\n", el.Signature))
			if el.Docstring != "" {
				sb.WriteString(fmt.Sprintf("**Doc:** %s\n", strings.TrimSpace(el.Docstring)))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// FormatRAGByType formats RAG output grouped by type
func FormatRAGByType(output *RAGOutput) string {
	var sb strings.Builder

	sb.WriteString(output.Summary)
	sb.WriteString("\n---\n\n")

	// Sort types
	types := make([]parser.ElementType, 0, len(output.ByType))
	for t := range output.ByType {
		types = append(types, t)
	}
	sort.Slice(types, func(i, j int) bool {
		return string(types[i]) < string(types[j])
	})

	for _, elemType := range types {
		elements := output.ByType[elemType]
		sb.WriteString(fmt.Sprintf("## %s (%d)\n\n", elemType, len(elements)))

		for _, el := range elements {
			sb.WriteString(fmt.Sprintf("### %s\n", el.Name))
			sb.WriteString(fmt.Sprintf("**Location:** %s:%d\n", el.File, el.Line))
			sb.WriteString(fmt.Sprintf("**Signature:** `%s`\n", el.Signature))
			if el.Docstring != "" {
				sb.WriteString(fmt.Sprintf("**Doc:** %s\n", strings.TrimSpace(el.Docstring)))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// FormatRAGCompact formats RAG output in compact list format
func FormatRAGCompact(output *RAGOutput) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Available Code Elements (%d total)\n\n", output.TotalElements))

	// Sort types
	types := make([]parser.ElementType, 0, len(output.ByType))
	for t := range output.ByType {
		types = append(types, t)
	}
	sort.Slice(types, func(i, j int) bool {
		return string(types[i]) < string(types[j])
	})

	for _, elemType := range types {
		elements := output.ByType[elemType]
		sb.WriteString(fmt.Sprintf("\n## %s\n\n", elemType))

		for _, el := range elements {
			sb.WriteString(fmt.Sprintf("- `%s` - %s:%d\n", el.Signature, el.File, el.Line))
		}
	}

	return sb.String()
}
