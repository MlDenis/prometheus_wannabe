package sendler

import (
	"context"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics"
)

type MetricsPusher interface {
	Push(ctx context.Context, metrics <-chan metrics.Metric) error
}
