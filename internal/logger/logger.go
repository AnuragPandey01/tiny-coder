package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/anuragpandey01/tiny-coder/internal/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

// InitAppLogger initializes and sets the default application logger with the provided configuration.
func InitAppLogger(config *config.AppConfig, sessionID string) error {

	logFilePath := filepath.Join(config.PathConfig.LogDir, config.LogConfig.LogFileName)

	if err := os.MkdirAll(config.PathConfig.LogDir, 0755); err != nil {
		return fmt.Errorf("create log directory: %w", err)
	}
	rotator := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    config.LogConfig.MaxSizeMB,
		MaxBackups: config.LogConfig.MaxBackups,
		MaxAge:     config.LogConfig.MaxAgeDays,
		Compress:   config.LogConfig.Compress,
	}

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	fileHandler := slog.NewJSONHandler(rotator, opts)

	logger := slog.New(fileHandler).With("session_id", sessionID)

	slog.SetDefault(logger)
	return nil
}
