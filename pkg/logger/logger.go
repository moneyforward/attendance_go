package logger

import (
	"log/slog"
	"os"
)

func InitalizeLogger() *slog.Logger {
	lv := slog.LevelInfo
	level := os.Getenv("LOG_LEVEL")

	if level != "" {
		err := lv.UnmarshalText([]byte(level))
		if err != nil {
			lv = slog.LevelInfo
		}
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: lv,
	}))

	return logger
}
