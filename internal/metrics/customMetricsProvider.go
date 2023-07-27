package metrics

import (
	"context"
	"fmt"
	"math/rand"
)

type customMetricsProvider struct {
	poolMetric   Metric
	randomMetric Metric
}

func NewCustomMetricsProvider() MetricsProvider {
	return &customMetricsProvider{
		poolMetric:   NewCounterMetric("PollCount"),
		randomMetric: NewGaugeMetric("RandomValue"),
	}
}

func (c *customMetricsProvider) GetMetrics() []Metric {
	return []Metric{
		c.poolMetric,
		c.randomMetric,
	}
}

func (c *customMetricsProvider) Update(context.Context) error {
	fmt.Println("Start collect custom metrics")

	c.poolMetric.SetValue(1)
	fmt.Println("Updated metric: %v. value: %v", c.poolMetric.GetName(), c.poolMetric.GetStringValue())

	c.randomMetric.SetValue(rand.Float64())
	fmt.Println("Updated metric: %v. value: %v", c.randomMetric.GetName(), c.randomMetric.GetStringValue())

	return nil
}
