package storage

import (
	"context"

	"github.com/MlDenis/prometheus_wannabe/internal/metrics"
)

type MetricsStorage interface {
	AddMetricValues(ctx context.Context, metric []metrics.Metric) ([]metrics.Metric, error)
	GetMetricValues(ctx context.Context) (map[string]map[string]string, error)
	GetMetric(ctx context.Context, metricType string, metricName string) (metrics.Metric, error)
	Restore(ctx context.Context, metricValues map[string]map[string]string) error
}
