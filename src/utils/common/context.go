package common

import (
	"context"
	"github.com/sirupsen/logrus"
	"os"
)

const (
	RequestIdKey = "requestid"
)

func GetLogger(ctx context.Context) *logrus.Entry {
	if ctx == nil {
		panic("cannot retrieve logger from a nil context")
	}
	logger := createLogger(ctx)
	return logger
}

func GetRequestId(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, ok := ctx.Value(RequestIdKey).(string)
	if !ok {
		return ""
	}
	return id
}

func createLogger(ctx context.Context) *logrus.Entry {
	log := &logrus.Logger{
		Out: os.Stderr,
		Formatter: &logrus.TextFormatter{
			DisableQuote: true,
		},
		Hooks:        make(logrus.LevelHooks),
		Level:        logrus.DebugLevel,
		ReportCaller: true,
	}
	requestId := GetRequestId(ctx)
	if requestId == "" {
		return log.WithContext(ctx)
	}
	return log.WithContext(ctx).WithField(RequestIdKey, GetRequestId(ctx))
}
