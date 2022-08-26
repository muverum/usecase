package log

import (
	"github.com/sirupsen/logrus"
	"log"
)

type UseCaseLogger interface {
	Log(args ...any)
}

type LogrusWrapLogger struct {
	logger *logrus.Logger
}

func (l *LogrusWrapLogger) Log(args ...any) {
	l.logger.Info(args...)
}

func NewWrappedLogrus(l *logrus.Logger) *LogrusWrapLogger {
	return &LogrusWrapLogger{
		logger: l,
	}
}

type LogWrapper struct {
	Logger *log.Logger
}

func (l *LogWrapper) Log(args ...any) {
	l.Logger.Println(args...)
}

func NewLogWrapper(l *log.Logger) *LogWrapper {
	return &LogWrapper{
		Logger: l,
	}
}
