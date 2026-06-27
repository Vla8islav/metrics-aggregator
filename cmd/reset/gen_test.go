package main

import "testing"

func TestGenerateResetMethod(t *testing.T) {
	fields := []resetField{
		{Name: "Name", ZeroValue: `""`, Kind: resetFieldScalar},
		{Name: "Age", ZeroValue: "0", Kind: resetFieldScalar},
		{Name: "Logger", ZeroValue: "nil", Kind: resetFieldPointer},
		{Name: "Tags", ZeroValue: "nil", Kind: resetFieldSlice},
		{Name: "Attrs", ZeroValue: "nil", Kind: resetFieldMap},
		{Name: "State", ZeroValue: "nil", Kind: resetFieldResetter},
	}

	got := generateResetMethod("User", fields)
	want := "func (v *User) Reset() {\n" +
		"\tif v == nil {\n" +
		"\t\treturn\n" +
		"\t}\n\n" +
		"\tv.Name = \"\"\n" +
		"\tv.Age = 0\n" +
		"\tif v.Logger != nil {\n" +
		"\t\tv.Logger = nil\n" +
		"\t}\n" +
		"\tv.Tags = v.Tags[:0]\n" +
		"\tclear(v.Attrs)\n" +
		"\tif resetter, ok := any(v.State).(interface{ Reset() }); ok && v.State != nil {\n" +
		"\t\tresetter.Reset()\n" +
		"\t}\n" +
		"}\n\n"

	if got != want {
		t.Fatalf("generateResetMethod() =\n%s\nwant:\n%s", got, want)
	}
}

func TestGenerateResetMethodWithoutFields(t *testing.T) {
	got := generateResetMethod("Empty", nil)
	want := "func (v *Empty) Reset() {\n" +
		"\tif v == nil {\n" +
		"\t\treturn\n" +
		"\t}\n\n" +
		"}\n\n"

	if got != want {
		t.Fatalf("generateResetMethod() =\n%s\nwant:\n%s", got, want)
	}
}

func TestResetFieldResetLine(t *testing.T) {
	tests := []struct {
		name     string
		field    resetField
		receiver string
		want     string
	}{
		{
			name:     "scalar",
			field:    resetField{Name: "Name", ZeroValue: `""`, Kind: resetFieldScalar},
			receiver: "v",
			want:     "\tv.Name = \"\"\n",
		},
		{
			name:     "pointer",
			field:    resetField{Name: "Logger", ZeroValue: "nil", Kind: resetFieldPointer},
			receiver: "v",
			want: "\tif v.Logger != nil {\n" +
				"\t\tv.Logger = nil\n" +
				"\t}\n",
		},
		{
			name:     "slice",
			field:    resetField{Name: "Tags", ZeroValue: "nil", Kind: resetFieldSlice},
			receiver: "v",
			want:     "\tv.Tags = v.Tags[:0]\n",
		},
		{
			name:     "map",
			field:    resetField{Name: "Attrs", ZeroValue: "nil", Kind: resetFieldMap},
			receiver: "v",
			want:     "\tclear(v.Attrs)\n",
		},
		{
			name:     "resetter",
			field:    resetField{Name: "State", ZeroValue: "nil", Kind: resetFieldResetter},
			receiver: "v",
			want: "\tif resetter, ok := any(v.State).(interface{ Reset() }); ok && v.State != nil {\n" +
				"\t\tresetter.Reset()\n" +
				"\t}\n",
		},
		{
			name:     "custom receiver",
			field:    resetField{Name: "Count", ZeroValue: "0", Kind: resetFieldScalar},
			receiver: "dst",
			want:     "\tdst.Count = 0\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.field.ResetLine(tt.receiver)
			if got != tt.want {
				t.Fatalf("ResetLine() =\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}
