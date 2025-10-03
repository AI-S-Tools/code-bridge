package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"time"
)

// GoParser parses Go source code
type GoParser struct{}

// NewGoParser creates a new Go parser
func NewGoParser() *GoParser {
	return &GoParser{}
}

// SupportsFile checks if the parser supports this file
func (p *GoParser) SupportsFile(filePath string) bool {
	ext := filepath.Ext(filePath)
	return ext == ".go"
}

// Parse parses Go source code and extracts elements
func (p *GoParser) Parse(filePath string, content []byte) (*ParseResult, error) {
	result := &ParseResult{
		Elements: make([]CodeElement, 0),
		Errors:   make([]ParseError, 0),
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		result.Errors = append(result.Errors, ParseError{
			Message: err.Error(),
		})
		return result, nil
	}

	// Extract imports
	imports := p.extractImports(file)

	// Walk AST
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if element := p.extractFunction(node, fset, filePath, string(content), imports); element != nil {
				result.Elements = append(result.Elements, *element)
			}
		case *ast.GenDecl:
			// Handle type declarations (struct, interface, type alias)
			for _, spec := range node.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if element := p.extractType(s, node, fset, filePath, string(content)); element != nil {
						result.Elements = append(result.Elements, *element)
					}
				}
			}
		}
		return true
	})

	return result, nil
}

// extractFunction extracts function/method information
func (p *GoParser) extractFunction(node *ast.FuncDecl, fset *token.FileSet, filePath, content string, imports []string) *CodeElement {
	if node.Name == nil {
		return nil
	}

	pos := fset.Position(node.Pos())
	endPos := fset.Position(node.End())

	body := p.extractNodeBody(node.Pos(), node.End(), content)
	docstring := p.extractDocstring(node.Doc)

	params := p.extractParams(node.Type.Params)
	returns := p.extractReturns(node.Type.Results)

	// Check if this is a method (has receiver)
	isMethod := node.Recv != nil
	name := node.Name.Name
	if isMethod && node.Recv.NumFields() > 0 {
		// Include receiver type in name for methods
		recvType := p.getReceiverType(node.Recv.List[0].Type)
		name = recvType + "." + name
	}

	return &CodeElement{
		Type:      TypeFunction,
		Name:      name,
		File:      filePath,
		Line:      pos.Line,
		EndLine:   endPos.Line,
		Hash:      HashCode(body),
		Params:    params,
		Returns:   returns,
		Body:      body,
		Docstring: docstring,
		Imports:   imports,
		Exports:   ast.IsExported(node.Name.Name),
		Language:  "go",
		IndexedAt: time.Now(),
	}
}

// extractType extracts struct, interface, or type alias
func (p *GoParser) extractType(spec *ast.TypeSpec, decl *ast.GenDecl, fset *token.FileSet, filePath, content string) *CodeElement {
	pos := fset.Position(decl.Pos())
	endPos := fset.Position(decl.End())

	body := p.extractNodeBody(decl.Pos(), decl.End(), content)
	docstring := p.extractDocstring(decl.Doc)

	element := &CodeElement{
		Name:      spec.Name.Name,
		File:      filePath,
		Line:      pos.Line,
		EndLine:   endPos.Line,
		Hash:      HashCode(body),
		Body:      body,
		Docstring: docstring,
		Exports:   ast.IsExported(spec.Name.Name),
		Language:  "go",
		IndexedAt: time.Now(),
	}

	switch typeNode := spec.Type.(type) {
	case *ast.StructType:
		element.Type = TypeStruct
		element.Fields = p.extractFields(typeNode.Fields)
		element.Methods = []string{} // Will be filled separately

	case *ast.InterfaceType:
		element.Type = TypeInterface
		element.Methods = p.extractInterfaceMethods(typeNode.Methods)

	default:
		element.Type = TypeType
	}

	return element
}

// extractParams extracts function parameters
func (p *GoParser) extractParams(fields *ast.FieldList) []Parameter {
	if fields == nil {
		return []Parameter{}
	}

	params := make([]Parameter, 0)
	for _, field := range fields.List {
		typeStr := p.exprToString(field.Type)

		if len(field.Names) == 0 {
			// Unnamed parameter
			params = append(params, Parameter{
				Name: "_",
				Type: typeStr,
			})
		} else {
			for _, name := range field.Names {
				params = append(params, Parameter{
					Name: name.Name,
					Type: typeStr,
				})
			}
		}
	}
	return params
}

// extractReturns extracts return types
func (p *GoParser) extractReturns(fields *ast.FieldList) string {
	if fields == nil || len(fields.List) == 0 {
		return ""
	}

	returns := make([]string, 0)
	for _, field := range fields.List {
		typeStr := p.exprToString(field.Type)
		if len(field.Names) > 0 {
			for _, name := range field.Names {
				returns = append(returns, name.Name+" "+typeStr)
			}
		} else {
			returns = append(returns, typeStr)
		}
	}

	if len(returns) == 1 {
		return returns[0]
	}
	return "(" + strings.Join(returns, ", ") + ")"
}

// extractFields extracts struct fields
func (p *GoParser) extractFields(fields *ast.FieldList) []string {
	if fields == nil {
		return []string{}
	}

	fieldNames := make([]string, 0)
	for _, field := range fields.List {
		if len(field.Names) > 0 {
			for _, name := range field.Names {
				fieldNames = append(fieldNames, name.Name)
			}
		}
	}
	return fieldNames
}

// extractInterfaceMethods extracts interface method names
func (p *GoParser) extractInterfaceMethods(fields *ast.FieldList) []string {
	if fields == nil {
		return []string{}
	}

	methods := make([]string, 0)
	for _, field := range fields.List {
		if len(field.Names) > 0 {
			for _, name := range field.Names {
				methods = append(methods, name.Name)
			}
		}
	}
	return methods
}

// extractImports extracts import statements
func (p *GoParser) extractImports(file *ast.File) []string {
	imports := make([]string, 0)
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		imports = append(imports, path)
	}
	return imports
}

// extractDocstring extracts documentation comment
func (p *GoParser) extractDocstring(doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}
	return doc.Text()
}

// extractNodeBody extracts source code for a node
func (p *GoParser) extractNodeBody(start, end token.Pos, content string) string {
	if start == 0 || end == 0 {
		return ""
	}
	startIdx := int(start) - 1
	endIdx := int(end) - 1
	if startIdx < 0 || endIdx > len(content) || startIdx >= endIdx {
		return ""
	}
	return content[startIdx:endIdx]
}

// exprToString converts expression to string
func (p *GoParser) exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + p.exprToString(t.X)
	case *ast.ArrayType:
		return "[]" + p.exprToString(t.Elt)
	case *ast.SelectorExpr:
		return p.exprToString(t.X) + "." + t.Sel.Name
	case *ast.MapType:
		return "map[" + p.exprToString(t.Key) + "]" + p.exprToString(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.FuncType:
		return "func"
	default:
		return "unknown"
	}
}

// getReceiverType extracts receiver type name
func (p *GoParser) getReceiverType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return p.getReceiverType(t.X)
	default:
		return "unknown"
	}
}
