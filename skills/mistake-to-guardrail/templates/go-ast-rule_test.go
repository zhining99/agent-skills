// Template: a stdlib-only AST rule for clients/mygo-native/guardrail_test.go.
// Copy the walker; keep the allowlist keyed by function name, never by line.
package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExampleRule(t *testing.T) {
	fset := token.NewFileSet()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(".", name), nil, 0)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			decl, ok := node.(*ast.FuncDecl)
			if !ok || decl.Body == nil {
				return true
			}
			found := false
			ast.Inspect(decl.Body, func(inner ast.Node) bool {
				call, ok := inner.(*ast.CallExpr)
				if !ok {
					return true
				}
				if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := selector.X.(*ast.Ident); ok && ident.Name == "exec" && selector.Sel.Name == "Command" {
						found = true
					}
				}
				return true
			})
			if found {
				pos := fset.Position(decl.Pos())
				t.Errorf("%s:%d %s spawns a process; call hideConsole(cmd) or add it to the allowlist with a reason",
					name, pos.Line, decl.Name.Name)
			}
			return true
		})
	}
}

// Allowlist shape:
//
//	var allowed = map[string]string{
//		"openURLInBrowser": "why: xdg-open has no console; review: 2027-05",
//	}
//
// A test asserts every entry has "why:" and "review:".
