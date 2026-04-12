package config

import (
	"encoding"
	"fmt"
	"strconv"
	"time"
)

type CustomSecondsDuration struct {
	time.Duration
	BeenSet bool
}

// a fancy go assertion because env is implicit and it's parsing is also implicit
var _ encoding.TextUnmarshaler = (*CustomSecondsDuration)(nil)

func (d *CustomSecondsDuration) String() string {
	return d.Duration.String()
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (d *CustomSecondsDuration) Set(value string) error {
	if dur, err := time.ParseDuration(value); err == nil {
		d.Duration = dur
		d.BeenSet = true
		return nil
	}

	if value == "false" {
		d.Duration = time.Duration(0)
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

func (d *CustomSecondsDuration) UnmarshalText(text []byte) error {
	return d.Set(string(text))
}
