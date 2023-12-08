package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/MlDenis/prometheus_wannabe/internal/config"
	"github.com/MlDenis/prometheus_wannabe/internal/hash"
	"github.com/MlDenis/prometheus_wannabe/internal/logger"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics/model"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics/provider/agregate"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics/provider/custom"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics/provider/gopsutil"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics/provider/runtime"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics/sendler/http"
	"github.com/MlDenis/prometheus_wannabe/internal/worker"
	"github.com/caarlos0/env/v7"
)

func main() {

	conf, err := createConfig()
	if err != nil {
		panic(logger.WrapError("initialize config", err))
	}

	logger.InitLogger(fmt.Sprint(conf.LogLevel))

	signer := hash.NewSigner(conf)
	converter := model.NewMetricsConverter(conf, signer)
	metricPusher, err := http.NewMetricsPusher(conf, converter)
	if err != nil {
		panic(logger.WrapError("create new metrics pusher", err))
	}

	runtimeMetricsProvider := runtime.NewRuntimeMetricsProvider(conf)
	customMetricsProvider := custom.NewCustomMetricsProvider()
	gopsutilMetricsProvider := gopsutil.NewGopsutilMetricsProvider()
	aggregateMetricsProvider := agregate.NewAggregateMetricsProvider(
		runtimeMetricsProvider,
		customMetricsProvider,
		gopsutilMetricsProvider,
	)
	getMetricsWorker := worker.NewHardWorker(aggregateMetricsProvider.Update)
	pushMetricsWorker := worker.NewHardWorker(func(workerContext context.Context) error {
		return metricPusher.Push(workerContext, aggregateMetricsProvider.GetMetrics())
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go getMetricsWorker.StartWork(ctx, conf.UpdateMetricsInterval)
	pushMetricsWorker.StartWork(ctx, conf.SendMetricsInterval)
}

func createConfig() (*config.Config, error) {
	conf := &config.Config{CollectMetricsList: []string{
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
	}}

	flag.StringVar(&conf.Key, "k", "", "Signer secret key")
	flag.StringVar(&conf.ServerURL, "a", "localhost:8080", "Metrics server URL")
	flag.IntVar(&conf.PushRateLimit, "l", 20, "Push metrics parallel workers limit")
	flag.IntVar(&conf.PushTimeout, "t", 10, "Push metrics timeout")
	flag.IntVar(&conf.SendMetricsInterval, "r", 10, "Send metrics interval")
	flag.IntVar(&conf.UpdateMetricsInterval, "p", 2, "Update metrics interval")
	flag.Parse()

	err := env.Parse(conf)
	return conf, err
}
