package config

import (
	"log/slog"
	"os"
)

func LoadLogger(cfg *Config) *slog.Logger {
	var logMode slog.Level
	if cfg.Mode == MODE_PRODUCTION {
		logMode = slog.LevelInfo
	} else {
		logMode = slog.LevelDebug
	}

	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logMode,
	}))
}
