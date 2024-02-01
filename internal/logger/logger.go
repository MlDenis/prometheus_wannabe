package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
)

const (
	LogDebug = "debug"
	LogInfo  = "info"
)

func InitLogger(debugLevel string) {
	switch debugLevel {
	case LogInfo:
		logrus.SetLevel(logrus.InfoLevel)
	case LogDebug:
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func Error(message string) {
	log.Printf("[ERROR]: %v\r\n", message)
}

func ErrorObj(err error) {
	ErrorFormat("%v", err)
}

func ErrorFormat(format string, v ...any) {
	Error(fmt.Sprintf(format, v...))
}

func WrapError(message string, err error) error {
	wrap := fmt.Errorf("failed to "+message+": %w", err)
	ErrorObj(wrap)

	return wrap
}
