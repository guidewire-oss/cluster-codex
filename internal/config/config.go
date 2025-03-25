package config

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func ConfigureLogger(level string) {
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr, PartsExclude: []string{"time"}}

	log.Logger = zerolog.New(consoleWriter).With().Logger()
	switch level {
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "warn", "warning":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}
}
