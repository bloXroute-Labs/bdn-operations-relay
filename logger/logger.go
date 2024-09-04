package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/log"
)

// InitLogger initializes the logger with the given log level.
func InitLogger(level string) {
	log.SetDefault(log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stdout, parseLevel(strings.ToLower(level)), true)))
}

func Debug(format string, v ...interface{}) {
	log.Debug(format, v...)
}

func Info(format string, v ...interface{}) {
	log.Info(format, v...)
}

func Warn(format string, v ...interface{}) {
	log.Warn(format, v...)
}

func Error(format string, v ...interface{}) {
	log.Error(format, v...)
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return log.LevelDebug
	case "info":
		return log.LevelInfo
	case "warn":
		return log.LevelWarn
	case "error":
		return log.LevelError
	default:
		return log.LevelInfo
	}
}
