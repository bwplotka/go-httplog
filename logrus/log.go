package httplogrus

import (
	"github.com/Bplotka/go-httplog"
	"github.com/Sirupsen/logrus"
)

type infoLogger struct {
	logrus.FieldLogger
}

func ToHTTPFieldLoggerInfo(l logrus.FieldLogger) httplog.FieldLogger {
	return &infoLogger{
		FieldLogger: l,
	}
}

func (l *infoLogger) WithFields(fields httplog.Fields) httplog.FieldLogger {
	return ToHTTPFieldLoggerInfo(l.FieldLogger.WithFields(logrus.Fields(fields)))
}

func (l *infoLogger) WithError(err error) httplog.FieldLogger {
	return ToHTTPFieldLoggerInfo(l.FieldLogger.WithError(err))
}

func (l *infoLogger) Log(args ...interface{}) {
	l.Info(args...)
}

type debugLogger struct {
	logrus.FieldLogger
}

func ToHTTPFieldLoggerDebug(l logrus.FieldLogger) httplog.FieldLogger {
	return &debugLogger{
		FieldLogger: l,
	}
}

func (l *debugLogger) WithFields(fields httplog.Fields) httplog.FieldLogger {
	return ToHTTPFieldLoggerDebug(l.FieldLogger.WithFields(logrus.Fields(fields)))
}

func (l *debugLogger) WithError(err error) httplog.FieldLogger {
	return ToHTTPFieldLoggerDebug(l.FieldLogger.WithError(err))
}

func (l *debugLogger) Log(args ...interface{}) {
	l.Debug(args...)
}
