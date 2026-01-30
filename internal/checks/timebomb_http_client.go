package checks

import (
	"go/ast"

	"goHealth/internal/doctor"

	"golang.org/x/tools/go/packages"
)

type TimebombHttpClient struct{}

func (c *TimebombHttpClient) Name() string {
	return "TimebombHttpClient"
}

func (c *TimebombHttpClient) Run(pkg *packages.Package) []doctor.Diagnosis {
	var diagnoses []doctor.Diagnosis

	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.CallExpr:
				if isHttpGetOrPost(node) {
					pos := pkg.Fset.Position(node.Pos())
					diagnoses = append(diagnoses, doctor.Diagnosis{
						Severity:    doctor.SeverityCritical,
						Message:     "You are using a client that waits forever. One slow API call will crash your platform.",
						WhyItMatters: "The default Go HTTP client has no timeout. If a third-party service hangs, your application will hang, exhaust file descriptors, and crash. Always set a timeout on HTTP clients.",
						File:         pos.Filename,
						Line:         pos.Line,
						CodeSnippet: "http.Get(...) or http.Post(...)",
					})
				}
			case *ast.CompositeLit:
				if isHttpClientWithoutTimeout(node) {
					pos := pkg.Fset.Position(node.Pos())
					diagnoses = append(diagnoses, doctor.Diagnosis{
						Severity:    doctor.SeverityCritical,
						Message:     "You are using a client that waits forever. One slow API call will crash your platform.",
						WhyItMatters: "HTTP clients without a timeout can cause your application to hang indefinitely on slow or unresponsive servers, leading to resource exhaustion and crashes. Set a reasonable timeout.",
						File:         pos.Filename,
						Line:         pos.Line,
						CodeSnippet: "&http.Client{...} without Timeout",
					})
				}
			}
			return true
		})
	}

	return diagnoses
}

func isHttpGetOrPost(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if sel.Sel.Name != "Get" && sel.Sel.Name != "Post" {
		return false
	}
	x, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return x.Name == "http"
}

func isHttpClientWithoutTimeout(lit *ast.CompositeLit) bool {
	// Check if it's http.Client
	typ := lit.Type
	switch t := typ.(type) {
	case *ast.SelectorExpr:
		if t.Sel.Name == "Client" {
			if x, ok := t.X.(*ast.Ident); ok && x.Name == "http" {
				// It's http.Client, check if Timeout is set
				return !hasTimeoutField(lit.Elts)
			}
		}
	case *ast.StarExpr: // &http.Client
		if sel, ok := t.X.(*ast.SelectorExpr); ok {
			if sel.Sel.Name == "Client" {
				if x, ok := sel.X.(*ast.Ident); ok && x.Name == "http" {
					return !hasTimeoutField(lit.Elts)
				}
			}
		}
	}
	return false
}

func hasTimeoutField(elts []ast.Expr) bool {
	for _, elt := range elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			if ident, ok := kv.Key.(*ast.Ident); ok && ident.Name == "Timeout" {
				return true
			}
		}
	}
	return false
}