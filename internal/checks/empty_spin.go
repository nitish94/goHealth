package checks

import (
	"go/ast"

	"goHealth/internal/doctor"

	"golang.org/x/tools/go/packages"
)

type EmptySpin struct{}

func (c *EmptySpin) Name() string {
	return "EmptySpin"
}

func (c *EmptySpin) Run(pkg *packages.Package) []doctor.Diagnosis {
	var diagnoses []doctor.Diagnosis

	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			forStmt, ok := n.(*ast.ForStmt)
			if !ok || forStmt.Cond != nil || forStmt.Init != nil || forStmt.Post != nil {
				// Not a for {} loop
				return true
			}

			// Check if body contains only select with default
			if isEmptySpin(forStmt.Body) {
				pos := pkg.Fset.Position(forStmt.Pos())
				diagnoses = append(diagnoses, doctor.Diagnosis{
					Severity:     doctor.SeverityCritical,
					Message:      "This loop spins 100% CPU waiting for nothing.",
					WhyItMatters: "A for loop with only a non-blocking select (default case) creates a busy-wait that consumes 100% CPU while waiting for channels. This wastes resources and can starve other goroutines.",
					Suggestion:   "Add a blocking case to the select (e.g., a channel receive or time.After timeout), or use time.Sleep for polling, or consider using sync.Cond for more efficient waiting.",
					File:         pos.Filename,
					Line:         pos.Line,
					CodeSnippet: "for { select { default: } } // CPU burner!",
				})
			}

			return true
		})
	}

	return diagnoses
}

func isEmptySpin(block *ast.BlockStmt) bool {
	if len(block.List) != 1 {
		return false
	}

	selectStmt, ok := block.List[0].(*ast.SelectStmt)
	if !ok {
		return false
	}

	// Check if only one case: default
	if len(selectStmt.Body.List) != 1 {
		return false
	}

	commClause, ok := selectStmt.Body.List[0].(*ast.CommClause)
	if !ok || commClause.Comm != nil { // Not default case
		return false
	}

	// Default case with no body
	return len(commClause.Body) == 0
}