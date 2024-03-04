package logger

import "go.uber.org/zap"

var logger = zap.Must(zap.NewProduction(zap.AddCaller(), zap.AddCallerSkip(1)))

func Sync() error {
	return logger.Sync()
}

func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}
