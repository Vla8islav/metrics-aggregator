package main

import (
	"fmt"
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

		fmt.Println("file:", file)
		fmt.Println("package:", packageName)
		for _, s := range structs {
			fmt.Println("struct:", s.Name)
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
