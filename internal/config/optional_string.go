package config

import "encoding/json"

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

// UnmarshalJSON stores a JSON string as a string and marks the value as explicitly set
func (v *OptionalString) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	v.Value = value
	v.BeenSet = true
	return nil
}
