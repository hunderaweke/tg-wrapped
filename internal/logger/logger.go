package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

var Log *slog.Logger
var logFile *os.File

func init() {
	Log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

// Init initializes the logger with the specified level, format, and optional file logging
func Init(level slog.Level, jsonFormat bool) {
	initWithWriter(level, jsonFormat, os.Stdout)
}

// InitWithFile initializes the logger to write to both stdout and a file
func InitWithFile(level slog.Level, jsonFormat bool, logDir string) error {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02")
	logPath := filepath.Join(logDir, "tg-wrapped-"+timestamp+".log")

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	// Close previous log file if exists
	if logFile != nil {
		logFile.Close()
	}
	logFile = file

	// Create multi-writer for both stdout and file
	multiWriter := io.MultiWriter(os.Stdout, file)
	initWithWriter(level, jsonFormat, multiWriter)

	return nil
}

func initWithWriter(level slog.Level, jsonFormat bool, w io.Writer) {
	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: level}

	if jsonFormat {
		handler = slog.NewJSONHandler(w, opts)
	} else {
		handler = slog.NewTextHandler(w, opts)
	}

	Log = slog.New(handler)
}

// Close closes the log file if one is open
func Close() {
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}
}

func Debug(msg string, args ...any) {
	Log.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	Log.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	Log.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	Log.Error(msg, args...)
}

func With(args ...any) *slog.Logger {
	return Log.With(args...)
}
