package checks

import (
	"go/ast"
	"strings"

	"goHealth/internal/doctor"

	"golang.org/x/tools/go/packages"
)

type UnclosedBody struct{}

func (c *UnclosedBody) Name() string {
	return "UnclosedBody"
}

func (c *UnclosedBody) Run(pkg *packages.Package) []doctor.Diagnosis {
	var diagnoses []doctor.Diagnosis

	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			// Find Blocks
			block, ok := n.(*ast.BlockStmt)
			if !ok {
				return true
			}

			// Iterate over statements in the block
			for i, stmt := range block.List {
				// Look for: resp, err := http.Get(...)
				assign, ok := stmt.(*ast.AssignStmt)
				if !ok {
					continue
				}

				respVarName := ""
				for _, expr := range assign.Lhs {
					if id, ok := expr.(*ast.Ident); ok {
						// Heuristic: Variable name contains "resp"
						if strings.Contains(strings.ToLower(id.Name), "resp") {
							respVarName = id.Name
							break
						}
					}
				}

				if respVarName == "" {
					continue
				}

				// Now check if it's closed in the *remainder* of the block
				// OR if it's returned (which means responsibility passes to caller)
				// This is a naive check.

				closed := false
				returned := false

				for j := i + 1; j < len(block.List); j++ {
					nextStmt := block.List[j]

					// Check for defer resp.Body.Close()
					if def, ok := nextStmt.(*ast.DeferStmt); ok {
						if isBodyClose(def.Call, respVarName) {
							closed = true
							break
						}
					}

					// Check for resp.Body.Close()
					if exprStmt, ok := nextStmt.(*ast.ExprStmt); ok {
						if call, ok := exprStmt.X.(*ast.CallExpr); ok {
							if isBodyClose(call, respVarName) {
								closed = true
								break
							}
						}
					}

					// Check for return resp
					if ret, ok := nextStmt.(*ast.ReturnStmt); ok {
						for _, res := range ret.Results {
							if id, ok := res.(*ast.Ident); ok && id.Name == respVarName {
								returned = true
							}
						}
					}
				}

				if !closed && !returned {
					diagnoses = append(diagnoses, doctor.Diagnosis{
						Severity: doctor.SeverityCritical,
						Message:  "Possible unclosed HTTP response body detected.",
						WhyItMatters: "Response bodies must be closed to avoid leaking file descriptors. " +
							"Use `defer " + respVarName + ".Body.Close()` immediately after checking for errors.",
						File:        pkg.Fset.Position(assign.Pos()).Filename,
						Line:        pkg.Fset.Position(assign.Pos()).Line,
						CodeSnippet: "resp, err := ... // Missing defer resp.Body.Close()",
					})
				}
			}

			return true
		})
	}

	return diagnoses
}

func isBodyClose(call *ast.CallExpr, varName string) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr) // ??.Close
	if !ok {
		return false
	}
	if sel.Sel.Name != "Close" {
		return false
	}

	// Check X: must be resp.Body
	sel2, ok := sel.X.(*ast.SelectorExpr) // resp.Body
	if !ok {
		return false
	}
	if sel2.Sel.Name != "Body" {
		return false
	}

	id, ok := sel2.X.(*ast.Ident) // resp
	if !ok {
		return false
	}

	return id.Name == varName
}
