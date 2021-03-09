package logger

import (
	"fmt"

	l "github.com/micro/go-micro/v2/logger"
)

type Level int8

const (
	// TraceLevel level. Designates finer-grained informational events than the Debug.
	TraceLevel Level = iota - 2
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
	// InfoLevel is the default logging priority.
	// General operational entries about what's going on inside the application.
	InfoLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	ErrorLevel
	// FatalLevel level. Logs and then calls `logger.Exit(1)`. highest level of severity.
	FatalLevel
)

func (l Level) String() string {
	switch l {
	case TraceLevel:
		return "trace"
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	}
	return ""
}

// Enabled returns true if the given level is at or above this level.
func (l Level) Enabled(lvl Level) bool {
	return lvl >= l
}

// GetLevel converts a level string into a logger Level value.
// returns an error if the input string does not match known values.
func GetLevel(levelStr string) (Level, error) {
	switch levelStr {
	case TraceLevel.String():
		return TraceLevel, nil
	case DebugLevel.String():
		return DebugLevel, nil
	case InfoLevel.String():
		return InfoLevel, nil
	case WarnLevel.String():
		return WarnLevel, nil
	case ErrorLevel.String():
		return ErrorLevel, nil
	case FatalLevel.String():
		return FatalLevel, nil
	}
	return InfoLevel, fmt.Errorf("Unknown Level String: '%s', defaulting to InfoLevel", levelStr)
}

func Info(args ...interface{}) {
	l.Info(args)
}

func Infof(template string, args ...interface{}) {
	l.Infof(template, args)
}

func Trace(args ...interface{}) {
	l.Trace(args)
}

func Tracef(template string, args ...interface{}) {
	l.Tracef(template, args)
}

func Debug(args ...interface{}) {
	l.Debug(args)
}

func Debugf(template string, args ...interface{}) {
	l.Debugf(template, args)
}

func Warn(args ...interface{}) {
	l.Warn(args)
}

func Warnf(template string, args ...interface{}) {
	l.Warnf(template, args)
}

func Error(args ...interface{}) {
	l.Error(args)
}

func Errorf(template string, args ...interface{}) {
	l.Errorf(template, args)
}

func Fatal(args ...interface{}) {
	l.Fatal(args)
}

func Fatalf(template string, args ...interface{}) {
	l.Fatal(template, args)
}

// Returns true if the given level is at or lower the current logger level
func V(lvl Level, log l.Logger) bool {
	return l.V(l.Level(lvl), log)
}
