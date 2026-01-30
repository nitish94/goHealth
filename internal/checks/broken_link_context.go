package checks

import (
	"go/ast"
	"strings"

	"goHealth/internal/doctor"

	"golang.org/x/tools/go/packages"
)

type BrokenLinkContext struct{}

func (c *BrokenLinkContext) Name() string {
	return "BrokenLinkContext"
}

func (c *BrokenLinkContext) Run(pkg *packages.Package) []doctor.Diagnosis {
	var diagnoses []doctor.Diagnosis

	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			// Check if function has a context parameter
			ctxParam := findContextParam(fn)
			if ctxParam == "" {
				return true
			}

			// Now inspect the function body for bad context usage
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				// Check if any argument is context.TODO() or context.Background()
				for _, arg := range call.Args {
					if isBadContextCall(arg) {
						pos := pkg.Fset.Position(call.Pos())
						diagnoses = append(diagnoses, doctor.Diagnosis{
							Severity:    doctor.SeverityCritical,
							Message:     "You dropped the context chain. If the request cancels, this work will keep running, wasting resources.",
							WhyItMatters: "Using context.TODO() or context.Background() breaks request cancellation and tracing. If the parent context cancels (e.g., user disconnects), this operation ignores it, leading to zombie work that burns CPU/memory until completion. Always pass the inherited context to maintain cancellation and observability.",
							File:         pos.Filename,
							Line:         pos.Line,
							CodeSnippet: "someFunc(context.Background()) // Should be someFunc(ctx)",
						})
						break // Only flag once per call
					}
				}

				return true
			})

			return true
		})
	}

	return diagnoses
}

func findContextParam(fn *ast.FuncDecl) string {
	if fn.Type.Params == nil {
		return ""
	}
	for _, param := range fn.Type.Params.List {
		if len(param.Names) > 0 {
			paramName := param.Names[0].Name
			if isContextType(param.Type) {
				return paramName
			}
		}
	}
	return ""
}



func isBadContextCall(arg ast.Expr) bool {
	call, ok := arg.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if sel.Sel.Name != "TODO" && sel.Sel.Name != "Background" {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return strings.HasSuffix(ident.Name, "context")
}