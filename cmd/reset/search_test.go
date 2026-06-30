package main

import (
	"go/ast"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFindResetStructs(t *testing.T) {
	t.Parallel()

	filePath := writeTempGoFile(t, `package audit

import (
	customjson "encoding/json"
	"time"
)

type IgnoredNoComment struct {
	Name string
}

// generate:reset
type Event struct {
	Name string
	Count int
	StartedAt time.Time
	Decoder *customjson.Decoder
}

// generate:reset
type Audit struct {
	Tags []string
	Attrs map[string]string
}
`)

	gotStructs, gotPackageName, err := findResetStructs(filePath)
	requireNoError(t, err)

	wantStructs := []GenerationMarkedStructInfo{
		{
			Name: "Event",
			Fields: []resetField{
				{Name: "Name", ZeroValue: `""`, Kind: resetFieldScalar},
				{Name: "Count", ZeroValue: "0", Kind: resetFieldScalar},
				{Name: "StartedAt", ZeroValue: "*new(time.Time)", Kind: resetFieldScalar, ImportPath: "time"},
				{Name: "Decoder", ZeroValue: "nil", ElemZeroValue: "*new(customjson.Decoder)", Kind: resetFieldPointer, ImportPath: "encoding/json"},
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

	if gotPackageName != "audit" {
		t.Fatalf("findResetStructs() package = %q, want %q", gotPackageName, "audit")
	}

	if !reflect.DeepEqual(gotStructs, wantStructs) {
		t.Fatalf("findResetStructs() structs = %#v, want %#v", gotStructs, wantStructs)
	}
}

func TestFindResetStructsWithBlockComment(t *testing.T) {
	t.Parallel()

	filePath := writeTempGoFile(t, `package audit

/* generate:reset */
type Event struct {
	Name string
}
`)

	gotStructs, gotPackageName, err := findResetStructs(filePath)
	requireNoError(t, err)

	wantStructs := []GenerationMarkedStructInfo{
		{
			Name: "Event",
			Fields: []resetField{
				{Name: "Name", ZeroValue: `""`, Kind: resetFieldScalar},
			},
		},
	}

	if gotPackageName != "audit" {
		t.Fatalf("findResetStructs() package = %q, want %q", gotPackageName, "audit")
	}

	if !reflect.DeepEqual(gotStructs, wantStructs) {
		t.Fatalf("findResetStructs() structs = %#v, want %#v", gotStructs, wantStructs)
	}
}

func TestFindResetStructsIgnoresNonStructTypes(t *testing.T) {
	t.Parallel()

	filePath := writeTempGoFile(t, `package audit

// generate:reset
type Name string

// generate:reset
type Event struct {
	Name string
}
`)

	gotStructs, _, err := findResetStructs(filePath)
	requireNoError(t, err)

	wantStructs := []GenerationMarkedStructInfo{
		{
			Name: "Event",
			Fields: []resetField{
				{Name: "Name", ZeroValue: `""`, Kind: resetFieldScalar},
			},
		},
	}

	if !reflect.DeepEqual(gotStructs, wantStructs) {
		t.Fatalf("findResetStructs() structs = %#v, want %#v", gotStructs, wantStructs)
	}
}

func TestFindResetStructsReturnsEmptyForUnmarkedStructs(t *testing.T) {
	t.Parallel()

	filePath := writeTempGoFile(t, `package audit

type Event struct {
	Name string
}
`)

	gotStructs, gotPackageName, err := findResetStructs(filePath)
	requireNoError(t, err)

	if gotPackageName != "audit" {
		t.Fatalf("findResetStructs() package = %q, want %q", gotPackageName, "audit")
	}

	if len(gotStructs) != 0 {
		t.Fatalf("findResetStructs() structs = %#v, want empty slice", gotStructs)
	}
}

func TestFindResetStructsReturnsParseError(t *testing.T) {
	t.Parallel()

	filePath := writeTempGoFile(t, `package audit

// generate:reset
type Event struct {
	Name string
`)

	gotStructs, gotPackageName, err := findResetStructs(filePath)
	if err == nil {
		t.Fatal("findResetStructs() error = nil, want error")
	}
	if gotStructs != nil {
		t.Fatalf("findResetStructs() structs = %#v, want nil", gotStructs)
	}
	if gotPackageName != "" {
		t.Fatalf("findResetStructs() package = %q, want empty string", gotPackageName)
	}
}

func TestHasGenerateResetComment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		commentGroup *ast.CommentGroup
		want         bool
	}{
		{
			name: "nil comment group",
		},
		{
			name: "line comment",
			commentGroup: &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "// generate:reset"},
				},
			},
			want: true,
		},
		{
			name: "block comment",
			commentGroup: &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "/* generate:reset */"},
				},
			},
			want: true,
		},
		{
			name: "wrong comment",
			commentGroup: &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "// generate:something-else"},
				},
			},
		},
		{
			name: "similar but not exact line comment",
			commentGroup: &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "//generate:reset"},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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

func requireNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
