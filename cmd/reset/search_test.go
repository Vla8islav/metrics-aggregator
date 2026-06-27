package main

import (
	"go/ast"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFindResetStructs(t *testing.T) {
	filePath := writeTempGoFile(t, `
package testdata

import (
	"time"
	customjson "encoding/json"
)

// generate:reset
type Event struct {
	Name string
	Count int64
	StartedAt time.Time
	*customjson.Decoder
}

type Ignored struct {
	Name string
}

/* generate:reset */
type Audit struct {
	Tags []string
	Attrs map[string]string
}
`)

	gotStructs, gotPackage, err := findResetStructs(filePath)
	requireNoError(t, err)

	wantPackage := "testdata"
	if gotPackage != wantPackage {
		t.Fatalf("findResetStructs() package = %q, want %q", gotPackage, wantPackage)
	}

	wantStructs := []GenerationMarkedStructInfo{
		{
			Name: "Event",
			Fields: []resetField{
				{Name: "Name", ZeroValue: `""`, Kind: resetFieldScalar},
				{Name: "Count", ZeroValue: "0", Kind: resetFieldScalar},
				{Name: "StartedAt", ZeroValue: "*new(time.Time)", Kind: resetFieldScalar, ImportPath: "time"},
				{Name: "Decoder", ZeroValue: "nil", Kind: resetFieldPointer, ImportPath: "encoding/json"},
			},
		},
		{
			Name: "Audit",
			Fields: []resetField{
				{Name: "Tags", ZeroValue: "nil", Kind: resetFieldSlice},
				{Name: "Attrs", ZeroValue: "nil", Kind: resetFieldMap},
			},
		},
	}

	if !reflect.DeepEqual(gotStructs, wantStructs) {
		t.Fatalf("findResetStructs() structs = %#v, want %#v", gotStructs, wantStructs)
	}
}

func TestFindResetStructsGroupedTypeDeclaration(t *testing.T) {
	filePath := writeTempGoFile(t, `
package testdata

// generate:reset
type (
	User struct {
		Name string
	}

	Order struct {
		Number string
	}

	Alias string
)
`)

	gotStructs, gotPackage, err := findResetStructs(filePath)
	requireNoError(t, err)

	if gotPackage != "testdata" {
		t.Fatalf("findResetStructs() package = %q, want %q", gotPackage, "testdata")
	}

	wantStructs := []GenerationMarkedStructInfo{
		{
			Name: "User",
			Fields: []resetField{
				{Name: "Name", ZeroValue: `""`, Kind: resetFieldScalar},
			},
		},
		{
			Name: "Order",
			Fields: []resetField{
				{Name: "Number", ZeroValue: `""`, Kind: resetFieldScalar},
			},
		},
	}

	if !reflect.DeepEqual(gotStructs, wantStructs) {
		t.Fatalf("findResetStructs() structs = %#v, want %#v", gotStructs, wantStructs)
	}
}

func TestFindResetStructsIgnoresCommentOnPreviousDeclaration(t *testing.T) {
	filePath := writeTempGoFile(t, `
package testdata

// generate:reset
const answer = 42

type Event struct {
	Name string
}
`)

	gotStructs, gotPackage, err := findResetStructs(filePath)
	requireNoError(t, err)

	if gotPackage != "testdata" {
		t.Fatalf("findResetStructs() package = %q, want %q", gotPackage, "testdata")
	}

	if len(gotStructs) != 0 {
		t.Fatalf("findResetStructs() structs = %#v, want empty slice", gotStructs)
	}
}

func TestFindResetStructsParseError(t *testing.T) {
	filePath := writeTempGoFile(t, `
package testdata

type Broken struct {
	Name string
`)

	gotStructs, gotPackage, err := findResetStructs(filePath)
	if err == nil {
		t.Fatal("findResetStructs() error = nil, want parse error")
	}

	if gotStructs != nil {
		t.Fatalf("findResetStructs() structs = %#v, want nil", gotStructs)
	}

	if gotPackage != "" {
		t.Fatalf("findResetStructs() package = %q, want empty string", gotPackage)
	}
}

func TestHasGenerateResetComment(t *testing.T) {
	tests := []struct {
		name         string
		commentGroup *ast.CommentGroup
		want         bool
	}{
		{
			name:         "nil comment group",
			commentGroup: nil,
			want:         false,
		},
		{
			name: "line comment",
			commentGroup: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "// generate:reset"},
			}},
			want: true,
		},
		{
			name: "block comment",
			commentGroup: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "/* generate:reset */"},
			}},
			want: true,
		},
		{
			name: "block comment with spaces",
			commentGroup: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "/*   generate:reset   */"},
			}},
			want: true,
		},
		{
			name: "different comment",
			commentGroup: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "// generate:other"},
			}},
			want: false,
		},
		{
			name: "same text but extra suffix",
			commentGroup: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "// generate:reset please"},
			}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasGenerateResetComment(tt.commentGroup)
			if got != tt.want {
				t.Fatalf("hasGenerateResetComment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func writeTempGoFile(t *testing.T, source string) string {
	t.Helper()

	root := t.TempDir()
	filePath := filepath.Join(root, "input.go")
	requireNoError(t, os.WriteFile(filePath, []byte(source), 0o644))

	return filePath
}
