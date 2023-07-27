package agent

import (
	"context"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics"
)

type MetricAgent interface {
	Send(ctx context.Context, metrics []metrics.Metric) error
}
