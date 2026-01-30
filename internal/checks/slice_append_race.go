package checks

import (
	"go/ast"

	"goHealth/internal/doctor"

	"golang.org/x/tools/go/packages"
)

type SliceAppendRace struct{}

func (c *SliceAppendRace) Name() string {
	return "SliceAppendRace"
}

func (c *SliceAppendRace) Run(pkg *packages.Package) []doctor.Diagnosis {
	var diagnoses []doctor.Diagnosis

	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			goStmt, ok := n.(*ast.GoStmt)
			if !ok {
				return true
			}

			// Check if the go func has append calls
			if hasAppendWithoutMutex(goStmt.Call) {
				pos := pkg.Fset.Position(goStmt.Pos())
				diagnoses = append(diagnoses, doctor.Diagnosis{
					Severity:    doctor.SeverityCritical,
					Message:     "You're appending to a slice in a goroutine without synchronization. This causes data races and random panics.",
					WhyItMatters: "Slices are not thread-safe. Concurrent appends from multiple goroutines can corrupt the slice's internal state, leading to random panics, data corruption, or crashes that are extremely hard to debug. Use a mutex to protect shared slices or consider thread-safe alternatives like channels.",
					File:         pos.Filename,
					Line:         pos.Line,
					CodeSnippet: "go func() { slice = append(slice, item) } // Race condition!",
				})
			}

			return true
		})
	}

	return diagnoses
}

func hasAppendWithoutMutex(call *ast.CallExpr) bool {
	fn, ok := call.Fun.(*ast.FuncLit)
	if !ok {
		return false
	}

	hasAppend := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if callExpr, ok := n.(*ast.CallExpr); ok {
			if ident, ok := callExpr.Fun.(*ast.Ident); ok && ident.Name == "append" {
				hasAppend = true
				// For simplicity, assume no mutex check; always flag if append present
				// In a real implementation, you'd check for mutex locks
			}
		}
		return true
	})

	return hasAppend
}