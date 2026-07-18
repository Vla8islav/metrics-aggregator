package config

import (
	"encoding"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// OptionalSecondsDuration duration flag value that accepts plain seconds or Go duration strings
type OptionalSecondsDuration struct {
	time.Duration
	BeenSet bool
}

// a fancy go assertion because env is implicit and it's parsing is also implicit
var _ encoding.TextUnmarshaler = (*OptionalSecondsDuration)(nil)

// String returns the duration formatted as a Go duration string.
func (d *OptionalSecondsDuration) String() string {
	return d.Duration.String()
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Set parses value as a Go duration string or as a number of seconds in bare numeric
func (d *OptionalSecondsDuration) Set(value string) error {
	if dur, err := time.ParseDuration(value); err == nil {
		d.Duration = dur
		d.BeenSet = true
		return nil
	}

	// if parsing fails, trying to parse it as an integer
	seconds, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid duration: %s", value)
	}
	// assume that this value is in seconds
	d.Duration = time.Duration(absInt(seconds)) * time.Second
	d.BeenSet = true
	return nil
}

// UnmarshalText parses text as a Go duration string or as a number of seconds
func (d *OptionalSecondsDuration) UnmarshalText(text []byte) error {
	return d.Set(string(text))
}

func (d *OptionalSecondsDuration) UnmarshalJSON(data []byte) error {
	var seconds int
	if err := json.Unmarshal(data, &seconds); err == nil {
		d.Duration = time.Duration(absInt(seconds)) * time.Second
		d.BeenSet = true
		return nil
	}

	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	value = strings.TrimSpace(value)
	if value == "" {
		d.Duration = 0
		d.BeenSet = true
		return nil
	}

	return d.Set(value)
}
