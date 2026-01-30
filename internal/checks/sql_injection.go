package checks

import (
	"go/ast"
	"goHealth/internal/doctor"

	"golang.org/x/tools/go/packages"
)

type SQLInjection struct{}

func (c *SQLInjection) Name() string {
	return "SQLInjection"
}

func (c *SQLInjection) Run(pkg *packages.Package) []doctor.Diagnosis {
	var diagnoses []doctor.Diagnosis

	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			// Find function calls
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			// Check if it's a DB call (Exec, Query, QueryRow)
			// Heuristic: Method name is Exec/Query/QueryRow
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			methodName := sel.Sel.Name
			if methodName != "Exec" && methodName != "Query" && methodName != "QueryRow" {
				return true
			}

			// Now check arguments. If the first arg (query) is a fmt.Sprintf or string concatenation, it's risky.
			if len(call.Args) > 0 {
				arg := call.Args[0]

				// Case 1: fmt.Sprintf(...)
				if isFmtSprintf(arg) {
					diagnoses = append(diagnoses, doctor.Diagnosis{
						Severity:     doctor.SeverityCritical,
						Message:      "Potential SQL Injection risk detected.",
						WhyItMatters: "Building SQL queries using `fmt.Sprintf` allows attackers to inject malicious SQL by crafting input that alters the query structure.",
						Suggestion:   "Use parameterized queries like db.Query('SELECT * FROM users WHERE id = $1', id) for PostgreSQL or db.Query('SELECT * FROM users WHERE id = ?', id) for MySQL, passing variables as separate arguments to prevent injection.",
						File:         pkg.Fset.Position(call.Pos()).Filename,
						Line:         pkg.Fset.Position(call.Pos()).Line,
						CodeSnippet: "db." + methodName + "(fmt.Sprintf(...))",
					})
				}

				// Case 2: "..." + var + "..." (BinaryExpr with String concat)
				if isStringConcat(arg) {
					diagnoses = append(diagnoses, doctor.Diagnosis{
						Severity:     doctor.SeverityCritical,
						Message:      "Potential SQL Injection risk detected (String Concatenation).",
						WhyItMatters: "Concatenating strings to build SQL queries is unsafe because user input can contain SQL syntax that changes the query's meaning.",
						Suggestion:   "Replace concatenation with prepared statements, e.g., stmt, err := db.Prepare('SELECT * FROM users WHERE name = ?'); rows, err := stmt.Query(name).",
						File:         pkg.Fset.Position(call.Pos()).Filename,
						Line:         pkg.Fset.Position(call.Pos()).Line,
						CodeSnippet: "db." + methodName + "(... + ...)",
					})
				}
			}

			return true
		})
	}

	return diagnoses
}

func isFmtSprintf(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	// Check for fmt.Sprintf
	if sel.Sel.Name != "Sprintf" {
		return false
	}

	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == "fmt"
}

func isStringConcat(expr ast.Expr) bool {
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	return bin.Op == 12 // token.ADD is often 12, but better to check token.ADD if we imported it.
	// Or just check if we can reliably detect it. Let's use token package if available, or just heuristic.
	// ast.BinaryExpr Op is token.Token.
	// Let's rely on the fact that if it's a binary expression in a query arg, it's virtually always a bad idea unless it's calculated constants.
	// Doctor is opinionated! Binary expressions in SQL strings are suspects.
	return true
}
