package platform

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type LogCloser interface {
	Close() error
}

type nopCloser struct{}

func (nopCloser) Close() error {
	return nil
}

type multiCloser struct {
	closers []io.Closer
}

func (c multiCloser) Close() error {
	var firstErr error
	for _, closer := range c.closers {
		if err := closer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func NewLogger(cfg Config, service string) (*slog.Logger, LogCloser, error) {
	level := parseLogLevel(cfg.LogLevel)

	writers := []io.Writer{os.Stdout}
	closers := []io.Closer{}

	if strings.TrimSpace(cfg.LogDir) != "" && strings.TrimSpace(cfg.LogFileName) != "" {
		if err := os.MkdirAll(cfg.LogDir, 0o755); err != nil {
			return nil, nopCloser{}, err
		}

		file, err := os.OpenFile(filepath.Join(cfg.LogDir, cfg.LogFileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, nopCloser{}, err
		}
		writers = append(writers, file)
		closers = append(closers, file)
	}

	handlerOptions := &slog.HandlerOptions{Level: level}
	writer := io.MultiWriter(writers...)

	var handler slog.Handler
	switch strings.ToLower(strings.TrimSpace(cfg.LogFormat)) {
	case "", "json":
		handler = slog.NewJSONHandler(writer, handlerOptions)
	default:
		handler = slog.NewTextHandler(writer, handlerOptions)
	}

	return slog.New(handler).With("service", service), multiCloser{closers: closers}, nil
}

func parseLogLevel(value string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
