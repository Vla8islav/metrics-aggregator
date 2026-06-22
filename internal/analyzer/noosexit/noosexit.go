// Package noosexit custom ast analyzer.
package noosexit

import (
	"go/ast"
	"go/types"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "noosexit checks that os.Exit is not called directly from func main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename

		// Skip generated/non-source files, including Go build cache test mains.
		if !strings.HasSuffix(filename, ".go") {
			continue
		}

		// Skip generated Go test main wrapper if it appears as source.
		if filepath.Base(filename) == "_testmain.go" {
			continue
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			if fn.Name.Name != "main" {
				continue
			}

			if fn.Recv != nil || fn.Body == nil {
				continue
			}

			ast.Inspect(fn.Body, func(n ast.Node) bool {
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

				obj := pass.TypesInfo.Uses[ident]
				pkgName, ok := obj.(*types.PkgName)
				if !ok {
					return true
				}

				if pkgName.Imported().Path() != "os" {
					return true
				}

				pass.Reportf(call.Pos(), "direct os.Exit call is forbidden in func main")

				return true
			})
		}
	}

	return nil, nil
}
