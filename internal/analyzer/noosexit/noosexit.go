// Package noosexit custom ast analyzer
package noosexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

const doc = "noosexit checks that os.Exit is not called directly from main package"

// Analyzer reports direct os.Exit calls inside package main.
var Analyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  doc,
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// The requirement is specifically about main package.
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			if selector.Sel.Name != "Exit" {
				return true
			}

			ident, ok := selector.X.(*ast.Ident)
			if !ok {
				return true
			}

			if ident.Name != "os" {
				return true
			}

			pass.Reportf(call.Pos(), "direct os.Exit call is forbidden in main package")

			return true
		})
	}

	return nil, nil
}
