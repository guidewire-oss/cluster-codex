package config

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

type ClxLogger struct {
	baseLogger zerolog.Logger
}

var Logger = &ClxLogger{}

//func ConfigureLogger(level string) {
//	var logLevel slog.Level
//
//	switch strings.ToLower(level) {
//	case "debug":
//		logLevel = slog.LevelDebug
//	case "info":
//		logLevel = slog.LevelInfo
//	case "warn", "warning":
//		logLevel = slog.LevelWarn
//	case "error":
//		logLevel = slog.LevelError
//	default:
//		logLevel = slog.LevelWarn
//	}
//
//	// Configure slog with the chosen log level
//	ClxLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
//		Level: logLevel,
//	}))
//	slog.SetDefault(ClxLogger)
//}

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

func (l *ClxLogger) Infof(message string, args ...interface{}) {
	l.baseLogger.Info().Msgf(message, args...)
}

func (l *ClxLogger) Error(message string) {
	l.baseLogger.Error().Msgf(message)
}

func (l *ClxLogger) Errorf(message string, args ...interface{}) {
	l.baseLogger.Error().Msgf(message, args...)
}

func (l *ClxLogger) Debugf(message string, args ...interface{}) {
	l.baseLogger.Debug().Msgf(message, args...)
}

func (l *ClxLogger) Warnf(message string, args ...interface{}) {
	l.baseLogger.Warn().Msgf(message, args...)
}

func (l *ClxLogger) Fatalf(message string, args ...interface{}) {
	l.baseLogger.Fatal().Msgf(message, args...)
}
