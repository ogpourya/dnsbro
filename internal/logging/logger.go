package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// Logger wraps the stdlib logger with optional file output.
type Logger struct {
	*log.Logger
	file  *os.File
	level Level
}

// Level represents the minimum severity that will be emitted.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelSilent
)

// New creates a logger writing to stdout and optionally a file, honoring a minimum level.
func New(path string, level string, prefix string) (*Logger, error) {
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
	return &Logger{
		Logger: l,
		file:   f,
		level:  parseLevel(level),
	}, nil
}

// Printf logs an info-level message if the logger is enabled for it.
func (l *Logger) Printf(format string, v ...interface{}) {
	l.logf(LevelInfo, format, v...)
}

// Debugf logs a debug-level message if enabled.
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.logf(LevelDebug, "DEBUG: "+format, v...)
}

// Infof logs an info-level message.
func (l *Logger) Infof(format string, v ...interface{}) {
	l.logf(LevelInfo, format, v...)
}

// Warnf logs a warning-level message.
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.logf(LevelWarn, "WARN: "+format, v...)
}

// Errorf logs an error-level message.
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.logf(LevelError, "ERROR: "+format, v...)
}

// Close closes any file handle the logger owns.
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

func (l *Logger) logf(level Level, format string, v ...interface{}) {
	if level < l.level || l.level == LevelSilent {
		return
	}
	l.Logger.Printf(format, v...)
}

func parseLevel(level string) Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return LevelDebug
	case "info", "":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	case "silent", "off", "none":
		return LevelSilent
	default:
		fmt.Fprintf(os.Stderr, "unknown log level %q, defaulting to info\n", level)
		return LevelInfo
	}
}
