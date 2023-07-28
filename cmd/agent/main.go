package main

import (
	"context"
	"flag"
	"time"

	"github.com/caarlos0/env/v7"

	"github.com/MlDenis/prometheus_wannabe/internal/agent"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics"
	"github.com/MlDenis/prometheus_wannabe/internal/worker"
)

type config struct {
	MetricReceiverAddress  string
	SendTimeout            time.Duration
	PollInterval           time.Duration
	ReportInterval         time.Duration
	ListOfCollectedMetrics []string
}

func main() {
	cfg, err := createConfig()
	if err != nil {
		panic(err)
	}

	metricAgent, err := agent.NewHTTPMetricsAgent(cfg)
	if err != nil {
		panic(err)
	}

	runtimeMetricsProvider := metrics.NewRuntimeMetricsProvider(cfg)
	customMetricsProvider := metrics.NewCustomMetricsProvider()
	aggregateMetricsProvider := metrics.NewAggregateMetricsProvider([]metrics.MetricsProvider{runtimeMetricsProvider, customMetricsProvider})
	getMetricsWorker := worker.NewHardWorker(aggregateMetricsProvider.Update)
	sendMetricsWorker := worker.NewHardWorker(func(workerContext context.Context) error {
		return metricAgent.Send(workerContext, aggregateMetricsProvider.GetMetrics())
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go getMetricsWorker.StartWork(ctx, cfg.PollInterval)
	sendMetricsWorker.StartWork(ctx, cfg.ReportInterval)
}

func createConfig() (*config, error) {

	cfg := &config{}
	cfg.ListOfCollectedMetrics = []string{
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
	}

	flag.StringVar(&cfg.MetricReceiverAddress, "a", "http://localhost:8080", "Metrics server URL")
	flag.DurationVar(&cfg.SendTimeout, "r", 10*time.Second, "Send metrics interval")
	flag.DurationVar(&cfg.PollInterval, "p", 2*time.Second, "Update metrics interval")
	flag.DurationVar(&cfg.ReportInterval, "t", 10*time.Second, "Push metrics timeout")
	flag.Parse()

	err := env.Parse(cfg)
	return cfg, err
}

func (c *config) MetricsList() []string {
	return c.ListOfCollectedMetrics
}

func (c *config) MetricsReceiverAddress() string {
	return c.MetricReceiverAddress
}

func (c *config) SendMetricsTimeout() time.Duration {
	return c.SendTimeout
}
