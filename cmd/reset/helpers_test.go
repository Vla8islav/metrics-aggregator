package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestFindGoFiles(t *testing.T) {
	root := t.TempDir()

	paths := []string{
		"main.go",
		"internal/service/service.go",
		"internal/service/service_test.go",
		"README.md",
		".git/ignored.go",
		".idea/ignored.go",
		"vendor/ignored.go",
	}

	for _, path := range paths {
		fullPath := filepath.Join(root, path)
		requireNoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
		requireNoError(t, os.WriteFile(fullPath, []byte("package main\n"), 0o644))
	}

	got, err := findGoFiles(root)
	requireNoError(t, err)

	want := []string{
		filepath.Join(root, "main.go"),
		filepath.Join(root, "internal/service/service.go"),
		filepath.Join(root, "internal/service/service_test.go"),
	}

	for i := range want {
		want[i], err = filepath.Abs(want[i])
		requireNoError(t, err)
	}

	sort.Strings(got)
	sort.Strings(want)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("findGoFiles() = %#v, want %#v", got, want)
	}
}

func TestReadImportPaths(t *testing.T) {
	file := parseFile(t, `
package testdata

import (
	"context"
	customjson "encoding/json"
	. "fmt"
	_ "net/http/pprof"
	"time"
)
`)

	got := readImportPaths(file)
	want := map[string]string{
		"context":    "context",
		"customjson": "encoding/json",
		"time":       "time",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("readImportPaths() = %#v, want %#v", got, want)
	}
}

func TestImportPathFor(t *testing.T) {
	imports := map[string]string{
		"time": "time",
		"json": "encoding/json",
	}

	tests := []struct {
		name string
		expr ast.Expr
		want string
	}{
		{
			name: "selector",
			expr: selectorExpr("time", "Time"),
			want: "time",
		},
		{
			name: "pointer to selector",
			expr: &ast.StarExpr{X: selectorExpr("json", "Decoder")},
			want: "encoding/json",
		},
		{
			name: "unknown selector",
			expr: selectorExpr("zap", "Logger"),
			want: "",
		},
		{
			name: "plain ident",
			expr: ast.NewIdent("string"),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := importPathFor(tt.expr, imports)
			if got != tt.want {
				t.Fatalf("importPathFor() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReadResetFields(t *testing.T) {
	file := parseFile(t, `
package testdata

import (
	"time"
	customjson "encoding/json"
)

type Event struct {
	Name string
	Count, Total int64
	Labels []string
	Attrs map[string]string
	Logger *Logger
	StartedAt time.Time
	*customjson.Decoder
}
`)

	imports := readImportPaths(file)
	st := firstStruct(t, file, "Event")

	got := readResetFields(st, imports)
	want := []resetField{
		{Name: "Name", ZeroValue: `""`, Kind: resetFieldScalar},
		{Name: "Count", ZeroValue: "0", Kind: resetFieldScalar},
		{Name: "Total", ZeroValue: "0", Kind: resetFieldScalar},
		{Name: "Labels", ZeroValue: "nil", Kind: resetFieldSlice},
		{Name: "Attrs", ZeroValue: "nil", Kind: resetFieldMap},
		{Name: "Logger", ZeroValue: "nil", Kind: resetFieldResetter},
		{Name: "StartedAt", ZeroValue: "*new(time.Time)", Kind: resetFieldScalar, ImportPath: "time"},
		{Name: "Decoder", ZeroValue: "nil", Kind: resetFieldPointer, ImportPath: "encoding/json"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("readResetFields() = %#v, want %#v", got, want)
	}
}

func TestEmbeddedFieldName(t *testing.T) {
	tests := []struct {
		name string
		expr ast.Expr
		want string
	}{
		{
			name: "ident",
			expr: ast.NewIdent("Logger"),
			want: "Logger",
		},
		{
			name: "selector",
			expr: selectorExpr("zap", "Logger"),
			want: "Logger",
		},
		{
			name: "pointer selector",
			expr: &ast.StarExpr{X: selectorExpr("zap", "Logger")},
			want: "Logger",
		},
		{
			name: "unsupported",
			expr: &ast.ArrayType{Elt: ast.NewIdent("string")},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := embeddedFieldName(tt.expr)
			if got != tt.want {
				t.Fatalf("embeddedFieldName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExprString(t *testing.T) {
	expr := &ast.MapType{
		Key:   ast.NewIdent("string"),
		Value: &ast.ArrayType{Elt: ast.NewIdent("int")},
	}

	got := exprString(expr)
	want := "map[string][]int"

	if got != want {
		t.Fatalf("exprString() = %q, want %q", got, want)
	}
}

func TestResetKindFor(t *testing.T) {
	tests := []struct {
		name string
		expr ast.Expr
		want resetFieldKind
	}{
		{
			name: "pointer to local ident is resetter",
			expr: &ast.StarExpr{X: ast.NewIdent("Logger")},
			want: resetFieldResetter,
		},
		{
			name: "pointer to selector is plain pointer",
			expr: &ast.StarExpr{X: selectorExpr("zap", "Logger")},
			want: resetFieldPointer,
		},
		{
			name: "slice",
			expr: &ast.ArrayType{Elt: ast.NewIdent("string")},
			want: resetFieldSlice,
		},
		{
			name: "map",
			expr: &ast.MapType{Key: ast.NewIdent("string"), Value: ast.NewIdent("int")},
			want: resetFieldMap,
		},
		{
			name: "scalar",
			expr: ast.NewIdent("string"),
			want: resetFieldScalar,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resetKindFor(tt.expr)
			if got != tt.want {
				t.Fatalf("resetKindFor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZeroValueFor(t *testing.T) {
	tests := []struct {
		name string
		expr ast.Expr
		want string
	}{
		{name: "string", expr: ast.NewIdent("string"), want: `""`},
		{name: "bool", expr: ast.NewIdent("bool"), want: "false"},
		{name: "int", expr: ast.NewIdent("int"), want: "0"},
		{name: "byte", expr: ast.NewIdent("byte"), want: "0"},
		{name: "error", expr: ast.NewIdent("error"), want: "nil"},
		{name: "any", expr: ast.NewIdent("any"), want: "nil"},
		{name: "named local type", expr: ast.NewIdent("UnixTime"), want: "*new(UnixTime)"},
		{name: "selector", expr: selectorExpr("time", "Time"), want: "*new(time.Time)"},
		{name: "pointer", expr: &ast.StarExpr{X: ast.NewIdent("Logger")}, want: "nil"},
		{name: "slice", expr: &ast.ArrayType{Elt: ast.NewIdent("string")}, want: "nil"},
		{name: "map", expr: &ast.MapType{Key: ast.NewIdent("string"), Value: ast.NewIdent("int")}, want: "nil"},
		{name: "chan", expr: &ast.ChanType{Value: ast.NewIdent("int")}, want: "nil"},
		{name: "func", expr: &ast.FuncType{}, want: "nil"},
		{name: "interface", expr: &ast.InterfaceType{}, want: "nil"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := zeroValueFor(tt.expr)
			if got != tt.want {
				t.Fatalf("zeroValueFor() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestZeroValueForIdent(t *testing.T) {
	tests := map[string]string{
		"string": "\"\"",
		"bool":   "false",
		"uint64": "0",
		"rune":   "0",
		"error":  "nil",
		"any":    "nil",
		"Money":  "*new(Money)",
	}

	for name, want := range tests {
		t.Run(name, func(t *testing.T) {
			got := zeroValueForIdent(name)
			if got != want {
				t.Fatalf("zeroValueForIdent(%q) = %q, want %q", name, got, want)
			}
		})
	}
}

func TestZeroValueForNamedType(t *testing.T) {
	got := zeroValueForNamedType(selectorExpr("time", "Time"))
	want := "*new(time.Time)"

	if got != want {
		t.Fatalf("zeroValueForNamedType() = %q, want %q", got, want)
	}
}

func TestExprFromName(t *testing.T) {
	got := exprFromName("Metric")
	ident, ok := got.(*ast.Ident)
	if !ok {
		t.Fatalf("exprFromName() returned %T, want *ast.Ident", got)
	}

	if ident.Name != "Metric" {
		t.Fatalf("exprFromName().Name = %q, want %q", ident.Name, "Metric")
	}
}

func parseFile(t *testing.T, source string) *ast.File {
	t.Helper()

	file, err := parser.ParseFile(token.NewFileSet(), "testdata.go", source, parser.ParseComments)
	requireNoError(t, err)

	return file
}

func firstStruct(t *testing.T, file *ast.File, name string) *ast.StructType {
	t.Helper()

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != name {
				continue
			}

			st, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				t.Fatalf("type %s is %T, want *ast.StructType", name, typeSpec.Type)
			}

			return st
		}
	}

	t.Fatalf("struct %s not found", name)
	return nil
}

func selectorExpr(pkgName string, typeName string) ast.Expr {
	return &ast.SelectorExpr{
		X:   ast.NewIdent(pkgName),
		Sel: ast.NewIdent(typeName),
	}
}

func requireNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}
