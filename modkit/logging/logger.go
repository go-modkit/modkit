// Package logging provides a minimal, structured logging interface for modkit applications.
package logging

// Logger is the core logging interface used throughout modkit.
// It provides structured logging with key-value pairs and log levels.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
}
