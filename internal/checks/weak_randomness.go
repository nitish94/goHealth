package checks

import (
	"go/ast"
	"strings"

	"goHealth/internal/doctor"

	"golang.org/x/tools/go/packages"
)

type WeakRandomness struct{}

func (c *WeakRandomness) Name() string {
	return "WeakRandomness"
}

func (c *WeakRandomness) Run(pkg *packages.Package) []doctor.Diagnosis {
	var diagnoses []doctor.Diagnosis

	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			if isWeakRandCall(call) && looksLikeSecurityContext(call, file) {
				pos := pkg.Fset.Position(call.Pos())
				diagnoses = append(diagnoses, doctor.Diagnosis{
					Severity:    doctor.SeverityCritical,
					Message:     "You are generating security tokens with a predictable RNG. Attackers can guess your session tokens.",
					WhyItMatters: "math/rand uses a deterministic algorithm seeded typically with time.Now().UnixNano(), making it predictable for attackers who can guess the seed range. This compromises security tokens, passwords, and session IDs. Suggestion: Use crypto/rand for secure random bytes (e.g., token := make([]byte, 32); crypto/rand.Read(token)), or higher-level packages like github.com/google/uuid for generating unique identifiers.",
					File:         pos.Filename,
					Line:         pos.Line,
					CodeSnippet: "math/rand.Int() // Predictable!",
				})
			}

			return true
		})
	}

	return diagnoses
}

func isWeakRandCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if ident, ok := sel.X.(*ast.Ident); !ok || !strings.HasSuffix(ident.Name, "rand") {
		return false
	}

	// Common math/rand functions
	randFuncs := []string{"Int", "Intn", "Int31", "Int31n", "Int63", "Int63n", "Float32", "Float64", "Perm", "Read"}
	for _, fn := range randFuncs {
		if sel.Sel.Name == fn {
			return true
		}
	}

	return false
}

func looksLikeSecurityContext(call *ast.CallExpr, file *ast.File) bool {
	// Heuristic: check if nearby comments or variable names suggest security use
	// For simplicity, check if the file contains words like "token", "password", "secret", "session"
	// Or if the function name contains them

	// Get the function containing this call
	var funcName string
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			ast.Inspect(fn, func(inner ast.Node) bool {
				if inner == call {
					funcName = fn.Name.Name
					return false
				}
				return true
			})
		}
		return true
	})

	securityKeywords := []string{"token", "password", "secret", "session", "auth", "key", "nonce"}
	nameLower := strings.ToLower(funcName)
	for _, kw := range securityKeywords {
		if strings.Contains(nameLower, kw) {
			return true
		}
	}

	// Also check comments in the file
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			commentLower := strings.ToLower(c.Text)
			for _, kw := range securityKeywords {
				if strings.Contains(commentLower, kw) {
					return true
				}
			}
		}
	}

	return false // Conservative: don't flag unless context suggests security
}