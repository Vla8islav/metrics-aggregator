package config

import "strconv"

// OptionalBool boolean value that tracks whether it was explicitly set
type OptionalBool struct {
	Value   bool
	BeenSet bool
}

// String returns the boolean value formatted as a string
func (b *OptionalBool) String() string {
	return strconv.FormatBool(b.Value)
}

// Set parses string as a boolean and marks the value as explicitly set
func (b *OptionalBool) Set(s string) error {
	val, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	b.Value = val
	b.BeenSet = true
	return nil
}

// UnmarshalText parses a byte array as a boolean and marks the value as explicitly set
func (b *OptionalBool) UnmarshalText(text []byte) error {
	v, err := strconv.ParseBool(string(text))
	if err != nil {
		return err
	}
	b.Value = v
	b.BeenSet = true
	return nil
}
