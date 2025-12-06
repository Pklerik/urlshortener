// Package osexitmainchecker contains check for usage of exit function in main function
package osexitmainchecker

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var ErrOsExitMainCheckAnalyzer = &analysis.Analyzer{
	Name: "osexitmainchecker",
	Doc:  "check call os.Exit in func main() of package main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if pass.Pkg.Name() != "main" {
			continue
		}
		// tests are generating build cache that has main package, ignoring such files
		if fullPath := pass.Fset.Position(file.Pos()).String(); strings.Contains(fullPath, "go-build") {
			continue
		}

		// Find func main declarations and inspect their bodies
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name == nil || fn.Name.Name != "main" || fn.Body == nil {
				continue
			}

			ast.Inspect(fn.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				ident, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}
				if ident.Name == "os" && sel.Sel.Name == "Exit" {
					pass.Reportf(sel.Sel.Pos(), "direct call of os.Exit is not allowed for package main's main() function")
				}
				return true
			})
		}
	}

	return nil, nil
}
