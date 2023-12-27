package config

import (
	"fmt"
	"log/slog"
)

// LogLevelValue реализует интерфейс flag.Value
type LogLevelValue struct {
	Value *slog.Level
}

func (l *LogLevelValue) String() string {
	if l.Value == nil {
		return ""
	}
	return l.Value.String()
}

func (l *LogLevelValue) Set(value string) error {
	level, found := logLevelMap[value]
	if !found {
		return fmt.Errorf("invalid log level: %s", value)
	}
	*l.Value = level
	return nil
}

var logLevelMap = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}
