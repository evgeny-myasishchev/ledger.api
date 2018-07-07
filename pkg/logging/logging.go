package logging

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type contextKey string

const loggerKey contextKey = "logger"

// Fields - represents fields structure
type Fields map[string]interface{}

// Logger - logger interface
type Logger interface {
	Errorf(format string, args ...interface{})
	Error(args ...interface{})

	Warnf(format string, args ...interface{})
	Warn(args ...interface{})

	Infof(format string, args ...interface{})
	Info(args ...interface{})

	Debugf(format string, args ...interface{})
	Debug(args ...interface{})

	WithError(err error) Logger
	WithField(key string, value interface{}) Logger
	WithFields(fields Fields) Logger
}

type logrusLogger struct {
	target *logrus.Logger
	entry  *logrus.Entry
}

type logrusTarget interface {
	WithError(err error) *logrus.Entry
	WithField(key string, value interface{}) *logrus.Entry
	WithFields(fields logrus.Fields) *logrus.Entry

	Info(args ...interface{})
	Infof(format string, args ...interface{})

	Debug(args ...interface{})
	Debugf(format string, args ...interface{})

	Warn(args ...interface{})
	Warnf(format string, args ...interface{})

	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

func (logger *logrusLogger) getTarget() logrusTarget {
	if logger.entry != nil {
		return logger.entry
	}
	return logger.target
}

func (logger *logrusLogger) WithError(err error) Logger {
	entry := logger.getTarget().WithError(err)
	childLogger := &logrusLogger{
		target: logger.target,
		entry:  entry,
	}
	return childLogger
}

func (logger *logrusLogger) WithField(key string, value interface{}) Logger {
	entry := logger.getTarget().WithField(key, value)
	childLogger := &logrusLogger{
		target: logger.target,
		entry:  entry,
	}
	return childLogger
}

func (logger *logrusLogger) WithFields(fields Fields) Logger {
	entry := logger.getTarget().WithFields(logrus.Fields(fields))
	childLogger := &logrusLogger{
		target: logger.target,
		entry:  entry,
	}
	return childLogger
}

func (logger *logrusLogger) Errorf(format string, args ...interface{}) {
	logger.getTarget().Errorf(format, args...)
}
func (logger *logrusLogger) Error(args ...interface{}) {
	logger.getTarget().Error(args...)
}

func (logger *logrusLogger) Warnf(format string, args ...interface{}) {
	logger.getTarget().Warnf(format, args...)
}
func (logger *logrusLogger) Warn(args ...interface{}) {
	logger.getTarget().Warn(args)
}

func (logger *logrusLogger) Infof(format string, args ...interface{}) {
	logger.getTarget().Infof(format, args...)
}
func (logger *logrusLogger) Info(args ...interface{}) {
	logger.getTarget().Info(args...)
}

func (logger *logrusLogger) Debugf(format string, args ...interface{}) {
	logger.getTarget().Debugf(format, args...)
}
func (logger *logrusLogger) Debug(args ...interface{}) {
	logger.getTarget().Debug(args...)
}

func (logger *logrusLogger) Writer() *io.PipeWriter {
	return logger.target.Writer()
}

//NewLogger - Create new logger instance
func NewLogger(env string) Logger {
	if env == "test" {
		return NewTestLogger()
	}

	if env == "dev" {
		return NewPrettyLogger(os.Stdout)
	}

	logger := logrusLogger{
		target: &logrus.Logger{
			Out:       os.Stdout,
			Formatter: new(logrus.JSONFormatter),
			Level:     logrus.DebugLevel, //For now using debug, to be changed to info
		},
	}
	return &logger
}

// NewTestLogger - Creates a new instance of a logger for tests
func NewTestLogger() Logger {
	path, err := filepath.Abs("../../test.log")
	if err != nil {
		panic(err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}

	return NewPrettyLogger(file)
}

// NewPrettyLogger - Creates a new instance of a logger that will pretty write
func NewPrettyLogger(out io.Writer) Logger {
	target := logrus.Logger{
		Out:       out,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}
	logger := logrusLogger{
		target: &target,
	}
	return &logger
}

// CreateContext - create a new context with logger value set
func CreateContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext - return logger from given context
func FromContext(ctx context.Context) Logger {
	return ctx.Value(loggerKey).(Logger)
}
