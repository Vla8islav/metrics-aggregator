package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type StructInfo struct {
	Name   string
	Fields []resetField
}

// findResetStructs ищет все структуры с комментарием // generate:reset
// перед объявлением типа.
func findResetStructs(filePath string) ([]StructInfo, string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(
		fset,
		filePath,
		nil,
		parser.ParseComments,
	)
	if err != nil {
		return nil, "", err
	}
	var structs []StructInfo
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		// type StructName struct {}
		if genDecl.Tok != token.TYPE {
			continue
		}
		if !hasGenerateResetComment(genDecl.Doc) {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			structs = append(structs, StructInfo{
				Name:   typeSpec.Name.Name,
				Fields: readResetFields(structType),
			})
		}
	}
	return structs, file.Name.Name, nil
}

func hasGenerateResetComment(commentGroup *ast.CommentGroup) bool {
	if commentGroup == nil {
		return false
	}
	for _, comment := range commentGroup.List {
		text := strings.TrimSpace(comment.Text)
		if text == "// generate:reset" {
			return true
		}
		// На случай блочного комментария:
		//
		// /* generate:reset */
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
		text = strings.TrimSpace(text)
		if text == "generate:reset" {
			return true
		}
	}
	return false
}
