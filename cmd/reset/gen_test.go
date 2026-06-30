package main

import "testing"

func TestResetField_ResetLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		field    resetField
		receiver string
		want     string
	}{
		{
			name: "scalar",
			field: resetField{
				Name:      "Name",
				ZeroValue: `""`,
				Kind:      resetFieldScalar,
			},
			receiver: "v",
			want:     "\tv.Name = \"\"\n",
		},
		{
			name: "pointer",
			field: resetField{
				Name: "Client",
				Kind: resetFieldPointer,
			},
			receiver: "v",
			want:     "\tresetPointer(v.Client)\n",
		},
		{
			name: "slice",
			field: resetField{
				Name: "Metrics",
				Kind: resetFieldSlice,
			},
			receiver: "v",
			want:     "\tv.Metrics = v.Metrics[:0]\n",
		},
		{
			name: "map",
			field: resetField{
				Name: "Values",
				Kind: resetFieldMap,
			},
			receiver: "v",
			want:     "\tclear(v.Values)\n",
		},
		{
			name: "resetter",
			field: resetField{
				Name: "Time",
				Kind: resetFieldResetter,
			},
			receiver: "v",
			want: "\tif resetter, ok := any(v.Time).(interface{ Reset() }); ok &&" +
				" v.Time != nil {\n\t\tresetter.Reset()\n\t}\n",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.field.ResetLine(tt.receiver)
			if got != tt.want {
				t.Fatalf("ResetLine() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateResetMethod(t *testing.T) {
	t.Parallel()

	fields := []resetField{
		{Name: "Name", ZeroValue: `""`, Kind: resetFieldScalar},
		{Name: "Metrics", Kind: resetFieldSlice},
		{Name: "Values", Kind: resetFieldMap},
	}

	got := generateResetMethod("Event", fields)
	want := "func (v *Event) Reset() {\n" +
		"\tif v == nil {\n" +
		"\t\treturn\n" +
		"\t}\n\n" +
		"\tv.Name = \"\"\n" +
		"\tv.Metrics = v.Metrics[:0]\n" +
		"\tclear(v.Values)\n" +
		"}\n\n"

	if got != want {
		t.Fatalf("generateResetMethod() = %q, want %q", got, want)
	}
}

func TestGenerateResetPointerHelper(t *testing.T) {
	t.Parallel()

	got := generateResetPointerHelper()
	want := `func resetPointer[T any](v *T) {
	if v == nil {
		return
	}

	if resetter, ok := any(v).(interface{ Reset() }); ok {
		resetter.Reset()
		return
	}

	*v = *new(T)
}

`

	if got != want {
		t.Fatalf("generateResetPointerHelper() = %q, want %q", got, want)
	}
}

func TestHasPointerFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		structs []GenerationMarkedStructInfo
		want    bool
	}{
		{
			name: "empty structs",
		},
		{
			name: "no pointer fields",
			structs: []GenerationMarkedStructInfo{
				{
					Name: "Event",
					Fields: []resetField{
						{Name: "Name", Kind: resetFieldScalar},
						{Name: "Metrics", Kind: resetFieldSlice},
					},
				},
			},
		},
		{
			name: "has pointer field",
			structs: []GenerationMarkedStructInfo{
				{
					Name: "Agent",
					Fields: []resetField{
						{Name: "Client", Kind: resetFieldPointer},
					},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := hasPointerFields(tt.structs)
			if got != tt.want {
				t.Fatalf("hasPointerFields() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateResetMethodsWithoutPointerFields(t *testing.T) {
	t.Parallel()

	structs := []GenerationMarkedStructInfo{
		{
			Name: "Event",
			Fields: []resetField{
				{Name: "Name", ZeroValue: `""`, Kind: resetFieldScalar},
			},
		},
	}

	got := generateResetMethods(structs)
	want := "func (v *Event) Reset() {\n" +
		"\tif v == nil {\n" +
		"\t\treturn\n" +
		"\t}\n\n" +
		"\tv.Name = \"\"\n" +
		"}\n\n"

	if got != want {
		t.Fatalf("generateResetMethods() = %q, want %q", got, want)
	}
}

func TestGenerateResetMethodsWithPointerFields(t *testing.T) {
	t.Parallel()

	structs := []GenerationMarkedStructInfo{
		{
			Name: "Agent",
			Fields: []resetField{
				{Name: "Client", Kind: resetFieldPointer},
				{Name: "ServerAddr", ZeroValue: `""`, Kind: resetFieldScalar},
			},
		},
		{
			Name: "WebSink",
			Fields: []resetField{
				{Name: "AuditURL", ZeroValue: `""`, Kind: resetFieldScalar},
			},
		},
	}

	got := generateResetMethods(structs)
	want := generateResetPointerHelper() +
		"func (v *Agent) Reset() {\n" +
		"\tif v == nil {\n" +
		"\t\treturn\n" +
		"\t}\n\n" +
		"\tresetPointer(v.Client)\n" +
		"\tv.ServerAddr = \"\"\n" +
		"}\n\n" +
		"func (v *WebSink) Reset() {\n" +
		"\tif v == nil {\n" +
		"\t\treturn\n" +
		"\t}\n\n" +
		"\tv.AuditURL = \"\"\n" +
		"}\n\n"

	if got != want {
		t.Fatalf("generateResetMethods() = %q, want %q", got, want)
	}
}
