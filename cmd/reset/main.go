package main

import (
	"fmt"
	"go/ast"
	"io/fs"
	"path/filepath"
)

func main() {
	files, err := findGoFiles(".")
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		structs, packageName, err := findResetStructs(file)
		if err != nil {
			panic(err)
		}

		for _, s := range structs {
			fmt.Println("file:", file)
			fmt.Println("package:", packageName)
			fmt.Println("struct:", s.Name)
			fmt.Print(generateResetMethod(s.Name, s.Fields))
		}
	}
}

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

		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

func readResetFields(st *ast.StructType) []resetField {
	var fields []resetField

	for _, field := range st.Fields.List {
		for _, name := range field.Names {
			fields = append(fields, resetField{
				Name:      name.Name,
				ZeroValue: zeroValueFor(field.Type),
				Kind:      resetKindFor(field.Type),
			})
		}
	}

	return fields
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
		switch expr.Name {
		case "string":
			return `""`
		case "bool":
			return "false"
		default:
			return "0"
		}
	case *ast.StarExpr:
		return zeroValueFor(expr.X)
	default:
		return "nil"
	}
}
