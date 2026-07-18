package config

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// OptionalInt integer value that tracks whether it was explicitly set
type OptionalInt struct {
	Value   int
	BeenSet bool
}

// String returns the integer value formatted as a string
func (b *OptionalInt) String() string {
	return strconv.Itoa(b.Value)
}

// Set parses s as an integer and marks the value as explicitly set
func (b *OptionalInt) Set(s string) error {
	val, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	b.Value = val
	b.BeenSet = true
	return nil
}

// UnmarshalText parses text as an integer and marks the value as explicitly set
func (b *OptionalInt) UnmarshalText(text []byte) error {
	v := string(text)
	result, err := strconv.Atoi(v)
	if err != nil {
		return fmt.Errorf("cannot convert string to int: %w", err)
	}
	b.Value = result
	b.BeenSet = true
	return nil
}

// UnmarshalJSON stores a JSON int as an int and marks the value as explicitly set
func (b *OptionalInt) UnmarshalJSON(data []byte) error {
	var value int
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	b.Value = value
	b.BeenSet = true
	return nil
}
