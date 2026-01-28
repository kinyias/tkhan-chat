package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

// Init initializes the logger
func Init(mode string) {
	var err error
	var config zap.Config

	if mode == "release" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	log, err = config.Build()
	if err != nil {
		panic(err)
	}
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	log.Info(msg, fields...)
}

// Error logs an error message
func Error(msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	log.Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	log.Fatal(msg, fields...)
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	log.Debug(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	log.Warn(msg, fields...)
}

// Sync flushes any buffered log entries
func Sync() {
	_ = log.Sync()
}
