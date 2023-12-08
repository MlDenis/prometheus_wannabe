package custom

import (
	"context"
	"math/rand"

	"github.com/MlDenis/prometheus_wannabe/internal/metrics"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics/types"

	"github.com/sirupsen/logrus"
)

type customMetricsProvider struct {
	poolMetric   metrics.Metric
	randomMetric metrics.Metric
}

func NewCustomMetricsProvider() metrics.MetricsProvider {
	return &customMetricsProvider{
		poolMetric:   types.NewCounterMetric("PollCount"),
		randomMetric: types.NewGaugeMetric("RandomValue"),
	}
}

func (c *customMetricsProvider) GetMetrics() <-chan metrics.Metric {
	result := make(chan metrics.Metric)
	go func() {
		defer close(result)
		result <- c.poolMetric
		result <- c.randomMetric
	}()

	return result
}

func (c *customMetricsProvider) Update(context.Context) error {
	logrus.Info("Start collect custom metrics")

	c.poolMetric.SetValue(1)
	logrus.Infof("Updated metric: %v. value: %v", c.poolMetric.GetName(), c.poolMetric.GetStringValue())

	c.randomMetric.SetValue(rand.Float64())
	logrus.Infof("Updated metric: %v. value: %v", c.randomMetric.GetName(), c.randomMetric.GetStringValue())

	return nil
}
