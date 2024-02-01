package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	LogDebug = "debug"
	LogInfo  = "info"
)

// Global logger
var log *zap.Logger
var SugarLogger *zap.SugaredLogger

// Initializing the logger with a given debug level
func InitLogger(debugLevel string) {
	var level zapcore.Level
	switch debugLevel {
	case LogInfo:
		level = zapcore.InfoLevel
	case LogDebug:
		level = zapcore.DebugLevel
	default:
		level = zapcore.InfoLevel
	}

	config := zap.NewProductionConfig()
	config.Level.SetLevel(level)
	var err error
	log, err = config.Build()
	if err != nil {
		panic(err)
	}
	SugarLogger = log.Sugar()
}

// Error logging
func Error(message string) {
	if log != nil {
		log.Error(message)
	}
}

func ErrorObj(err error) {
	if err != nil {
		Error(err.Error())
	}
}

func ErrorFormat(format string, v ...interface{}) {
	Error(fmt.Sprintf(format, v...))
}

func WrapError(message string, err error) error {
	wrap := fmt.Errorf("failed to "+message+": %w", err)
	ErrorObj(wrap)

	return wrap
}
