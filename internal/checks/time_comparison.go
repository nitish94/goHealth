package checks

import (
	"go/ast"
	"go/token"
	"strings"

	"goHealth/internal/doctor"

	"golang.org/x/tools/go/packages"
)

type TimeComparison struct{}

func (c *TimeComparison) Name() string {
	return "TimeComparison"
}

func (c *TimeComparison) Run(pkg *packages.Package) []doctor.Diagnosis {
	var diagnoses []doctor.Diagnosis

	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			bin, ok := n.(*ast.BinaryExpr)
			if !ok || bin.Op != token.EQL { // == operator
				return true
			}

			// Check if both sides are time.Time (heuristic: selector ending with time)
			if isTimeType(bin.X) && isTimeType(bin.Y) {
				pos := pkg.Fset.Position(bin.Pos())
				diagnoses = append(diagnoses, doctor.Diagnosis{
					Severity:     doctor.SeverityCritical,
					Message:      "Time is complex. == will fail even if times imply the same instant.",
					WhyItMatters: "Go's time.Time includes monotonic clock readings that can differ even when wall clock times are identical. Direct equality checks may unexpectedly return false for logically equal times.",
					Suggestion:   "Use t1.Equal(t2) for comparing times, or compare specific fields like t1.Unix() == t2.Unix() if you only care about wall clock time.",
					File:         pos.Filename,
					Line:         pos.Line,
					CodeSnippet: "time1 == time2 // Unreliable!",
				})
			}

			return true
		})
	}

	return diagnoses
}

func isTimeType(expr ast.Expr) bool {
	// Check if it's a selector like time.Time
	sel, ok := expr.(*ast.SelectorExpr)
	if ok {
		if ident, ok := sel.X.(*ast.Ident); ok && strings.HasSuffix(ident.Name, "time") {
			return sel.Sel.Name == "Time"
		}
	}

	// Heuristic: if it's an ident with name containing "time"
	if ident, ok := expr.(*ast.Ident); ok {
		return strings.Contains(strings.ToLower(ident.Name), "time")
	}

	return false
}