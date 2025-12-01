package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/config"
	"gorm.io/gorm/logger"
)

var (
	// Logger is the global application logger
	Logger *slog.Logger
	// logFile holds the file handle for cleanup
	logFile *os.File
)

// Initialize sets up the logger based on configuration
func Initialize(cfg *config.LoggingConfig) error {
	level := parseLevel(cfg.Level)

	var writers []io.Writer

	// Configure outputs
	switch cfg.Output {
	case "stdout":
		writers = append(writers, os.Stdout)
	case "file":
		f, err := openLogFile(cfg.Path)
		if err != nil {
			return err
		}
		logFile = f
		writers = append(writers, f)
	case "both":
		writers = append(writers, os.Stdout)
		f, err := openLogFile(cfg.Path)
		if err != nil {
			return err
		}
		logFile = f
		writers = append(writers, f)
	default:
		writers = append(writers, os.Stdout)
	}

	// Create multi-writer
	var writer io.Writer
	if len(writers) == 1 {
		writer = writers[0]
	} else {
		writer = io.MultiWriter(writers...)
	}

	// Create handler based on format
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: level,
	}

	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(writer, opts)
	} else {
		handler = slog.NewTextHandler(writer, opts)
	}

	Logger = slog.New(handler)
	slog.SetDefault(Logger)

	return nil
}

// openLogFile opens or creates a log file in the specified directory
func openLogFile(path string) (*os.File, error) {
	// Use current directory if path is empty
	if path == "" {
		path = "."
	}

	// Ensure directory exists
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file with date
	filename := fmt.Sprintf("openvpn-mng-%s.log", time.Now().Format("2006-01-02"))
	filepath := filepath.Join(path, filename)

	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return f, nil
}

// parseLevel converts string level to slog.Level
func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Close closes the log file if open
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

// Info logs an info message
func Info(msg string, args ...any) {
	Logger.Info(msg, args...)
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	Logger.Debug(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	Logger.Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	Logger.Error(msg, args...)
}

// GetGormLogLevel returns GORM logger level based on config
func GetGormLogLevel(cfg *config.LoggingConfig) logger.LogLevel {
	switch cfg.Level {
	case "debug":
		return logger.Info // GORM Info shows SQL queries
	case "info":
		return logger.Warn // Only warnings and errors
	case "warn":
		return logger.Warn
	case "error":
		return logger.Error
	default:
		return logger.Warn
	}
}

// GormLogger wraps slog for GORM compatibility
type GormLogger struct {
	SlowThreshold time.Duration
	LogLevel      logger.LogLevel
}

// NewGormLogger creates a new GORM logger
func NewGormLogger(cfg *config.LoggingConfig) *GormLogger {
	return &GormLogger{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      GetGormLogLevel(cfg),
	}
}

// LogMode implements logger.Interface
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info implements logger.Interface
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		Logger.Info(fmt.Sprintf(msg, data...))
	}
}

// Warn implements logger.Interface
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		Logger.Warn(fmt.Sprintf(msg, data...))
	}
}

// Error implements logger.Interface
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		Logger.Error(fmt.Sprintf(msg, data...))
	}
}

// Trace implements logger.Interface
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && l.LogLevel >= logger.Error:
		Logger.Error("GORM query error",
			"error", err,
			"elapsed", elapsed,
			"rows", rows,
			"sql", sql,
		)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		Logger.Warn("GORM slow query",
			"elapsed", elapsed,
			"threshold", l.SlowThreshold,
			"rows", rows,
			"sql", sql,
		)
	case l.LogLevel >= logger.Info:
		Logger.Debug("GORM query",
			"elapsed", elapsed,
			"rows", rows,
			"sql", sql,
		)
	}
}
