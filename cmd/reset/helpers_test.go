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
	t.Parallel()

	root := t.TempDir()

	writeFile(t, filepath.Join(root, "main.go"), "package main\n")
	writeFile(t, filepath.Join(root, "README.md"), "# readme\n")
	writeFile(t, filepath.Join(root, "internal", "service.go"), "package internal\n")
	writeFile(t, filepath.Join(root, ".git", "ignored.go"), "package ignored\n")
	writeFile(t, filepath.Join(root, ".idea", "ignored.go"), "package ignored\n")
	writeFile(t, filepath.Join(root, "vendor", "ignored.go"), "package ignored\n")

	got, err := findGoFiles(root)
	if err != nil {
		t.Fatalf("findGoFiles() error = %v", err)
	}

	sort.Strings(got)

	want := []string{
		absPath(t, filepath.Join(root, "internal", "service.go")),
		absPath(t, filepath.Join(root, "main.go")),
	}
	sort.Strings(want)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("findGoFiles() = %#v, want %#v", got, want)
	}
}

func TestImportPathFor(t *testing.T) {
	t.Parallel()

	imports := map[string]string{
		"time":    "time",
		"helpers": "github.com/example/project/internal/helpers",
	}

	tests := []struct {
		name string
		expr ast.Expr
		want string
	}{
		{
			name: "selector expression",
			expr: selectorExpr("time", "Duration"),
			want: "time",
		},
		{
			name: "pointer to selector expression",
			expr: &ast.StarExpr{X: selectorExpr("helpers", "HTTPRetryClient")},
			want: "github.com/example/project/internal/helpers",
		},
		{
			name: "local identifier",
			expr: ast.NewIdent("LocalType"),
		},
		{
			name: "unknown selector alias",
			expr: selectorExpr("unknown", "Type"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := importPathFor(tt.expr, imports)
			if got != tt.want {
				t.Fatalf("importPathFor() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReadImportPaths(t *testing.T) {
	t.Parallel()

	file := parseFile(t, `package sample

import (
	"time"
	alias "github.com/example/project/pkg/value"
	"github.com/example/project/internal/helpers"
	. "github.com/example/project/pkg/dot"
	_ "github.com/example/project/pkg/blank"
)
`)

	got := readImportPaths(file)
	want := map[string]string{
		"time":    "time",
		"alias":   "github.com/example/project/pkg/value",
		"helpers": "github.com/example/project/internal/helpers",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("readImportPaths() = %#v, want %#v", got, want)
	}
}

func TestReadResetFields(t *testing.T) {
	t.Parallel()

	file := parseFile(t, `package sample

import (
	"time"
	"github.com/example/project/internal/helpers"
)

type LocalType struct{}

type Event struct {
	Name string
	Count int
	StartedAt time.Time
	Client *helpers.HTTPRetryClient
	Tags []string
	Values map[string]int
	LocalType
	*helpers.Embedded
}
`)

	imports := readImportPaths(file)
	structType := findStructType(t, file, "Event")

	got := readResetFields(structType, imports)
	want := []resetField{
		{Name: "Name", ZeroValue: `""`, Kind: resetFieldScalar},
		{Name: "Count", ZeroValue: "0", Kind: resetFieldScalar},
		{Name: "StartedAt", ZeroValue: "*new(time.Time)", Kind: resetFieldScalar, ImportPath: "time"},
		{Name: "Client", ZeroValue: "nil", ElemZeroValue: "*new(helpers.HTTPRetryClient)", Kind: resetFieldPointer, ImportPath: "github.com/example/project/internal/helpers"},
		{Name: "Tags", ZeroValue: "nil", Kind: resetFieldSlice},
		{Name: "Values", ZeroValue: "nil", Kind: resetFieldMap},
		{Name: "LocalType", ZeroValue: "*new(LocalType)", Kind: resetFieldScalar},
		{Name: "Embedded", ZeroValue: "nil", ElemZeroValue: "*new(helpers.Embedded)", Kind: resetFieldPointer, ImportPath: "github.com/example/project/internal/helpers"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("readResetFields() = %#v, want %#v", got, want)
	}
}

func TestEmbeddedFieldName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		expr ast.Expr
		want string
	}{
		{
			name: "identifier",
			expr: ast.NewIdent("LocalType"),
			want: "LocalType",
		},
		{
			name: "selector",
			expr: selectorExpr("time", "Time"),
			want: "Time",
		},
		{
			name: "pointer",
			expr: &ast.StarExpr{X: selectorExpr("helpers", "Client")},
			want: "Client",
		},
		{
			name: "unsupported",
			expr: &ast.ArrayType{Elt: ast.NewIdent("string")},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := embeddedFieldName(tt.expr)
			if got != tt.want {
				t.Fatalf("embeddedFieldName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExprString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		expr ast.Expr
		want string
	}{
		{
			name: "identifier",
			expr: ast.NewIdent("LocalType"),
			want: "LocalType",
		},
		{
			name: "selector",
			expr: selectorExpr("time", "Duration"),
			want: "time.Duration",
		},
		{
			name: "array",
			expr: &ast.ArrayType{Len: &ast.BasicLit{Kind: token.INT, Value: "3"}, Elt: ast.NewIdent("string")},
			want: "[3]string",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := exprString(tt.expr)
			if got != tt.want {
				t.Fatalf("exprString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResetKindFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		expr ast.Expr
		want resetFieldKind
	}{
		{
			name: "pointer",
			expr: &ast.StarExpr{X: ast.NewIdent("LocalType")},
			want: resetFieldPointer,
		},
		{
			name: "slice",
			expr: &ast.ArrayType{Elt: ast.NewIdent("string")},
			want: resetFieldSlice,
		},
		{
			name: "array",
			expr: &ast.ArrayType{Len: &ast.BasicLit{Kind: token.INT, Value: "3"}, Elt: ast.NewIdent("string")},
			want: resetFieldScalar,
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := resetKindFor(tt.expr)
			if got != tt.want {
				t.Fatalf("resetKindFor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZeroValueFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		expr ast.Expr
		want string
	}{
		{
			name: "string",
			expr: ast.NewIdent("string"),
			want: `""`,
		},
		{
			name: "bool",
			expr: ast.NewIdent("bool"),
			want: "false",
		},
		{
			name: "int",
			expr: ast.NewIdent("int"),
			want: "0",
		},
		{
			name: "error",
			expr: ast.NewIdent("error"),
			want: "nil",
		},
		{
			name: "named local type",
			expr: ast.NewIdent("UnixTime"),
			want: "*new(UnixTime)",
		},
		{
			name: "selector",
			expr: selectorExpr("time", "Time"),
			want: "*new(time.Time)",
		},
		{
			name: "pointer",
			expr: &ast.StarExpr{X: ast.NewIdent("UnixTime")},
			want: "nil",
		},
		{
			name: "slice",
			expr: &ast.ArrayType{Elt: ast.NewIdent("string")},
			want: "nil",
		},
		{
			name: "array",
			expr: &ast.ArrayType{Len: &ast.BasicLit{Kind: token.INT, Value: "3"}, Elt: ast.NewIdent("string")},
			want: "*new([3]string)",
		},
		{
			name: "map",
			expr: &ast.MapType{Key: ast.NewIdent("string"), Value: ast.NewIdent("int")},
			want: "nil",
		},
		{
			name: "channel",
			expr: &ast.ChanType{Value: ast.NewIdent("int")},
			want: "nil",
		},
		{
			name: "function",
			expr: &ast.FuncType{Params: &ast.FieldList{}},
			want: "nil",
		},
		{
			name: "interface",
			expr: &ast.InterfaceType{Methods: &ast.FieldList{}},
			want: "nil",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := zeroValueFor(tt.expr)
			if got != tt.want {
				t.Fatalf("zeroValueFor() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestElemZeroValueFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		expr ast.Expr
		want string
	}{
		{
			name: "non pointer",
			expr: ast.NewIdent("int"),
		},
		{
			name: "pointer to int",
			expr: &ast.StarExpr{X: ast.NewIdent("int")},
			want: "0",
		},
		{
			name: "pointer to string",
			expr: &ast.StarExpr{X: ast.NewIdent("string")},
			want: `""`,
		},
		{
			name: "pointer to named local type",
			expr: &ast.StarExpr{X: ast.NewIdent("UnixTime")},
			want: "*new(UnixTime)",
		},
		{
			name: "pointer to selector",
			expr: &ast.StarExpr{X: selectorExpr("time", "Time")},
			want: "*new(time.Time)",
		},
		{
			name: "double pointer",
			expr: &ast.StarExpr{X: &ast.StarExpr{X: ast.NewIdent("UnixTime")}},
			want: "nil",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := elemZeroValueFor(tt.expr)
			if got != tt.want {
				t.Fatalf("elemZeroValueFor() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestZeroValueForIdent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "string", in: "string", want: `""`},
		{name: "bool", in: "bool", want: "false"},
		{name: "int64", in: "int64", want: "0"},
		{name: "uint", in: "uint", want: "0"},
		{name: "float64", in: "float64", want: "0"},
		{name: "complex128", in: "complex128", want: "0"},
		{name: "byte", in: "byte", want: "0"},
		{name: "rune", in: "rune", want: "0"},
		{name: "error", in: "error", want: "nil"},
		{name: "any", in: "any", want: "nil"},
		{name: "custom", in: "Custom", want: "*new(Custom)"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := zeroValueForIdent(tt.in)
			if got != tt.want {
				t.Fatalf("zeroValueForIdent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestZeroValueForNamedType(t *testing.T) {
	t.Parallel()

	got := zeroValueForNamedType(selectorExpr("time", "Time"))
	want := "*new(time.Time)"

	if got != want {
		t.Fatalf("zeroValueForNamedType() = %q, want %q", got, want)
	}
}

func TestExprFromName(t *testing.T) {
	t.Parallel()

	got := exprFromName("Custom")
	ident, ok := got.(*ast.Ident)
	if !ok {
		t.Fatalf("exprFromName() returned %T, want *ast.Ident", got)
	}
	if ident.Name != "Custom" {
		t.Fatalf("exprFromName().Name = %q, want %q", ident.Name, "Custom")
	}
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func absPath(t *testing.T, path string) string {
	t.Helper()

	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("Abs() error = %v", err)
	}

	return abs
}

func parseFile(t *testing.T, src string) *ast.File {
	t.Helper()

	file, err := parser.ParseFile(token.NewFileSet(), "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	return file
}

func findStructType(t *testing.T, file *ast.File, name string) *ast.StructType {
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

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				t.Fatalf("type %s is %T, want *ast.StructType", name, typeSpec.Type)
			}

			return structType
		}
	}

	t.Fatalf("struct %s not found", name)
	return nil
}

func selectorExpr(pkg string, name string) ast.Expr {
	return &ast.SelectorExpr{
		X:   ast.NewIdent(pkg),
		Sel: ast.NewIdent(name),
	}
}
