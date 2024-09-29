package log

import (
	"io"
	"log/slog"
	"os"
)

// Logger is the global log instance.
// Overwrite it with any slog.New() compatible logger to use your custom logger.
var Logger = slog.Default()

// LoggerFileMode controls the file permission set by the custom file loggers of this package.
var LoggerFileMode os.FileMode = 0600

// CustomLoggerLevel stores the current log level.
// Set it with Level()
var CustomLoggerLevel = &slog.LevelVar{}

// Level sets the global log level and returns the previous value.
func Level(level slog.Level) (oldLevel slog.Level) {
	oldLevel = CustomLoggerLevel.Level()
	CustomLoggerLevel.Set(level)

	if Logger == slog.Default() {
		oldLevel = slog.SetLogLoggerLevel(level)
	}
	return oldLevel
}

// SetDefault will set the global Logger variable to slog.Default()
func SetDefault() {
	Logger = slog.Default()
}

// SetStderrText will set a custom os.Stderr logger with text format.
func SetStderrText() {
	opts := &slog.HandlerOptions{
		Level: CustomLoggerLevel,
	}

	Logger = slog.New(slog.NewTextHandler(os.Stderr, opts))
}

// SetStderrJSON will set a custom os.Stderr logger with JSON format.
func SetStderrJSON() {
	opts := &slog.HandlerOptions{
		Level: CustomLoggerLevel,
	}

	Logger = slog.New(slog.NewJSONHandler(os.Stderr, opts))
}

// SetFileText will set a custom file logger with text format.
func SetFileText(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, LoggerFileMode)
	if err != nil {
		return err
	}

	opts := &slog.HandlerOptions{
		Level: CustomLoggerLevel,
	}

	Logger = slog.New(slog.NewTextHandler(file, opts))
	return nil
}

// SetFileJSON will set a custom file logger with JSON format.
func SetFileJSON(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, LoggerFileMode)
	if err != nil {
		return err
	}

	opts := &slog.HandlerOptions{
		Level: CustomLoggerLevel,
	}

	Logger = slog.New(slog.NewJSONHandler(file, opts))
	return nil
}

// SetMultiText will set a custom logger with text format, that will write into a file and os.Stderr.
func SetMultiText(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, LoggerFileMode)
	if err != nil {
		return err
	}
	writer := io.MultiWriter(file, os.Stderr)

	opts := &slog.HandlerOptions{
		Level: CustomLoggerLevel,
	}

	Logger = slog.New(slog.NewTextHandler(writer, opts))
	return nil
}

// SetMultiJSON will set a custom logger with JSON format, that will write into a file and os.Stderr.
func SetMultiJSON(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, LoggerFileMode)
	if err != nil {
		return err
	}
	writer := io.MultiWriter(file, os.Stderr)

	opts := &slog.HandlerOptions{
		Level: CustomLoggerLevel,
	}

	Logger = slog.New(slog.NewJSONHandler(writer, opts))
	return nil
}

// Error will call [slog.Logger.Error] on the global logger.
func Error(msg string, args ...any) {
	Logger.Error(msg, args...)
}

// Warn will call [slog.Logger.Warn] on the global logger.
func Warn(msg string, args ...any) {
	Logger.Warn(msg, args...)
}

// Info will call [slog.Logger.Info] on the global logger.
func Info(msg string, args ...any) {
	Logger.Info(msg, args...)
}

// Debug will call [slog.Logger.Debug] on the global logger.
func Debug(msg string, args ...any) {
	Logger.Debug(msg, args...)
}
