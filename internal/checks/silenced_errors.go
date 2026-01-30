package checks

import (
	"go/ast"
	"go/token"

	"goHealth/internal/doctor"

	"golang.org/x/tools/go/packages"
)

type SilencedErrors struct{}

func (c *SilencedErrors) Name() string {
	return "SilencedErrors"
}

func (c *SilencedErrors) Run(pkg *packages.Package) []doctor.Diagnosis {
	var diagnoses []doctor.Diagnosis

	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			assign, ok := n.(*ast.AssignStmt)
			if !ok || assign.Tok != token.ASSIGN {
				return true
			}

			// Look for _, err := ...
			if len(assign.Lhs) >= 2 {
				if ident, ok := assign.Lhs[1].(*ast.Ident); ok && ident.Name == "_" {
					// Check if the RHS is a call to a critical function
					if len(assign.Rhs) == 1 {
						if call, ok := assign.Rhs[0].(*ast.CallExpr); ok && isCriticalFunction(call) {
							pos := pkg.Fset.Position(assign.Pos())
							diagnoses = append(diagnoses, doctor.Diagnosis{
								Severity:    doctor.SeverityCritical,
								Message:     "You silenced a critical error. Your user thinks it worked, but your database failed.",
								WhyItMatters: "Ignoring errors from critical operations like JSON marshaling, database writes, or HTTP responses can lead to silent failures. Users may believe operations succeeded when they actually failed (e.g., corrupt data saved to DB, empty responses sent). Always check and handle errors appropriately.",
								File:         pos.Filename,
								Line:         pos.Line,
								CodeSnippet: "_, err := json.Marshal(...) // Error ignored!",
							})
						}
					}
				}
			}

			return true
		})
	}

	return diagnoses
}

func isCriticalFunction(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	criticalFuncs := []string{"Marshal", "Unmarshal", "Exec", "Query", "Write", "WriteHeader"}
	for _, fn := range criticalFuncs {
		if sel.Sel.Name == fn {
			return true
		}
	}

	return false
}