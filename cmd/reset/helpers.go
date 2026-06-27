package main

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
)

func findGoFiles(root string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			switch d.Name() {
			case ".git", ".idea", "vendor":
				return filepath.SkipDir
			}

			return nil
		}

		if filepath.Ext(path) != ".go" {
			return nil
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		files = append(files, absPath)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

func importPathFor(expr ast.Expr, imports map[string]string) string {
	switch expr := expr.(type) {
	case *ast.SelectorExpr:
		if x, ok := expr.X.(*ast.Ident); ok {
			return imports[x.Name]
		}
	case *ast.StarExpr:
		return importPathFor(expr.X, imports)
	}

	return ""
}

func readImportPaths(file *ast.File) map[string]string {
	imports := make(map[string]string)

	for _, importSpec := range file.Imports {
		importPath, err := strconv.Unquote(importSpec.Path.Value)
		if err != nil {
			continue
		}

		if importSpec.Name != nil {
			if importSpec.Name.Name == "." || importSpec.Name.Name == "_" {
				continue
			}

			imports[importSpec.Name.Name] = importPath
			continue
		}

		lastSlash := strings.LastIndex(importPath, "/")
		if lastSlash == -1 {
			imports[importPath] = importPath
			continue
		}

		imports[importPath[lastSlash+1:]] = importPath
	}

	return imports
}

func readResetFields(st *ast.StructType, imports map[string]string) []resetField {
	var fields []resetField

	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			fields = append(fields, resetField{
				Name:       embeddedFieldName(field.Type),
				ZeroValue:  zeroValueFor(field.Type),
				Kind:       resetKindFor(field.Type),
				ImportPath: importPathFor(field.Type, imports),
			})
			continue
		}

		for _, name := range field.Names {
			fields = append(fields, resetField{
				Name:       name.Name,
				ZeroValue:  zeroValueFor(field.Type),
				Kind:       resetKindFor(field.Type),
				ImportPath: importPathFor(field.Type, imports),
			})
		}
	}

	return fields
}

func embeddedFieldName(expr ast.Expr) string {
	switch expr := expr.(type) {
	case *ast.Ident:
		return expr.Name
	case *ast.SelectorExpr:
		return expr.Sel.Name
	case *ast.StarExpr:
		return embeddedFieldName(expr.X)
	default:
		return ""
	}
}

func exprString(expr ast.Expr) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, token.NewFileSet(), expr); err != nil {
		return ""
	}

	return buf.String()
}

func resetKindFor(expr ast.Expr) resetFieldKind {
	switch expr := expr.(type) {
	case *ast.StarExpr:
		if _, ok := expr.X.(*ast.Ident); ok {
			return resetFieldResetter
		}

		return resetFieldPointer
	case *ast.ArrayType:
		return resetFieldSlice
	case *ast.MapType:
		return resetFieldMap
	default:
		return resetFieldScalar
	}
}

func zeroValueFor(expr ast.Expr) string {
	switch expr := expr.(type) {
	case *ast.Ident:
		return zeroValueForIdent(expr.Name)

	case *ast.SelectorExpr:
		return zeroValueForNamedType(expr)

	case *ast.StarExpr, *ast.ArrayType, *ast.MapType, *ast.ChanType, *ast.FuncType, *ast.InterfaceType:
		return "nil"

	default:
		return "nil"
	}
}

func zeroValueForIdent(name string) string {
	switch name {
	case "string":
		return `""`
	case "bool":
		return "false"
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"float32", "float64", "complex64", "complex128",
		"byte", "rune":
		return "0"
	case "error", "any":
		return "nil"
	default:
		return zeroValueForNamedType(exprFromName(name))
	}
}

func zeroValueForNamedType(expr ast.Expr) string {
	return "*new(" + exprString(expr) + ")"
}

func exprFromName(name string) ast.Expr {
	return ast.NewIdent(name)
}
