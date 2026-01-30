package checks

import (
	"go/ast"
	"strings"

	"goHealth/internal/doctor"

	"golang.org/x/tools/go/packages"
)

type ExitInLibrary struct{}

func (c *ExitInLibrary) Name() string {
	return "ExitInLibrary"
}

func (c *ExitInLibrary) Run(pkg *packages.Package) []doctor.Diagnosis {
	var diagnoses []doctor.Diagnosis

	// Only check non-main packages
	if pkg.Name == "main" {
		return diagnoses
	}

	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			if isExitCall(call) {
				pos := pkg.Fset.Position(call.Pos())
				diagnoses = append(diagnoses, doctor.Diagnosis{
					Severity:    doctor.SeverityCritical,
					Message:     "This library package has a kill switch. Only main should decide when to die.",
					WhyItMatters: "Library packages should never call os.Exit, log.Fatal, or panic. These terminate the entire process and prevent the caller from handling errors gracefully. Libraries must return errors instead, allowing the main package (or caller) to decide how to respond - whether to log, retry, or exit.",
					File:         pos.Filename,
					Line:         pos.Line,
					CodeSnippet: "os.Exit(...) or log.Fatal(...) or panic(...)",
				})
			}

			return true
		})
	}

	return diagnoses
}

func isExitCall(call *ast.CallExpr) bool {
	// Check for builtin panic
	if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "panic" {
		return true
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	// Check for os.Exit
	if sel.Sel.Name == "Exit" {
		if ident, ok := sel.X.(*ast.Ident); ok && strings.HasSuffix(ident.Name, "os") {
			return true
		}
	}

	// Check for log.Fatal*
	if strings.HasPrefix(sel.Sel.Name, "Fatal") {
		if ident, ok := sel.X.(*ast.Ident); ok && strings.HasSuffix(ident.Name, "log") {
			return true
		}
	}

	return false
}