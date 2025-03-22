package logger

import (
	"log/slog"
	"os"
)

func CreateLogger(logLevel *slog.LevelVar) *slog.Logger {
	logOpts := &slog.HandlerOptions{Level: logLevel}
	logHandler := slog.NewJSONHandler(os.Stdout, logOpts)
	logger := slog.New(logHandler)
	slog.SetDefault(logger)
	return logger
}
