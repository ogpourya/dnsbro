package logging

import (
	"io"
	"log"
	"os"
)

// Logger wraps the stdlib logger with optional file output.
type Logger struct {
	*log.Logger
	file *os.File
}

// New creates a logger writing to stdout and optionally a file.
func New(path string, prefix string) (*Logger, error) {
	writers := []io.Writer{os.Stdout}
	var f *os.File

	if path != "" {
		var err error
		f, err = os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, err
		}
		writers = append(writers, f)
	}

	l := log.New(io.MultiWriter(writers...), prefix, log.LstdFlags)
	return &Logger{Logger: l, file: f}, nil
}

// Close closes any file handle the logger owns.
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
