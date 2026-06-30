package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	files, err := findGoFiles(".")
	if err != nil {
		panic(err)
	}

	packageToName := make(map[string]string)
	packageToStructs := make(map[string][]GenerationMarkedStructInfo)
	packageToImports := make(map[string]map[string]struct{})

	for _, file := range files {
		packageDir := filepath.Dir(file)
		structs, packageName, err := findResetStructs(file)
		if err != nil {
			panic(err)
		}
		for _, s := range structs {
			fmt.Println("file:", file)
			fmt.Println("package:", packageName)
			fmt.Println("struct:", s.Name)
			fmt.Print(generateResetMethod(s.Name, s.Fields))
			packageToName[packageDir] = packageName
			packageToStructs[packageDir] = append(packageToStructs[packageDir], s)
			if _, found := packageToImports[packageDir]; !found {
				packageToImports[packageDir] = make(map[string]struct{})
			}
			for _, field := range s.Fields {
				if field.ImportPath != "" {
					packageToImports[packageDir][field.ImportPath] = struct{}{}
				}
			}
		}
	}

	for packageDir, structs := range packageToStructs {
		var content strings.Builder
		content.WriteString("package ")
		content.WriteString(packageToName[packageDir])
		content.WriteString("\n\n")
		imports := packageToImports[packageDir]
		if len(imports) > 0 {
			var importPaths []string
			for importPath := range imports {
				importPaths = append(importPaths, importPath)
			}
			sort.Strings(importPaths)
			var importBlock strings.Builder
			importBlock.WriteString("import (\n")
			for _, importPath := range importPaths {
				importBlock.WriteString("\t\"")
				importBlock.WriteString(importPath)
				importBlock.WriteString("\"\n")
			}
			importBlock.WriteString(")\n\n")
			content.WriteString(importBlock.String())
		}
		content.WriteString(generateResetMethods(structs))
		outputPath := filepath.Join(packageDir, "reset.gen.go")
		if err := os.WriteFile(outputPath, []byte(content.String()), 0o644); err != nil {
			panic(err)
		}
	}
}
