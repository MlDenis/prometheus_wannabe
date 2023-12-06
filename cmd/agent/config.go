package main

import (
	"github.com/sirupsen/logrus"
)

type config struct {
	Key                   string       `env:"KEY"`
	ServerURL             string       `env:"ADDRESS"`
	PushRateLimit         int          `env:"RATE_LIMIT"`
	PushTimeout           int          `env:"PUSH_TIMEOUT"`
	SendMetricsInterval   int          `env:"REPORT_INTERVAL"`
	UpdateMetricsInterval int          `env:"POLL_INTERVAL"`
	LogLevel              logrus.Level `env:"LOG_LEVEL"`
	CollectMetricsList    []string
}
