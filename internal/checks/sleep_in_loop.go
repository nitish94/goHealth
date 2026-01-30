package checks

import (
	"go/ast"

	"goHealth/internal/doctor"

	"golang.org/x/tools/go/packages"
)

type SleepInLoop struct{}

func (c *SleepInLoop) Name() string {
	return "SleepInLoop"
}

func (c *SleepInLoop) Run(pkg *packages.Package) []doctor.Diagnosis {
	var diagnoses []doctor.Diagnosis

	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			// Find for loops
			loop, ok := n.(*ast.ForStmt)
			if ok {
				// Inspect inside the loop
				ast.Inspect(loop.Body, func(inner ast.Node) bool {
					call, ok := inner.(*ast.CallExpr)
					if !ok {
						return true
					}

					// Check if it's time.Sleep
					// This is a naive check using selector expression
					sel, ok := call.Fun.(*ast.SelectorExpr)
					if ok {
						// We need to check if X is "time" and Sel is "Sleep"
						//Ideally we check type info, but efficient string matching works for v1
						if isTimeSleep(pkg, sel) {
							diagnoses = append(diagnoses, doctor.Diagnosis{
								Severity: doctor.SeverityCritical,
								Message:  "Blocking `time.Sleep` call detected inside a loop.",
								WhyItMatters: "Sleeping inside a loop blocks the entire goroutine. " +
									"If this loop processes requests or events, it will freeze. " +
									"Use a `time.Ticker` with a `select` statement to allow for cancellation/context awareness.",
								File:        pkg.Fset.Position(call.Pos()).Filename,
								Line:        pkg.Fset.Position(call.Pos()).Line,
								CodeSnippet: "time.Sleep(...)",
							})
						}
					}

					// Also check for Range loops
					return true
				})
			}

			// Also need to handle RangeStmt
			rangeLoop, ok := n.(*ast.RangeStmt)
			if ok {
				ast.Inspect(rangeLoop.Body, func(inner ast.Node) bool {
					call, ok := inner.(*ast.CallExpr)
					if !ok {
						return true
					}
					sel, ok := call.Fun.(*ast.SelectorExpr)
					if ok {
						if isTimeSleep(pkg, sel) {
							diagnoses = append(diagnoses, doctor.Diagnosis{
								Severity: doctor.SeverityCritical,
								Message:  "Blocking `time.Sleep` call detected inside a range loop.",
								WhyItMatters: "Sleeping inside a loop blocks the entire goroutine. " +
									"Use a `time.Ticker` with a `select` statement.",
								File:        pkg.Fset.Position(call.Pos()).Filename,
								Line:        pkg.Fset.Position(call.Pos()).Line,
								CodeSnippet: "time.Sleep(...)",
							})
						}
					}
					return true
				})
			}

			return true
		})
	}

	return diagnoses
}

func isTimeSleep(pkg *packages.Package, sel *ast.SelectorExpr) bool {
	// 1. Check if the method name is Sleep
	if sel.Sel.Name != "Sleep" {
		return false
	}

	// 2. Check if the package is "time"
	// Using Uses[sel.Sel] gives us the Object (Func)
	// Then we check its Pkg()
	if pkg.TypesInfo == nil {
		// Fallback to naive check if TypesInfo is missing (shouldn't happen with correct Load mode)
		// But sel.X is an expression (identifier "time")
		if id, ok := sel.X.(*ast.Ident); ok {
			return id.Name == "time"
		}
		return false
	}

	obj := pkg.TypesInfo.Uses[sel.Sel]
	if obj != nil && obj.Pkg() != nil && obj.Pkg().Name() == "time" {
		return true
	}

	// If obj is nil (sometimes imports are weird), try string fallback
	if id, ok := sel.X.(*ast.Ident); ok {
		return id.Name == "time"
	}

	return false
}
