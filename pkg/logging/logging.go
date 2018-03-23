package logging

import (
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

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

func (logger *logrusLogger) WithError(err error) Logger {
	entry := logger.target.WithError(err)
	childLogger := &logrusLogger{
		target: logger.target,
		entry:  entry,
	}
	return childLogger
}

func (logger *logrusLogger) WithField(key string, value interface{}) Logger {
	entry := logger.target.WithField(key, value)
	childLogger := &logrusLogger{
		target: logger.target,
		entry:  entry,
	}
	return childLogger
}

func (logger *logrusLogger) WithFields(fields Fields) Logger {
	entry := logger.target.WithFields(logrus.Fields(fields))
	childLogger := &logrusLogger{
		target: logger.target,
		entry:  entry,
	}
	return childLogger
}

func (logger *logrusLogger) Errorf(format string, args ...interface{}) {
	if logger.entry != nil {
		logger.entry.Errorf(format, args...)
	} else {
		logger.target.Errorf(format, args...)
	}
}
func (logger *logrusLogger) Error(args ...interface{}) {
	if logger.entry != nil {
		logger.entry.Error(args...)
	} else {
		logger.target.Error(args...)
	}
}

func (logger *logrusLogger) Warnf(format string, args ...interface{}) {
	logger.target.Warnf(format, args...)
}
func (logger *logrusLogger) Warn(args ...interface{}) {
	logger.target.Warn(args)
}

func (logger *logrusLogger) Infof(format string, args ...interface{}) {
	if logger.entry != nil {
		logger.entry.Infof(format, args...)
	} else {
		logger.target.Infof(format, args...)
	}
}
func (logger *logrusLogger) Info(args ...interface{}) {
	if logger.entry != nil {
		logger.entry.Info(args...)
	} else {
		logger.target.Info(args...)
	}
}

func (logger *logrusLogger) Debugf(format string, args ...interface{}) {
	if logger.entry != nil {
		logger.entry.Debugf(format, args...)
	} else {
		logger.target.Debugf(format, args...)
	}
}
func (logger *logrusLogger) Debug(args ...interface{}) {
	if logger.entry != nil {
		logger.entry.Debug(args...)
	} else {
		logger.target.Debug(args...)
	}
}

func (logger *logrusLogger) Writer() *io.PipeWriter {
	return logger.target.Writer()
}

//NewLogger - Create new logger instance
func NewLogger() Logger {
	logger := logrusLogger{
		target: &logrus.Logger{
			Out: os.Stderr,
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
	target := logrus.Logger{
		Out:       file,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}
	logger := logrusLogger{
		target: &target,
	}
	return &logger
}
