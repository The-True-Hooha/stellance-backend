package logger

import (
	"log/slog"
	"os"
)

func Logger() *slog.Logger {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}
