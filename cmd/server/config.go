package main

import "go.uber.org/zap"

type config struct {
	Key           string          `env:"KEY"`
	ServerURL     string          `env:"ADDRESS"`
	StoreInterval int             `env:"STORE_INTERVAL"`
	StoreFile     string          `env:"STORE_FILE"`
	Restore       bool            `env:"RESTORE"`
	DB            string          `env:"DATABASE_DSN"`
	LogLevel      zap.AtomicLevel `env:"LOG_LEVEL"`
}
