package audit

import (
	"context"
	"encoding/json"
	"os"
)

// FileSink writes audit events to a file as JSON lines
type FileSink struct {
	path string
}

// NewFileSink creates a FileSink that writes audit events to path
func NewFileSink(path string) *FileSink {
	return &FileSink{path: path}
}

// Write appends e to the sink file as a JSON line
func (s *FileSink) Write(_ context.Context, e Event) error {
	payload, err := json.Marshal(e)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(payload, '\n'))
	if err != nil {
		return err
	}
	return nil
}
