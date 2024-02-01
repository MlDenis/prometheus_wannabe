package config

import (
	"time"

	"github.com/sirupsen/logrus"
)

type Config struct {
	Key                   string       `env:"KEY"`
	ServerURL             string       `env:"ADDRESS"`
	PushRateLimit         int          `env:"RATE_LIMIT"`
	PushTimeout           int          `env:"PUSH_TIMEOUT"`
	SendMetricsInterval   int          `env:"REPORT_INTERVAL"`
	UpdateMetricsInterval int          `env:"POLL_INTERVAL"`
	LogLevel              logrus.Level `env:"LOG_LEVEL"`
	CollectMetricsList    []string
}

func (c *Config) MetricsList() []string {
	return c.CollectMetricsList
}

func (c *Config) MetricsServerURL() string {
	return c.ServerURL
}

func (c *Config) PushMetricsTimeout() time.Duration {
	return time.Duration(c.PushTimeout) * time.Second
}

func (c *Config) ParallelLimit() int {
	return c.PushRateLimit
}

func (c *Config) GetKey() []byte {
	return []byte(c.Key)
}

func (c *Config) SignMetrics() bool {
	return c.Key != ""
}
