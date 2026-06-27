package main

import "fmt"

func generateResetMethod(structName string, fields []resetField) string {
	result := fmt.Sprintf("func (v *%s) Reset() {\n", structName)
	result += "\tif v == nil {\n"
	result += "\t\treturn\n"
	result += "\t}\n\n"

	for _, field := range fields {
		result += field.ResetLine("v")
	}

	result += "}\n\n"

	return result
}

type resetFieldKind int

const (
	resetFieldScalar resetFieldKind = iota
	resetFieldPointer
	resetFieldSlice
	resetFieldMap
	resetFieldResetter
)

type resetField struct {
	Name       string
	ZeroValue  string
	Kind       resetFieldKind
	ImportPath string
}

func (f resetField) ResetLine(receiver string) string {
	switch f.Kind {
	case resetFieldPointer:
		return fmt.Sprintf("\tif %s.%s != nil {\n\t\t%s.%s = %s\n\t}\n",
			receiver, f.Name, receiver, f.Name, f.ZeroValue)
	case resetFieldSlice:
		return fmt.Sprintf("\t%s.%s = %s.%s[:0]\n",
			receiver, f.Name, receiver, f.Name)
	case resetFieldMap:
		return fmt.Sprintf("\tclear(%s.%s)\n",
			receiver, f.Name)
	case resetFieldResetter:
		return fmt.Sprintf("\tif resetter, ok := any(%s.%s).(interface{ Reset() }); ok &&"+
			" %s.%s != nil {\n\t\tresetter.Reset()\n\t}\n",
			receiver, f.Name, receiver, f.Name)
	default:
		return fmt.Sprintf("\t%s.%s = %s\n", receiver, f.Name, f.ZeroValue)
	}
}
