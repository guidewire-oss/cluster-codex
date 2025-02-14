package config

import (
	"log/slog"
	"os"
)

var ClxLogger *slog.Logger

func init() {
	// Configure the logger as desired. Here we use a text handler writing to stdout.
	ClxLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

var GlobalConfig Config

type Config struct {
}
