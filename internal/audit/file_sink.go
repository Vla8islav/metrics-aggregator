package audit

import (
	"context"
	"encoding/json"
	"os"
)

type FileSink struct {
	path string
}

func NewFileSink(path string) *FileSink {
	return &FileSink{path: path}
}

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
	_, err = f.Write(payload)
	if err != nil {
		return err
	}
	return nil
}
