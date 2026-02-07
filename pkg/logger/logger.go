package logger

import (
	"os"

	"github.com/rs/zerolog"
)

type Logger struct {
	*zerolog.Logger
}

func NewLogger(level, env string) *Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Set log level
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(logLevel)

	var log zerolog.Logger

	if env == "production" {
		// JSON format for production
		log = zerolog.New(os.Stdout).
			With().
			Timestamp().
			Logger()
	} else {
		// Pretty console output for development
		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"}
		log = zerolog.New(output).
			With().
			Timestamp().
			Logger()
	}

	return &Logger{&log}
}
