package config

// OptionalString string value that tracks whether it was explicitly set.
type OptionalString struct {
	Value   string
	BeenSet bool
}

// String returns the string value
func (b *OptionalString) String() string {
	return b.Value
}

// Set stores s and marks the value as explicitly set
func (b *OptionalString) Set(s string) error {
	b.Value = s
	b.BeenSet = true
	return nil
}

// UnmarshalText stores a byte array as a string and marks the value as explicitly set
func (b *OptionalString) UnmarshalText(text []byte) error {
	v := string(text)
	b.Value = v
	b.BeenSet = true
	return nil
}
