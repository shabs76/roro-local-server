package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Setup configures the application logging.
// It sets up Gin's default writer and the standard library's structured logger (slog)
// to write to both stdout and a rotating log file.
func Setup(logDir, filename string) {
	// Ensure the log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		slog.Error("Failed to create log directory", "error", err)
		os.Exit(1)
	}

	logPath := filepath.Join(logDir, filename)

	// Configure lumberjack for log rotation
	rotator := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    10,   // megabytes
		MaxBackups: 5,    // number of backups
		MaxAge:     28,   // days
		Compress:   true, // compress rolled files
	}

	// Create a MultiWriter to write to both the file and stdout
	// This ensures logs are visible in 'docker logs' and saved to the persisted file
	mw := io.MultiWriter(os.Stdout, rotator)

	// Set Gin's output to the multi-writer
	gin.DefaultWriter = mw

	// Configure the default slog logger to use JSON format (good for parsing)
	// and write to our multi-writer
	handler := slog.NewJSONHandler(mw, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Info("Logger initialized", "path", logPath)
}
