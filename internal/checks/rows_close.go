package checks

import (
	"go/ast"
	"go/token"
	"strings"

	"goHealth/internal/doctor"

	"golang.org/x/tools/go/packages"
)

type RowsClose struct{}

func (c *RowsClose) Name() string {
	return "RowsClose"
}

func (c *RowsClose) Run(pkg *packages.Package) []doctor.Diagnosis {
	var diagnoses []doctor.Diagnosis

	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			// Look for blocks
			block, ok := n.(*ast.BlockStmt)
			if !ok {
				return true
			}

			for i, stmt := range block.List {
				// Look for assignment: rows, err := db.Query(...)
				assign, ok := stmt.(*ast.AssignStmt)
				if !ok || assign.Tok != token.DEFINE {
					continue
				}

				// Check LHS for a variable named "rows" (heuristic)
				rowsVarName := ""
				for _, expr := range assign.Lhs {
					if id, ok := expr.(*ast.Ident); ok {
						if strings.Contains(strings.ToLower(id.Name), "rows") {
							rowsVarName = id.Name
							break
						}
					}
				}

				if rowsVarName == "" {
					continue
				}

				// Check RHS is a call (heuristic)
				// Ideally check valid DB call, but let's assume if var is named `rows` and rhs is a call, it's a DB rows.
				if len(assign.Rhs) == 0 {
					continue
				}
				_, isCall := assign.Rhs[0].(*ast.CallExpr)
				if !isCall {
					continue
				}

				// Scan rest of block for defer rows.Close()
				closed := false

				for j := i + 1; j < len(block.List); j++ {
					nextStmt := block.List[j]

					if def, ok := nextStmt.(*ast.DeferStmt); ok {
						if isRowsClose(def.Call, rowsVarName) {
							closed = true
							break
						}
					}
					if exprStmt, ok := nextStmt.(*ast.ExprStmt); ok {
						if call, ok := exprStmt.X.(*ast.CallExpr); ok {
							if isRowsClose(call, rowsVarName) {
								closed = true
								break
							}
						}
					}
				}

				if !closed {
					diagnoses = append(diagnoses, doctor.Diagnosis{
						Severity: doctor.SeverityCritical,
						Message:  "Possible unclosed database rows detected.",
						WhyItMatters: "`sql.Rows` holds a database connection until you call `Close()`. " +
							"Failing to close it will quickly exhaust your connection pool. " +
							"Use `defer " + rowsVarName + ".Close()` immediately after checking for errors.",
						File:        pkg.Fset.Position(assign.Pos()).Filename,
						Line:        pkg.Fset.Position(assign.Pos()).Line,
						CodeSnippet: rowsVarName + ", err := db.Query(...) // Missing defer Close()",
					})
				}
			}
			return true
		})
	}
	return diagnoses
}

func isRowsClose(call *ast.CallExpr, varName string) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if sel.Sel.Name != "Close" {
		return false
	}

	// Check X: should be variable name
	id, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	return id.Name == varName
}
