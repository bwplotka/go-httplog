package httplogrus

import (
	"github.com/Bplotka/go-httplog"
	"github.com/Sirupsen/logrus"
)

type infoLogger struct {
	logrus.FieldLogger
}

// ToHTTPFieldLoggerInfo returns httplog from logrus logger with claim to log with info level.
func ToHTTPFieldLoggerInfo(l logrus.FieldLogger) httplog.FieldLogger {
	return &infoLogger{
		FieldLogger: l,
	}
}

// WithFields adds new fields to structured logger.
func (l *infoLogger) WithFields(fields httplog.Fields) httplog.FieldLogger {
	return ToHTTPFieldLoggerInfo(l.FieldLogger.WithFields(logrus.Fields(fields)))
}

// Logs write log line with logrus info level.
func (l *infoLogger) Log(args ...interface{}) {
	l.Info(args...)
}

type debugLogger struct {
	logrus.FieldLogger
}

// ToHTTPFieldLoggerDebug returns httplog from logrus logger with claim to log with info debug.
func ToHTTPFieldLoggerDebug(l logrus.FieldLogger) httplog.FieldLogger {
	return &debugLogger{
		FieldLogger: l,
	}
}

// WithFields adds new fields to structured logger.
func (l *debugLogger) WithFields(fields httplog.Fields) httplog.FieldLogger {
	return ToHTTPFieldLoggerDebug(l.FieldLogger.WithFields(logrus.Fields(fields)))
}

// Logs write log line with logrus debug level.
func (l *debugLogger) Log(args ...interface{}) {
	l.Debug(args...)
}
