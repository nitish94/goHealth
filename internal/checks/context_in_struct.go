package checks

import (
	"go/ast"
	"goHealth/internal/doctor"

	"golang.org/x/tools/go/packages"
)

type ContextInStruct struct{}

func (c *ContextInStruct) Name() string {
	return "ContextInStruct"
}

func (c *ContextInStruct) Run(pkg *packages.Package) []doctor.Diagnosis {
	var diagnoses []doctor.Diagnosis

	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			// Find Type Specs (struct definitions)
			typeSpec, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				return true // It's an interface or alias
			}

			for _, field := range structType.Fields.List {
				// Check field type
				// Identifying context.Context
				if isContextType(field.Type) {
					diagnoses = append(diagnoses, doctor.Diagnosis{
						Severity: doctor.SeverityWarning,
						Message:  "Do not store `context.Context` in a struct type.",
						WhyItMatters: "`context.Context` is request-scoped and should be passed as the first argument to functions, not stored in a struct whose lifecycle might exceed the request. " +
							"Storing it can lead to memory leaks or retaining a cancelled context.",
						File:        pkg.Fset.Position(field.Pos()).Filename,
						Line:        pkg.Fset.Position(field.Pos()).Line,
						CodeSnippet: "type " + typeSpec.Name.Name + " struct { ... " + getFieldName(field) + " context.Context ... }",
					})
				}
			}

			return true
		})
	}

	return diagnoses
}

func isContextType(expr ast.Expr) bool {
	// Check for "context.Context" selector
	sel, ok := expr.(*ast.SelectorExpr)
	if ok {
		if sel.Sel.Name == "Context" {
			if id, ok := sel.X.(*ast.Ident); ok {
				return id.Name == "context"
			}
		}
	}
	return false
}

func getFieldName(field *ast.Field) string {
	if len(field.Names) > 0 {
		return field.Names[0].Name
	}
	return "" // Embedded
}
