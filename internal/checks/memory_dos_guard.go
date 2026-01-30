package checks

import (
	"go/ast"
	"strings"

	"goHealth/internal/doctor"

	"golang.org/x/tools/go/packages"
)

type MemoryDOSGuard struct{}

func (c *MemoryDOSGuard) Name() string {
	return "MemoryDOSGuard"
}

func (c *MemoryDOSGuard) Run(pkg *packages.Package) []doctor.Diagnosis {
	var diagnoses []doctor.Diagnosis

	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			if isUnlimitedReadAll(call) {
				pos := pkg.Fset.Position(call.Pos())
				diagnoses = append(diagnoses, doctor.Diagnosis{
					Severity:    doctor.SeverityCritical,
					Message:     "You are reading unlimited data into memory. One huge request will OOM your server.",
					WhyItMatters: "io.ReadAll reads the entire input into memory without limits. An attacker can send a massive payload (e.g., 10GB JSON) that exhausts RAM and crashes your server. Always wrap readers in io.LimitReader to cap memory usage, or stream/process data incrementally.",
					File:         pos.Filename,
					Line:         pos.Line,
					CodeSnippet: "io.ReadAll(r.Body) // No limit!",
				})
			}

			return true
		})
	}

	return diagnoses
}

func isUnlimitedReadAll(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "ReadAll" {
		return false
	}
	if ident, ok := sel.X.(*ast.Ident); !ok || !strings.HasSuffix(ident.Name, "io") {
		return false
	}

	if len(call.Args) == 0 {
		return false
	}

	// Check if the argument is NOT io.LimitReader(...)
	if limitCall, ok := call.Args[0].(*ast.CallExpr); ok {
		if isLimitReader(limitCall) {
			return false // It's limited, so not unlimited
		}
	}

	return true // Flag as unlimited
}

func isLimitReader(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "LimitReader" {
		return false
	}
	if ident, ok := sel.X.(*ast.Ident); !ok || !strings.HasSuffix(ident.Name, "io") {
		return false
	}
	return true
}