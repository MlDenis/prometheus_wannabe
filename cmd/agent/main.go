package main

import (
	"context"
	"time"

	"github.com/MlDenis/prometheus_wannabe/internal/agent"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics"
	"github.com/MlDenis/prometheus_wannabe/internal/worker"
)

const (
	metricsReceiverAddress = "http://localhost:8080"
	sendTimeout            = 10 * time.Second
	pollInterval           = 2 * time.Second
	reportInterval         = 10 * time.Second
)

func main() {
	conf := createConfig()
	metricAgent := agent.NewHTTPMetricsAgent(conf)
	runtimeMetricsProvider := metrics.NewRuntimeMetricsProvider(conf)
	customMetricsProvider := metrics.NewCustomMetricsProvider()
	aggregateMetricsProvider := metrics.NewAggregateMetricsProvider([]metrics.MetricsProvider{runtimeMetricsProvider, customMetricsProvider})
	getMetricsWorker := worker.NewHardWorker(aggregateMetricsProvider.Update)
	sendMetricsWorker := worker.NewHardWorker(func(workerContext context.Context) error {
		return metricAgent.Send(workerContext, aggregateMetricsProvider.GetMetrics())
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go getMetricsWorker.StartWork(ctx, pollInterval)
	sendMetricsWorker.StartWork(ctx, reportInterval)
}

type config struct {
	metricReceiverAddress  string
	sendTimeout            time.Duration
	pollInterval           time.Duration
	reportInterval         time.Duration
	listOfCollectedMetrics []string
}

func (c *config) MetricsList() []string {
	return c.listOfCollectedMetrics
}

func (c *config) MetricsReceiverAddress() string {
	return c.metricReceiverAddress
}

func (c *config) SendMetricsTimeout() time.Duration {
	return c.sendTimeout
}

func createConfig() *config {
	return &config{
		metricReceiverAddress: metricsReceiverAddress,
		sendTimeout:           sendTimeout,
		pollInterval:          pollInterval,
		reportInterval:        reportInterval,
		listOfCollectedMetrics: []string{
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
		},
	}
}
