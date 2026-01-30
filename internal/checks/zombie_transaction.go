package checks

import (
	"go/ast"
	"go/token"
	"strings"

	"goHealth/internal/doctor"

	"golang.org/x/tools/go/packages"
)

type ZombieTransaction struct{}

func (c *ZombieTransaction) Name() string {
	return "ZombieTransaction"
}

func (c *ZombieTransaction) Run(pkg *packages.Package) []doctor.Diagnosis {
	var diagnoses []doctor.Diagnosis

	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			// Look for blocks
			block, ok := n.(*ast.BlockStmt)
			if !ok {
				return true
			}

			for i, stmt := range block.List {
				// Look for assignment: tx, err := db.BeginTx(...)
				assign, ok := stmt.(*ast.AssignStmt)
				if !ok || assign.Tok != token.DEFINE {
					continue
				}

				// Check LHS for a variable named like "tx" (heuristic)
				txVarName := ""
				for _, expr := range assign.Lhs {
					if id, ok := expr.(*ast.Ident); ok {
						nameLower := strings.ToLower(id.Name)
						if strings.Contains(nameLower, "tx") || strings.Contains(nameLower, "transaction") {
							txVarName = id.Name
							break
						}
					}
				}

				if txVarName == "" {
					continue
				}

				// Check RHS is a call to Begin or BeginTx
				if len(assign.Rhs) == 0 {
					continue
				}
				call, ok := assign.Rhs[0].(*ast.CallExpr)
				if !ok || !isBeginCall(call) {
					continue
				}

				// Scan rest of block for defer tx.Rollback()
				rollbackDeferred := false

				for j := i + 1; j < len(block.List); j++ {
					nextStmt := block.List[j]

					if def, ok := nextStmt.(*ast.DeferStmt); ok {
						if isRollbackCall(def.Call, txVarName) {
							rollbackDeferred = true
							break
						}
					}
				}

				if !rollbackDeferred {
					diagnoses = append(diagnoses, doctor.Diagnosis{
						Severity:     doctor.SeverityCritical,
						Message:      "This transaction won't auto-rollback on error. You risk locking your database tables.",
						WhyItMatters: "Database transactions hold locks and connections until committed or rolled back. If your code panics or returns early without deferring Rollback(), the transaction remains open, blocking other operations and potentially locking tables.",
						Suggestion:   "Always defer tx.Rollback() immediately after starting a transaction, even if you plan to commit later.",
						File:         pkg.Fset.Position(assign.Pos()).Filename,
						Line:         pkg.Fset.Position(assign.Pos()).Line,
						CodeSnippet: txVarName + ", err := db.BeginTx(...) // Missing defer " + txVarName + ".Rollback()",
					})
				}
			}
			return true
		})
	}

	return diagnoses
}

func isBeginCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	return sel.Sel.Name == "Begin" || sel.Sel.Name == "BeginTx"
}

func isRollbackCall(call *ast.CallExpr, varName string) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if sel.Sel.Name != "Rollback" {
		return false
	}

	// Check X: should be the tx variable
	id, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	return id.Name == varName
}